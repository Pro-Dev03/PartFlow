import { getArabicErrorMessage, isRetryableError } from '../../lib/error-messages';
import { appConfig, getActiveApiUrl, getCloudApiUrl, getLocalApiUrl, shouldUseLocalApi } from '../../lib/config/app';
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
  private refreshFailedForSession = false;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    // Use TokenManager for consistent token retrieval
    this.token = TokenManager.getToken();
  }

  setToken(token: string) {
    this.token = token;
    this.refreshFailedForSession = false;
    // Use TokenManager for consistent token storage
    TokenManager.setToken(token);
  }

  clearToken() {
    this.token = null;
    this.refreshFailedForSession = false;
    // Use TokenManager for consistent token clearing
    TokenManager.clearToken();
  }

  logout() {
    this.clearToken();
    this.clearCache();
    if (typeof window !== 'undefined') {
      localStorage.removeItem('cloud_token');
      localStorage.removeItem('cloud_refresh_token');
    }
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
    if (this.refreshFailedForSession) {
      return null;
    }

    if (this.refreshInFlight) {
      return this.refreshInFlight;
    }

    const refreshToken = TokenManager.getRefreshToken();
    if (!refreshToken) {
      this.refreshFailedForSession = true;
      return null;
    }

    const refreshPromise = (async () => {
      const baseUrl = typeof window !== 'undefined' && shouldUseLocalApi(window.location.hostname)
        ? getLocalApiUrl()
        : getCloudApiUrl();

      const refreshResponse = await fetch(`${baseUrl}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      const refreshData = await refreshResponse.json().catch(() => ({}));
      if (!refreshResponse.ok) {
        this.refreshFailedForSession = true;
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
        this.refreshFailedForSession = true;
        return null;
      }

      this.refreshFailedForSession = false;
      this.setToken(newToken);
      if (newRefreshToken) {
        TokenManager.setRefreshToken(newRefreshToken);
      }
      return newToken as string;
    })();

    this.refreshInFlight = refreshPromise;
    try {
      return await refreshPromise;
    } catch (error) {
      this.refreshFailedForSession = true;
      throw error;
    } finally {
      if (this.refreshInFlight === refreshPromise) {
        this.refreshInFlight = null;
      }
    }
  }

  private async createLocalSessionFromCloud(): Promise<string | null> {
    const cloudToken = this.getCloudAccessToken();
    if (!cloudToken) return null;

    const response = await fetch(`${getLocalApiUrl()}/auth/cloud-session`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ cloud_token: cloudToken }),
    });
    const payload = await response.json().catch(() => ({}));
    if (!response.ok) return null;

    const data = payload?.data ?? payload;
    const token = data?.access_token || data?.token;
    if (!token) return null;

    this.setToken(token);
    if (data?.refresh_token) {
      TokenManager.setRefreshToken(data.refresh_token);
    }
    return token;
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

      // A failed refresh cannot be fixed by retrying the same request. Let the
      // auth store handle the invalid session instead of creating a request loop.
      if (error?.code === 'AUTH_REFRESH_FAILED' || error?.code === 'INVALID_TOKEN') {
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
    if (typeof response.text === 'function') {
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

    return {
      success: response.ok,
      data: undefined as T,
      error: response.ok ? undefined : { code: String(response.status), message: response.statusText || 'Request failed' },
    };
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

    const cloudToken = this.getCloudAccessToken();
    if (cloudToken) {
      headers['X-PartFlow-Cloud-Token'] = cloudToken;
      if (options.method === 'POST' && endpoint.startsWith('/settings/sync')) {
        headers['Authorization'] = `Bearer ${cloudToken}`;
      }
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
        // Local JWT 401s refresh against the local API; when the local backend
        // also validates the cloud session before accepting the request, the
        // stale cloud token must be refreshed and re-attached before retrying.
        if (response.status === 401 && (this.token || this.getCloudAccessToken())) {
          try {
            let newToken: string | null = null;
            let refreshAttempted = false;
            const cloudTokenBeforeRefresh = this.getCloudAccessToken();

            if (this.token) {
              refreshAttempted = true;
              newToken = await this.refreshAccessToken().catch(async refreshError => {
                if (refreshError?.status === 401 && this.getCloudAccessToken()) {
                  return this.createLocalSessionFromCloud();
                }
                throw refreshError;
              });
              if (newToken) {
                this.setToken(newToken);
                headers['Authorization'] = `Bearer ${newToken}`;
              }
            }

            let refreshedCloudToken: string | null = null;
            if (cloudTokenBeforeRefresh) {
              refreshedCloudToken = await this.refreshCloudAccessToken();
              if (refreshedCloudToken) {
                headers['X-PartFlow-Cloud-Token'] = refreshedCloudToken;

                // A local JWT can be signed by an older local API instance or
                // secret even when the cloud session refresh succeeds. Rebuild
                // the local session from the fresh cloud token so local
                // middleware and the retried request use the same issuer.
                const localSessionToken = await this.createLocalSessionFromCloud();
                if (localSessionToken) {
                  newToken = localSessionToken;
                  this.setToken(localSessionToken);
                  headers['Authorization'] = `Bearer ${localSessionToken}`;
                }
              }
            }

            const retryToken = newToken ?? this.token;
            const retryCloudToken = refreshedCloudToken ?? this.getCloudAccessToken();

            // Only retry after a successful refresh. If the session refresh is
            // rejected, the stale local token must not be used for another request
            // and the caller should receive AUTH_REFRESH_FAILED instead of looping.
            const hasFreshCredentials = Boolean(newToken || refreshedCloudToken);
            if (hasFreshCredentials) {
              if (retryToken) {
                headers['Authorization'] = `Bearer ${retryToken}`;
              }
              if (retryCloudToken) {
                headers['X-PartFlow-Cloud-Token'] = retryCloudToken;
              }

              // Retry request with the latest local and cloud credentials.
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

            if (refreshAttempted && this.refreshFailedForSession) {
              throw Object.assign(new Error('Session refresh failed. Your session will remain active until you log out manually.'), {
                status: 401,
                code: 'AUTH_REFRESH_FAILED',
                response: data,
              });
            }
          } catch (refreshError) {
            console.error('Token refresh failed:', refreshError);

            if ((refreshError as any)?.status === 401 || (refreshError as any)?.code === 'INVALID_TOKEN') {
              throw refreshError;
            }
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

        // For non-cloud requests or non-401 errors, create error object with status
        const responseError = data.error as ApiResponse<T>['error'] | string | undefined;
        const errorMessage = typeof responseError === 'string'
          ? responseError
          : responseError?.message || 'An error occurred';
        const error: any = new Error(errorMessage);
        error.status = response.status;
        error.code = data.error?.code;
        error.response = data;

        // A 403 is not always a subscription expiry (for example, the
        // administrator-only settings routes intentionally return
        // ADMIN_REQUIRED). Redirect only for an explicit subscription/cloud
        // authorization decision and leave ordinary permission errors to the
        // caller.
        if (response.status === 403 && (
          data.error?.code === 'SUBSCRIPTION_EXPIRED' ||
          data.error?.code === 'CLOUD_AUTH_REQUIRED'
        )) {
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
          error.message = 'اشتراكك منتهي، يرجى التواصل مع الإدارة لتجديد الخدمة.';
          error.code = 'SUBSCRIPTION_EXPIRED';
          error.arabicMessage = error.message;
          throw error;
        }
        
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

  private getCloudAccessToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('cloud_token');
  }

  private async refreshCloudAccessToken(): Promise<string | null> {
    if (typeof window === 'undefined') return null;
    const refreshToken = localStorage.getItem('cloud_refresh_token');
    if (!refreshToken) return null;

    const refreshResponse = await fetch(`${getCloudApiUrl()}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    const refreshData = await refreshResponse.json().catch(() => ({}));
    if (!refreshResponse.ok) return null;

    const payload = refreshData?.data && typeof refreshData.data === 'object'
      ? refreshData.data
      : refreshData;
    const nextToken = payload?.access_token || payload?.token;
    if (!nextToken) return null;

    localStorage.setItem('cloud_token', nextToken);
    const nextRefresh = payload?.refresh_token || payload?.refreshToken;
    if (nextRefresh) {
      localStorage.setItem('cloud_refresh_token', nextRefresh);
    }
    return nextToken as string;
  }

  private async ensureMutationAllowed(endpoint: string): Promise<void> {
    // Business data may be written to the local API/SQLite database, but the
    // cloud remains the authority for whether the account may use the app.
    if (endpoint.startsWith('/auth/')) return;

    if (typeof navigator !== 'undefined' && !navigator.onLine) {
      throw new Error('يلزم اتصال بالإنترنت للتحقق من الاشتراك قبل تنفيذ العملية.');
    }

    let cloudToken = this.getCloudAccessToken();
    if (!cloudToken) {
      throw new Error('يلزم تسجيل الدخول قبل تنفيذ العملية.');
    }

    const validate = (token: string) => fetch(`${getCloudApiUrl()}/auth/validate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        'X-PartFlow-Cloud-Token': token,
      },
      body: '{}',
    });

    let response: Response;
    try {
      response = await validate(cloudToken);
      if (response.status === 401) {
        const refreshed = await this.refreshCloudAccessToken();
        if (refreshed) {
          cloudToken = refreshed;
          response = await validate(cloudToken);
        }
      }
    } catch {
      throw new Error('تعذر الاتصال بالخادم للتحقق من الاشتراك. لم تُنفذ العملية.');
    }

    if (!response.ok) {
      this.logout();
      if (typeof window !== 'undefined') {
        window.location.hash = response.status === 403
          ? '#/subscription-expired'
          : '#/login';
      }
      throw new Error(response.status === 403
        ? 'الحساب غير نشط أو أن الاشتراك منتهٍ. لم تُنفذ العملية.'
        : 'يلزم تسجيل الدخول قبل تنفيذ العملية.');
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
