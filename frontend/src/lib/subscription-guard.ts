export type LocalSubscriptionGuard = {
  userId?: string;
  email?: string;
  subscriptionStatus?: string;
  subscriptionExpiresAt?: string | null;
  token?: string | null;
  issuedAt?: string;
};

const STORAGE_KEY = 'partflow-subscription-guard';
const CLOCK_DRIFT_MS = 5 * 60 * 1000;

function readGuard(): LocalSubscriptionGuard | null {
  if (typeof window === 'undefined') return null;

  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as LocalSubscriptionGuard;
    return parsed && typeof parsed === 'object' ? parsed : null;
  } catch {
    return null;
  }
}

export function writeGuard(guard: LocalSubscriptionGuard | null): void {
  if (typeof window === 'undefined') return;

  if (!guard) {
    localStorage.removeItem(STORAGE_KEY);
    return;
  }

  localStorage.setItem(STORAGE_KEY, JSON.stringify(guard));
}

export function clearSubscriptionGuard(): void {
  writeGuard(null);
}

function normalizeStatus(status?: string | null): string {
  return (status || 'active').toLowerCase();
}

function parseExpiry(value?: string | null): number | null {
  if (!value) return null;

  const parsed = new Date(value).getTime();
  return Number.isFinite(parsed) ? parsed : null;
}

function decodeJwtClaims(token?: string | null): Record<string, unknown> | null {
  if (!token) return null;

  const parts = token.split('.');
  if (parts.length < 2) return null;

  try {
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, '=');
    const json = atob(padded);
    return JSON.parse(json) as Record<string, unknown>;
  } catch {
    return null;
  }
}

export function buildSubscriptionGuardFromToken(token?: string | null, fallback?: Partial<LocalSubscriptionGuard>): LocalSubscriptionGuard | null {
  if (!token) return null;

  const claims = decodeJwtClaims(token);
  const statusFromToken = typeof claims?.subscription_status === 'string'
    ? claims.subscription_status
    : typeof claims?.subscriptionStatus === 'string'
      ? claims.subscriptionStatus
      : fallback?.subscriptionStatus;

  const expiresAtFromToken = typeof claims?.subscription_expires_at === 'string'
    ? claims.subscription_expires_at
    : typeof claims?.subscriptionExpiresAt === 'string'
      ? claims.subscriptionExpiresAt
      : fallback?.subscriptionExpiresAt;

  return {
    userId: fallback?.userId ?? (typeof claims?.user_id === 'string' ? claims.user_id : undefined),
    email: fallback?.email ?? (typeof claims?.email === 'string' ? claims.email : undefined),
    subscriptionStatus: statusFromToken || 'active',
    subscriptionExpiresAt: expiresAtFromToken || null,
    token,
    issuedAt: new Date().toISOString(),
  };
}

export function isOfflineSubscriptionBlocked(token?: string | null, fallback?: Partial<LocalSubscriptionGuard>): boolean {
  const activeRecord = readGuard() || buildSubscriptionGuardFromToken(token, fallback);

  if (!activeRecord) return false;

  const normalizedStatus = normalizeStatus(activeRecord.subscriptionStatus);
  if (normalizedStatus === 'canceled' || normalizedStatus === 'cancelled' || normalizedStatus === 'expired') {
    return true;
  }

  const expiryMs = parseExpiry(activeRecord.subscriptionExpiresAt ?? null);
  if (expiryMs === null) {
    return false;
  }

  return Date.now() > expiryMs + CLOCK_DRIFT_MS;
}

export function syncSubscriptionGuard(token?: string | null, fallback?: Partial<LocalSubscriptionGuard>): void {
  const guard = buildSubscriptionGuardFromToken(token, fallback) || readGuard();
  if (!guard) {
    clearSubscriptionGuard();
    return;
  }

  writeGuard(guard);
}
