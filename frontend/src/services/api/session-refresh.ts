import { logSessionDiagnostic } from './sessionDiagnostics';

type RefreshPayload = Record<string, any>;
export type SessionAuthFailureKind = 'subscription_expired' | 'user_deleted' | 'user_disabled' | 'session_invalid' | 'temporary' | 'unclassified';
type BrowserLockManager = {
  request<T>(name: string, options: { mode: 'exclusive' }, callback: () => Promise<T>): Promise<T>;
};

const refreshesInFlight = new Map<string, Promise<RefreshPayload>>();

const confirmedBlockKinds: Record<string, SessionAuthFailureKind> = {
  SUBSCRIPTION_EXPIRED: 'subscription_expired',
  ACCOUNT_DELETED: 'user_deleted',
  ACCOUNT_SUSPENDED: 'user_disabled',
};

const invalidSessionCodes = new Set([
  'AUTH_REFRESH_FAILED',
  'INVALID_TOKEN',
  'REFRESH_TOKEN_INVALID',
  'REFRESH_TOKEN_REVOKED',
  'SESSION_EXPIRED',
  'SESSION_INVALID',
  'TOKEN_EXPIRED',
]);

const temporaryAuthCodes = new Set([
  'AUTH_REFRESH_PENDING',
  'AUTH_REFRESH_RESPONSE_INVALID',
  'AUTH_SERVICE_UNAVAILABLE',
  'CLOUD_CONNECTION_REQUIRED',
  'NETWORK_ERROR',
  'OFFLINE_GRACE_EXPIRED',
  'TIMEOUT',
  'TIMEOUT_ERROR',
]);

/** Extract a server code from the response and error wrappers used by our APIs. */
export function extractAuthErrorCode(value: unknown): string | undefined {
  const codes: string[] = [];
  const seen = new Set<object>();
  const visit = (candidate: unknown, depth = 0): void => {
    if (depth > 5 || !candidate || typeof candidate !== 'object' || seen.has(candidate as object)) return;
    seen.add(candidate as object);
    const record = candidate as Record<string, unknown>;
    if (typeof record.code === 'string' && record.code.trim()) codes.push(record.code.trim().toUpperCase());
    for (const key of ['error', 'data', 'response', 'body', 'cause']) visit(record[key], depth + 1);
  };

  if (typeof value === 'string') codes.push(value.trim().toUpperCase());
  else visit(value);
  return codes.find((code) => code in confirmedBlockKinds)
    ?? codes.find((code) => invalidSessionCodes.has(code))
    ?? codes.find(Boolean);
}

export function classifyAuthFailure(value: unknown): SessionAuthFailureKind {
  const code = extractAuthErrorCode(value);
  if (code && confirmedBlockKinds[code]) return confirmedBlockKinds[code];
  if (code && invalidSessionCodes.has(code)) return 'session_invalid';
  if (code && temporaryAuthCodes.has(code)) return 'temporary';

  const record = value as { status?: number; response?: { status?: number }; name?: string; cause?: { code?: string } } | null;
  const status = Number(record?.status ?? record?.response?.status);
  const transportCode = String(record?.cause?.code ?? '').toUpperCase();
  if (record?.name === 'AbortError' || record?.name === 'TypeError'
    || [0, 408, 425, 429, 500, 502, 503, 504].includes(status)
    || ['ECONNREFUSED', 'ECONNRESET', 'ENOTFOUND', 'ETIMEDOUT', 'EAI_AGAIN'].includes(transportCode)) {
    return 'temporary';
  }
  if (status === 401) return 'session_invalid';
  return 'unclassified';
}

function refreshLockName(apiBaseUrl: string): string {
  return `partflow-auth-refresh:${new URL(apiBaseUrl).origin}`;
}

function responseCode(payload: any): string | undefined {
  return extractAuthErrorCode(payload);
}

