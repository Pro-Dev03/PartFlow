import { getArabicErrorMessage, isRetryableError } from '../../lib/error-messages';
import { appConfig } from '../../lib/config/app';
import { TokenManager } from '../../lib/token-manager';

const API_BASE_URL = appConfig.apiUrl;

interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: Record<string, any>;
  error?: {
    code: string;
    message: string;
  };
}

const MAX_RETRIES = 3;
const RETRY_DELAY = 1000; // 1 second
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

class ApiClient {
  private baseURL: string;
  private token: string | null = null;
  private cache: Map<string, { data: any; timestamp: number }> = new Map();

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    // Use TokenManager for consistent token retrieval
    this.token = TokenManager.getToken();
  }

  setToken(token: string) {
    this.token = token;
    // Use TokenManager for consistent token storage
    TokenManager.setToken(token);
  }

  clearToken() {
    this.token = null;
    // Use TokenManager for consistent token clearing
    TokenManager.clearToken();
  }

  logout() {
    this.clearToken();
    this.clearCache();
  }

  private getCacheKey(endpoint: string, options: RequestInit): string {
    return `${endpoint}:${JSON.stringify(options)}`;
  }

  private getFromCache<T>(key: string): ApiResponse<T> | null {
    const cached = this.cache.get(key);
    if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
      return cached.data;
    }
    return null;
  }

  private setCache<T>(key: string, data: ApiResponse<T>): void {
    this.cache.set(key, { data, timestamp: Date.now() });
  }

  private clearCache(): void {
    this.cache.clear();
  }

  private clearCachePattern(pattern: string): void {
    for (const key of this.cache.keys()) {
      if (key.includes(pattern)) {
        this.cache.delete(key);
      }
    }
  }

  private async sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  private async requestWithRetry<T>(
    endpoint: string,
    options: RequestInit = {},
    retryCount: number = 0
  ): Promise<ApiResponse<T>> {
    try {
      return await this.request<T>(endpoint, options);
    } catch (error: any) {
      // Don't retry abort errors or timeout errors
      if (error.name === 'AbortError' || error.message?.includes('timeout')) {
        throw error;
      }

      // Don't retry 400 Bad Request (client errors) - these won't succeed on retry
      if (error?.status === 400) {
        throw error;
      }

      // Don't retry if error message indicates resource not found (already deleted)
      if (error?.message === 'category not found' || error?.message === 'not found') {
        throw error;
      }

      // Don't retry DELETE requests that fail with 400 (likely already deleted)
      if (options.method === 'DELETE' && error?.status === 400) {
        throw error;
      }

      // Check if error is retryable
      if (isRetryableError(error) && retryCount < MAX_RETRIES) {
        console.log(`Retrying request (${retryCount + 1}/${MAX_RETRIES})...`);
        await this.sleep(RETRY_DELAY * (retryCount + 1)); // Exponential backoff
        return this.requestWithRetry<T>(endpoint, options, retryCount + 1);
      }
      throw error;
    }
  }

  /**
   * يحلل استجابة HTTP بأمان. يقرأ النص أولاً ثم يحاول JSON.parse،
   * وإلا يعيد خطأً يحمل نص الاستجابة الخام (مثل "404 page not found"
   * التي تُرجعها gin افتراضياً) بدل إطلاق SyntaxError.
   */
  private async parseResponse<T>(response: Response): Promise<ApiResponse<T>> {
    const text = await response.text();
    if (!text) {
      return { success: response.ok, data: undefined as any, error: response.ok ? undefined : { code: String(response.status), message: response.statusText } };
    }
    try {
      return JSON.parse(text) as ApiResponse<T>;
    } catch {
      return {
        success: false,
        data: undefined as any,
        error: { code: String(response.status), message: text },
      };
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    // Create abort controller for timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 30000); // 30 second timeout

    try {
      const response = await fetch(url, {
        ...options,
        headers,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      const data: ApiResponse<T> = await this.parseResponse<T>(response);

      if (!response.ok) {
        // Handle 401 Unauthorized - try to refresh token
        if (response.status === 401 && this.token) {
          try {
            // Try to refresh token using auth API directly
            const refreshResponse = await fetch(`${this.baseURL}/auth/refresh`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
            });

            if (refreshResponse.ok) {
              const refreshData = await refreshResponse.json();
              const newToken = refreshData.data?.token;
              
              if (newToken) {
                this.setToken(newToken);
                headers['Authorization'] = `Bearer ${newToken}`;
                
                // Retry request with new token
                const retryResponse = await fetch(url, {
                  ...options,
                  headers,
                });
                const retryData: ApiResponse<T> = await this.parseResponse<T>(retryResponse);
                
                if (!retryResponse.ok) {
                  const error: any = new Error(retryData.error?.message || 'An error occurred');
                  error.status = retryResponse.status;
                  error.code = retryData.error?.code;
                  error.response = retryData;
                  error.arabicMessage = getArabicErrorMessage(error);
                  throw error;
                }
                
                return retryData;
              }
            }
          } catch (refreshError) {
            console.error('Token refresh failed:', refreshError);
          }
          
          // If refresh failed, clear token and redirect
          this.clearToken();
          window.location.href = '/login';
          throw new Error('Session expired. Please login again.');
        }
        
        // Create error object with status
        const error: any = new Error(data.error?.message || 'An error occurred');
        error.status = response.status;
        error.code = data.error?.code;
        error.response = data;
        
        // Add Arabic message
        error.arabicMessage = getArabicErrorMessage(error);
        
        throw error;
      }

      return data;
    } catch (error: any) {
      clearTimeout(timeoutId);

      console.error('API request failed:', error);

      // Handle abort errors (timeout)
      if (error.name === 'AbortError') {
        throw new Error('Request timeout - server took too long to respond');
      }

      // Enhance error with Arabic message
      if (error && !error.arabicMessage) {
        error.arabicMessage = getArabicErrorMessage(error);
      }

      throw error;
    }
  }

  async get<T = any>(endpoint: string, params?: any, useCache: boolean = true): Promise<ApiResponse<T>> {
    // Build URL with query parameters
    let url = endpoint;
    if (params && Object.keys(params).length > 0) {
      const queryParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          queryParams.append(key, String(value));
        }
      });
      const queryString = queryParams.toString();
      if (queryString) {
        url = `${endpoint}?${queryString}`;
      }
    }
    
    const cacheKey = this.getCacheKey(url, { method: 'GET' });
    
    if (useCache) {
      const cached = this.getFromCache<T>(cacheKey);
      if (cached) {
        return cached;
      }
    }
    
    const result = await this.requestWithRetry<T>(url, { method: 'GET' });
    
    if (useCache) {
      this.setCache(cacheKey, result);
    }
    
    return result;
  }

  async post<T = any>(endpoint: string, body: any): Promise<ApiResponse<T>> {
    this.clearCachePattern(endpoint.split('/')[1]); // Clear cache for related endpoints
    return this.requestWithRetry<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  }

  async put<T = any>(endpoint: string, body: any): Promise<ApiResponse<T>> {
    this.clearCachePattern(endpoint.split('/')[1]); // Clear cache for related endpoints
    return this.requestWithRetry<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(body),
    });
  }

  async patch<T = any>(endpoint: string, body: any): Promise<ApiResponse<T>> {
    this.clearCachePattern(endpoint.split('/')[1]); // Clear cache for related endpoints
    return this.requestWithRetry<T>(endpoint, {
      method: 'PATCH',
      body: JSON.stringify(body),
    });
  }

  async delete<T = any>(endpoint: string, params?: any): Promise<ApiResponse<T>> {
    // Build URL with query parameters
    let url = endpoint;
    if (params && Object.keys(params).length > 0) {
      const queryParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          queryParams.append(key, String(value));
        }
      });
      const queryString = queryParams.toString();
      if (queryString) {
        url = `${endpoint}?${queryString}`;
      }
    }
    
    this.clearCachePattern(endpoint.split('/')[1]); // Clear cache for related endpoints
    return this.requestWithRetry<T>(url, { method: 'DELETE' });
  }
}

export const apiClient = new ApiClient(API_BASE_URL);