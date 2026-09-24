import { getArabicErrorMessage, isNetworkError, isRetryableError } from '../../lib/error-messages';
import { appConfig, getActiveApiUrl, getCloudApiUrl, getConnectionMode, getLocalApiUrl, shouldUseLocalApi } from '../../lib/config/app';
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

function isExpectedBarcodeMiss(error: any): boolean {
  if (error?.status !== 404) return false;
  const message = String(error?.message || error?.response?.error?.message || '').toLowerCase();
  return error?.code === '404' || message.includes('barcode not found');
}

const SUBSCRIPTION_BLOCK_CODES = new Set([
  'SUBSCRIPTION_EXPIRED',
  'SUBSCRIPTION_SUSPENDED',
  'ACCOUNT_SUSPENDED',
  'ACCOUNT_DELETED',
]);

const MAX_RETRIES = 3;
const RETRY_DELAY = 1000; // 1 second
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

class ApiClient {
  private baseURL: string;
  private token: string | null = null;
  private cloudToken: string | null = null;
  private cache: Map<string, { data: unknown; timestamp: number }> = new Map();
  private refreshInFlight: Promise<string | null> | null = null;
  private refreshFailedForSession = false;
  private authInvalidationDispatched = false;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    this.syncSessionFromStorage();
  }

  private syncSessionFromStorage() {
    this.token = TokenManager.getToken();
    this.cloudToken = this.getCloudAccessToken();
  }

  setToken(token: string) {
    this.token = token;
    this.refreshFailedForSession = false;
    this.authInvalidationDispatched = false;
    TokenManager.setToken(token);
  }

  setCloudToken(cloudToken: string | null) {
    this.cloudToken = cloudToken ?? null;
    TokenManager.setCloudToken(cloudToken ?? null);
    if (typeof window !== 'undefined' && !cloudToken) {
      localStorage.removeItem('cloud_token');
    }
  }

  clearToken() {
    this.token = null;
    this.refreshFailedForSession = false;
    TokenManager.clearToken();
  }

  clearCloudToken() {
    this.cloudToken = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('cloud_token');
      localStorage.removeItem('cloud_refresh_token');
    }
  }

  logout() {
    this.clearToken();
    this.clearCloudToken();
    this.clearCache();
  }

  private notifyAuthInvalidated(reason: string, definitive = false, code?: string): void {
    if (this.authInvalidationDispatched || typeof window === 'undefined') return;

    this.authInvalidationDispatched = true;
    window.dispatchEvent(new CustomEvent('partflow:auth-invalidated', {
      detail: { reason, definitive, code },
    }));
  }

  private notifyCloudVerificationPending(reason: string): void {
    if (typeof window === 'undefined') return;

    window.dispatchEvent(new CustomEvent('partflow:cloud-verification-pending', {
      detail: { reason },
    }));
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

    const refreshPromise = (async () => {
      const baseUrl = typeof window !== 'undefined' && shouldUseLocalApi(window.location.hostname)
        ? getLocalApiUrl()
        : getCloudApiUrl();

      const refreshResponse = await fetch(`${baseUrl}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: '{}',
      });

      const refreshData = await refreshResponse.json().catch(() => ({}));
      if (!refreshResponse.ok) {
        this.refreshFailedForSession = refreshResponse.status === 401 || refreshResponse.status === 403;
        const error: any = new Error(
          refreshData?.error?.message || refreshData?.error || 'Session refresh failed'
        );
        error.status = refreshResponse.status;
        error.code = refreshData?.code || refreshData?.error?.code;
        error.response = refreshData;
        throw error;
      }

      const refreshPayload = refreshData?.data && typeof refreshData.data === 'object'
        ? refreshData.data
        : refreshData;
      const newToken = refreshPayload?.access_token || refreshPayload?.token;
      if (!newToken) {
        this.refreshFailedForSession = false;
        return null;
      }

      this.refreshFailedForSession = false;
      this.setToken(newToken);
      return newToken as string;
    })();

    this.refreshInFlight = refreshPromise;
    try {
      return await refreshPromise;
    } catch (error) {
      const status = Number((error as { status?: number })?.status);
      this.refreshFailedForSession = status === 401 || status === 403;
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
    if (!response.ok) {
      const code = payload?.code || payload?.error?.code;
      const isBlockedAccount = SUBSCRIPTION_BLOCK_CODES.has(String(code || ''))
        && (response.status === 401 || response.status === 403);
      if (!isBlockedAccount) {
        return null;
      }
      const error: any = new Error(payload?.error?.message || payload?.error || 'Failed to create local session');
      error.status = response.status;
      error.code = code;
      error.response = payload;
      throw error;
    }

    const data = payload?.data ?? payload;
    const token = data?.access_token || data?.token;
    if (!token) return null;

    this.setToken(token);
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

      if (error?.code === 'AUTH_REFRESH_PENDING') {
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
    this.syncSessionFromStorage();

    const baseURL = this.getBaseURL();
    const url = `${baseURL}${endpoint}`;
    const authHeader = TokenManager.getToken();

    if (authHeader && !this.token) {
      this.token = authHeader;
    }

    const normalizedHeaders = new Headers(options.headers ?? {});
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    normalizedHeaders.forEach((value, key) => {
      headers[key] = value;
    });

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const cloudToken = this.getCloudAccessToken();
    if (cloudToken) {
      headers['X-PartFlow-Cloud-Token'] = cloudToken;
      if (getConnectionMode() === 'cloud' || (options.method === 'POST' && endpoint.startsWith('/settings/sync'))) {
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
        credentials: 'include',
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
                if (refreshError?.status === 401 && this.getCloudAccessToken() && getConnectionMode() === 'local') {
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
                if (getConnectionMode() === 'local') {
                  const localSessionToken = await this.createLocalSessionFromCloud();
                  if (localSessionToken) {
                    newToken = localSessionToken;
                    this.setToken(localSessionToken);
                    headers['Authorization'] = `Bearer ${localSessionToken}`;
                  }
                } else {
                  newToken = refreshedCloudToken;
                  this.setToken(refreshedCloudToken);
                  headers['Authorization'] = `Bearer ${refreshedCloudToken}`;
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
                credentials: 'include',
              });
              const retryData: ApiResponse<T> = await this.parseResponse<T>(retryResponse);

              if (!retryResponse.ok) {
                const error: any = new Error(typeof retryData.error === 'string' ? retryData.error : retryData.error?.message || 'An error occurred');
                error.status = retryResponse.status;
                error.code = retryData.error?.code || (retryData as any).code;
                error.response = retryData;
                error.arabicMessage = getArabicErrorMessage(error);
                throw error;
              }

              return retryData;
            }

            if (refreshAttempted && this.refreshFailedForSession) {
              throw Object.assign(new Error('Session refresh failed; authentication is required again.'), {
                status: 401,
                code: 'AUTH_REFRESH_FAILED',
                response: data,
              });
            }
          } catch (refreshError) {
            console.error('Token refresh failed:', refreshError);

            if ((Number((refreshError as any)?.status) === 401 || Number((refreshError as any)?.status) === 403)
              && SUBSCRIPTION_BLOCK_CODES.has(String((refreshError as any)?.code || ''))) {
              this.notifyAuthInvalidated('Cloud subscription or account authorization was rejected', true, (refreshError as any)?.code);
              if (typeof window !== 'undefined' && !window.location.hash.includes('/subscription-expired')) {
                window.location.hash = '#/subscription-expired';
              }
              throw refreshError;
            }

            if ((refreshError as any)?.status === 401 || (refreshError as any)?.code === 'INVALID_TOKEN') {
              throw refreshError;
            }

            if (isNetworkError(refreshError) || [408, 429, 500, 502, 503, 504].includes(Number((refreshError as any)?.status))) {
              throw Object.assign(
                refreshError instanceof Error ? refreshError : new Error('Cloud session refresh is temporarily unavailable'),
                { code: 'AUTH_REFRESH_PENDING' },
              );
            }
          }

          // A rejected refresh proves that the local session cannot be trusted.
          // Invalidate it immediately instead of leaving a stale session active.
          const refreshError: any = new Error('Session refresh failed; authentication is required again.');
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
        error.code = data.error?.code || (data as any).code;
        error.response = data;

        // A 403 is not always a subscription expiry (for example, the
        // administrator-only settings routes intentionally return
        // ADMIN_REQUIRED). Redirect only for an explicit subscription/cloud
        // authorization decision and leave ordinary permission errors to the
        // caller.
        if (response.status === 403 && SUBSCRIPTION_BLOCK_CODES.has(error.code)) {
          this.notifyAuthInvalidated('Cloud subscription or account authorization was rejected', true, error.code);
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
          error.message = error.code === 'SUBSCRIPTION_SUSPENDED' || error.code === 'ACCOUNT_SUSPENDED'
            ? 'تم إيقاف الحساب، يرجى التواصل مع الإدارة لإعادة تفعيله.'
            : error.code === 'ACCOUNT_DELETED'
              ? 'هذا الحساب محذوف ولم يعد مسموحًا له بالدخول.'
              : 'اشتراكك منتهي، يرجى التواصل مع الإدارة لتجديد الخدمة.';
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

      if ((error?.status === 401 || error?.code === 'AUTH_REFRESH_FAILED' || error?.code === 'INVALID_TOKEN')
        && (this.token || this.getCloudAccessToken())) {
        this.notifyAuthInvalidated('Cloud session expired or was revoked', true, error?.code);
      }
      if (error?.code === 'AUTH_REFRESH_PENDING') {
        this.notifyCloudVerificationPending('Cloud session refresh is temporarily unavailable');
      }
      if (error?.code === 'OFFLINE_GRACE_EXPIRED' || error?.code === 'CLOUD_AUTH_REQUIRED') {
        this.notifyCloudVerificationPending('Cloud authorization is unavailable or the offline grace period has ended');
      }
      if (error?.code === 'AUTH_SERVICE_UNAVAILABLE'
        || [408, 429, 500, 502, 503, 504].includes(Number(error?.status))) {
        this.notifyCloudVerificationPending('Authentication service is temporarily unavailable');
      }

      const isExpectedAuthInvalidation = error?.code === 'SUBSCRIPTION_EXPIRED'
        || error?.code === 'SUBSCRIPTION_SUSPENDED'
        || error?.code === 'ACCOUNT_SUSPENDED'
        || error?.code === 'ACCOUNT_DELETED'
        || error?.code === 'CLOUD_AUTH_REQUIRED'
        || error?.code === 'OFFLINE_GRACE_EXPIRED'
        || error?.code === 'AUTH_SERVICE_UNAVAILABLE'
        || error?.code === 'AUTH_REFRESH_PENDING';
      const isExpectedLocalLoginFallback = getConnectionMode() === 'local'
        && endpoint === '/auth/login'
        && error?.status === 401;
      if (!isExpectedAuthInvalidation && !isExpectedBarcodeMiss(error) && !isExpectedLocalLoginFallback) {
        console.error('API request failed:', error);
      }

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
    return this.cloudToken ?? TokenManager.getCloudToken();
  }

  private async refreshCloudAccessToken(): Promise<string | null> {
    if (typeof window === 'undefined') return null;
    const refreshResponse = await fetch(`${getCloudApiUrl()}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
    });
    const refreshData = await refreshResponse.json().catch(() => ({}));
    if (!refreshResponse.ok) {
      if (refreshResponse.status >= 500 || refreshResponse.status === 408 || refreshResponse.status === 429) {
        const error = new Error('Cloud token refresh is temporarily unavailable') as Error & { status?: number };
        error.status = refreshResponse.status;
        (error as Error & { code?: string }).code = refreshData?.code || refreshData?.error?.code;
        throw error;
      }
      if (refreshResponse.status === 403) {
        const error: any = new Error(refreshData?.error?.message || refreshData?.error || 'Cloud account is not authorized');
        error.status = refreshResponse.status;
        error.code = refreshData?.code || refreshData?.error?.code;
        error.response = refreshData;
        throw error;
      }
      return null;
    }

    const payload = refreshData?.data && typeof refreshData.data === 'object'
      ? refreshData.data
      : refreshData;
    const nextToken = payload?.access_token || payload?.token;
    if (!nextToken) return null;

    TokenManager.setCloudToken(nextToken);
    this.cloudToken = nextToken;
    return nextToken as string;
  }

  private async ensureMutationAllowed(endpoint: string): Promise<void> {
    // Every protected API request is checked by the backend. The local API
    // applies the signed, bounded offline grant when the cloud is unreachable;
    // a frontend preflight would reject those valid grace-period requests.
    void endpoint;
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
    
    // Never serve cached business data without a fresh request. The local API
    // validates the cloud session on every request, so returning an in-memory
    // response here could allow a revoked account to keep using stale data.
    const result = await this.requestWithRetry<T>(url, {
      method: 'GET',
      ...(useCache ? {} : { cache: 'no-store' as RequestCache }),
    });
    
    if (useCache) {
      this.setCache(cacheKey, result);
    }
    
    return result;
  }

  async post<T = any>(endpoint: string, body: any, idempotencyKey?: string): Promise<ApiResponse<T>> {
    await this.ensureMutationAllowed(endpoint);
    this.clearCachePattern(endpoint.split('/')[1]); // Clear cache for related endpoints
    const headers = endpoint === '/sales' ? {
      'Idempotency-Key': idempotencyKey || globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`,
    } : undefined;
    return this.requestWithRetry<T>(endpoint, {
      method: 'POST',
      ...(headers ? { headers } : {}),
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
