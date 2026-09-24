export function normalizePurchaseStatus(status?: string | null): string {
  const normalized = (status ?? '').toString().trim().toLowerCase();

  if (!normalized) return '';

  const compact = normalized
    .replace(/[_\-\s]+/g, '_')
    .replace(/_+/g, '_')
    .replace(/^_|_$/g, '');

  if (['completed', 'complete', 'done'].includes(compact)) return 'received';
  if (['canceled', 'cancelled'].includes(compact)) return 'cancelled';
  if (['partially_received', 'partiallyreceived', 'partiallyreceived_purchase'].includes(compact)) return 'partially_received';
  if (['reversed', 'reverse'].includes(compact)) return 'reversed';
  if (['pending', 'draft', 'ordered', 'processing'].includes(compact)) return compact;

  return compact;
}

export function isReceivedPurchaseStatus(status?: string | null): boolean {
  const normalized = normalizePurchaseStatus(status);
  return normalized === 'received' || normalized === 'partially_received';
}

export function matchesPurchaseViewFilter(status?: string | null, viewFilter?: string): boolean {
  const normalized = normalizePurchaseStatus(status);

  if (!viewFilter || viewFilter === 'all') return true;
  if (viewFilter === 'closed') return normalized === 'reversed' || normalized === 'cancelled';
  if (viewFilter === 'received') return isReceivedPurchaseStatus(normalized);
  if (viewFilter === 'active') return !['reversed', 'cancelled'].includes(normalized) && !isReceivedPurchaseStatus(normalized);

  return normalized === viewFilter;
}
