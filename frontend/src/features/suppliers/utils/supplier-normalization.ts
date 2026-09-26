export type SupplierSummaryLike = {
  id?: string | number;
  name?: string;
  code?: string;
  phone?: string;
  email?: string;
  totalPurchases?: number;
  total_purchases?: number;
  paidAmount?: number;
  paid_amount?: number;
  outstanding?: number;
  credit_balance?: number;
  current_balance?: number;
  lastPurchase?: string | null;
  last_purchase?: string | null;
  [key: string]: unknown;
};

export function normalizeSupplier(raw: SupplierSummaryLike) {
  return {
    ...raw,
    id: raw.id ?? '',
    name: raw.name ?? '',
    code: raw.code ?? '-',
    phone: raw.phone ?? '',
    email: raw.email ?? '',
    totalPurchases: Number(raw.totalPurchases ?? raw.total_purchases ?? 0),
    paidAmount: Number(raw.paidAmount ?? raw.paid_amount ?? 0),
    outstanding: Math.max(Number(raw.outstanding ?? raw.current_balance ?? 0), 0),
    creditBalance: Number(raw.credit_balance ?? Math.max(-Number(raw.outstanding ?? raw.current_balance ?? 0), 0)),
    lastPurchase: raw.lastPurchase ?? raw.last_purchase ?? null,
  };
}