function makeRefreshError(status: number, code?: string, payload?: any): Error & { status: number; code?: string; response?: any } {
  const error = new Error(payload?.error?.message || (typeof payload?.error === 'string' ? payload.error : '') || 'Session refresh failed') as Error & {
    status: number;
    code?: string;
    response?: any;
  };
  error.status = status;
  error.code = code;
  error.response = payload;
  return error;
}

async function performRefresh(apiBaseUrl: string): Promise<RefreshPayload> {
  const endpoint = `${apiBaseUrl.replace(/\/+$/, '')}/auth/refresh`;
  const lockManager = typeof navigator !== 'undefined'
    ? (navigator as Navigator & { locks?: BrowserLockManager }).locks
    : undefined;

  const send = async () => {
    logSessionDiagnostic('refresh.attempt', { apiOrigin: new URL(apiBaseUrl).origin });
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 30_000);
    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        credentials: 'include',
        signal: controller.signal,
        headers: { 'Content-Type': 'application/json' },
        body: '{}',
      });
      const payload = await response.json().catch(() => ({}));
      const code = responseCode(payload);
      logSessionDiagnostic('refresh.result', {
        apiOrigin: new URL(apiBaseUrl).origin,
        status: response.status,
        code,
        ok: response.ok,
      });
      if (!response.ok) throw makeRefreshError(response.status, code, payload);

      const data = payload?.data && typeof payload.data === 'object' ? payload.data : payload;
      if (!(data?.access_token || data?.token)) {
        const error = makeRefreshError(response.status, 'AUTH_REFRESH_RESPONSE_INVALID', payload);
        throw error;
      }
      return data as RefreshPayload;
    } catch (error) {
      if (typeof (error as { status?: number })?.status !== 'number') {
        logSessionDiagnostic('refresh.transport-failure', {
          apiOrigin: new URL(apiBaseUrl).origin,
          errorName: error instanceof Error ? error.name : 'UnknownError',
        });
      }
      throw error;
    } finally {
      clearTimeout(timeoutId);
    }
  };

  if (lockManager) {
    return lockManager.request(refreshLockName(apiBaseUrl), { mode: 'exclusive' }, send);
  }
  logSessionDiagnostic('refresh.cross-tab-lock-unavailable', { apiOrigin: new URL(apiBaseUrl).origin });
  return send();
}

function rejectWhenAborted<T>(promise: Promise<T>, signal?: AbortSignal): Promise<T> {
  if (!signal) return promise;
  if (signal.aborted) return Promise.reject(new DOMException('The operation was aborted', 'AbortError'));

  return new Promise<T>((resolve, reject) => {
    const onAbort = () => reject(new DOMException('The operation was aborted', 'AbortError'));
    signal.addEventListener('abort', onAbort, { once: true });
    promise.then(resolve, reject).finally(() => signal.removeEventListener('abort', onAbort));
  });
}

/** Share refresh work inside a tab and serialize cookie rotation across tabs. */
export function refreshSessionCookie(apiBaseUrl: string, signal?: AbortSignal): Promise<RefreshPayload> {
  const key = new URL(apiBaseUrl).origin;
  let refresh = refreshesInFlight.get(key);
  if (!refresh) {
    refresh = performRefresh(apiBaseUrl);
    refreshesInFlight.set(key, refresh);
    const current = refresh;
    void refresh.finally(() => {
      if (refreshesInFlight.get(key) === current) refreshesInFlight.delete(key);
    }).catch(() => undefined);
  } else {
    logSessionDiagnostic('refresh.joined-in-flight', { apiOrigin: key });
  }
  return rejectWhenAborted(refresh, signal);
}

export function isDefinitiveRefreshRejection(error: unknown): boolean {
  const record = error as { status?: number; response?: { status?: number } } | null;
  const status = Number(record?.status ?? record?.response?.status);
  if (status !== 401 && status !== 403) return false;
  const kind = classifyAuthFailure(error);
  return kind === 'subscription_expired' || kind === 'user_deleted' || kind === 'user_disabled';
}

export function isSubscriptionRefreshRejection(error: unknown): boolean {
  return classifyAuthFailure(error) === 'subscription_expired';
}
