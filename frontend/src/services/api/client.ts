import { getArabicErrorMessage, isRetryableError } from '../../lib/error-messages';
import { appConfig, getActiveApiUrl, getCloudApiUrl } from '../../lib/config/app';
import { TokenManager } from '../../lib/token-manager';

const API_BASE_URL = appConfig.apiUrl;

interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: Record<string, unknown>;
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
  private cache: Map<string, { data: unknown; timestamp: number }> = new Map();
  private refreshInFlight: Promise<string | null> | null = null;

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

  private getBaseURL(): string {
    const nextBaseUrl = getActiveApiUrl() || this.baseURL;
    this.baseURL = nextBaseUrl;
    return nextBaseUrl;
  }

  private getCacheKey(endpoint: string, options: RequestInit): string {
    return `${this.getBaseURL()}:${endpoint}:${JSON.stringify(options)}`;
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

  /**
   * Refresh a cloud access token once for all requests that encounter the
   * same expired session at the same time. Without this guard, a dashboard
   * burst can send one refresh request per endpoint and invalidate the same
   * refresh token repeatedly.
   */
  private async refreshAccessToken(): Promise<string | null> {
    if (this.refreshInFlight) {
      return this.refreshInFlight;
    }

    const refreshToken = TokenManager.getRefreshToken();
    if (!refreshToken) {
      return null;
    }

    const refreshPromise = (async () => {
      const refreshResponse = await fetch(`${getCloudApiUrl()}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      const refreshData = await refreshResponse.json().catch(() => ({}));
      if (!refreshResponse.ok) {
        const error: any = new Error(
          refreshData?.error?.message || refreshData?.error || 'Session refresh failed'
        );
        error.status = refreshResponse.status;
        error.response = refreshData;
        throw error;
      }

      const refreshPayload = refreshData?.data && typeof refreshData.data === 'object'
        ? refreshData.data
        : refreshData;
      const newToken = refreshPayload?.access_token || refreshPayload?.token;
      const newRefreshToken = refreshPayload?.refresh_token
        || refreshPayload?.refreshToken
        || refreshToken;

      if (!newToken) {
        return null;
      }

      this.setToken(newToken);
      if (newRefreshToken) {
        TokenManager.setRefreshToken(newRefreshToken);
      }
      return newToken as string;
    })();

    this.refreshInFlight = refreshPromise;
    try {
      return await refreshPromise;
    } finally {
      if (this.refreshInFlight === refreshPromise) {
        this.refreshInFlight = null;
      }
    }
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
      return { success: response.ok, data: undefined as T, error: response.ok ? undefined : { code: String(response.status), message: response.statusText } };
    }
    try {
      return JSON.parse(text) as ApiResponse<T>;
    } catch {
      return {
        success: false,
        data: undefined as T,
        error: { code: String(response.status), message: text },
      };
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const baseURL = this.getBaseURL();
    const url = `${baseURL}${endpoint}`;
    const authHeader = TokenManager.getToken();

    if (authHeader && !this.token) {
      this.token = authHeader;
    }

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
            // Tokens belong to the cloud authority even though business data
            // is served by the local SQLite API. Refresh against Render so a
            // valid cloud session can continue using local operations.
            const newToken = await this.refreshAccessToken();
            if (newToken) {
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
          } catch (refreshError) {
            console.error('Token refresh failed:', refreshError);
          }

          // Do not automatically log the user out.
          // Keep the current session alive and let the user continue
          // until they explicitly choose to log out or log in again.
          const refreshError: any = new Error('Session refresh failed. Your session will remain active until you log out manually.');
          refreshError.status = 401;
          refreshError.code = 'AUTH_REFRESH_FAILED';
          refreshError.response = data;
          throw refreshError;
        }

        if (response.status === 403) {
          if (typeof window !== 'undefined' && !window.location.hash.includes('/subscription-expired')) {
            try {
              window.location.hash = '#/subscription-expired';
              window.dispatchEvent(new HashChangeEvent('hashchange'));
            } catch {
              const currentUrl = new URL(window.location.href);
              currentUrl.hash = '#/subscription-expired';
              window.location.href = currentUrl.toString();
            }
          }
          const subscriptionError: any = new Error('اشتراكك منتهي، يرجى التواصل مع الإدارة لتجديد الخدمة.');
          subscriptionError.status = 403;
          subscriptionError.code = 'SUBSCRIPTION_EXPIRED';
          subscriptionError.arabicMessage = 'اشتراكك منتهي، يرجى التواصل مع الإدارة لتجديد الخدمة.';
          throw subscriptionError;
        }
        
        // Create error object with status
        const responseError = data.error as ApiResponse<T>['error'] | string | undefined;
        const errorMessage = typeof responseError === 'string'
          ? responseError
          : responseError?.message || 'An error occurred';
        const error: any = new Error(errorMessage);
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

  private async ensureMutationAllowed(endpoint: string): Promise<void> {
    // Business data may be written to the local API/SQLite database, but the
    // cloud remains the authority for whether the account may use the app.
    if (endpoint.startsWith('/auth/')) return;

    if (typeof navigator !== 'undefined' && !navigator.onLine) {
      throw new Error('يلزم اتصال بالإنترنت للتحقق من الاشتراك قبل تنفيذ العملية.');
    }

    const token = TokenManager.getToken();
    if (!token) {
      throw new Error('يلزم تسجيل الدخول قبل تنفيذ العملية.');
    }

    let response: Response;
    try {
      response = await fetch(`${getCloudApiUrl()}/auth/validate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: '{}',
      });
    } catch {
      throw new Error('تعذر الاتصال بالخادم للتحقق من الاشتراك. لم تُنفذ العملية.');
    }

    if (!response.ok) {
      if (response.status === 401 || response.status === 403) {
        this.logout();
        if (typeof window !== 'undefined') window.location.hash = '#/subscription-expired';
      }
      throw new Error('الحساب غير نشط أو أن الاشتراك منتهٍ. لم تُنفذ العملية.');
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
    
    const result = await this.requestWithRetry<T>(url, {
      method: 'GET',
      ...(useCache ? {} : { cache: 'no-store' as RequestCache }),
    });
    
    if (useCache) {
      this.setCache(cacheKey, result);
    }
    
    return result;
  }

  async post<T = any>(endpoint: string, body: any): Promise<ApiResponse<T>> {
    await this.ensureMutationAllowed(endpoint);
    this.clearCachePattern(endpoint.split('/')[1]); // Clear cache for related endpoints
    return this.requestWithRetry<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  }

  async put<T = any>(endpoint: string, body: any): Promise<ApiResponse<T>> {
    await this.ensureMutationAllowed(endpoint);
    this.clearCachePattern(endpoint.split('/')[1]); // Clear cache for related endpoints
    return this.requestWithRetry<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(body),
    });
  }

  async patch<T = any>(endpoint: string, body: any): Promise<ApiResponse<T>> {
    await this.ensureMutationAllowed(endpoint);
    this.clearCachePattern(endpoint.split('/')[1]); // Clear cache for related endpoints
    return this.requestWithRetry<T>(endpoint, {
      method: 'PATCH',
      body: JSON.stringify(body),
    });
  }

  async delete<T = any>(endpoint: string, params?: any): Promise<ApiResponse<T>> {
    await this.ensureMutationAllowed(endpoint);
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
