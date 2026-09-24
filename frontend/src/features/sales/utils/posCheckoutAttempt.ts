export interface PosCheckoutAttempt {
  idempotencyKey: string;
  paymentOrderId: string;
}

const STORAGE_KEY = 'partflow.pos.checkout-attempt.v1';

function createToken(): string {
  return globalThis.crypto?.randomUUID?.()
    ?? `${Date.now()}-${Math.random().toString(36).slice(2, 12)}`;
}

export function getPosCheckoutAttempt(): PosCheckoutAttempt {
  try {
    const stored = sessionStorage.getItem(STORAGE_KEY);
    if (stored) {
      const attempt = JSON.parse(stored) as Partial<PosCheckoutAttempt>;
      if (typeof attempt.idempotencyKey === 'string' && typeof attempt.paymentOrderId === 'string') {
        return attempt as PosCheckoutAttempt;
      }
    }
  } catch {
    // Storage may be unavailable in a restricted browser context.
  }

  const attempt = {
    idempotencyKey: `pos-sale-${createToken()}`,
    paymentOrderId: createToken(),
  };

  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(attempt));
  } catch {
    // The request and server middleware still protect this attempt in memory.
  }

  return attempt;
}

export function clearPosCheckoutAttempt(): void {
  try {
    sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    // Ignore unavailable browser storage.
  }
}
