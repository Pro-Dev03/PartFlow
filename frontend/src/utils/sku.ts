export function generateSku(): string {
  const randomPart = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID().replace(/-/g, '').slice(0, 10)
    : `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 8)}`;

  return `SKU-${randomPart.toUpperCase()}`;
}