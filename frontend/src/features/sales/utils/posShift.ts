export interface PosShiftState {
  status: 'open' | 'closed';
  openedAt: string;
  openingCash: number;
  closedAt?: string;
  closingCash?: number;
  salesTotal: number;
  saleCount: number;
}

export function createDefaultPosShift(): PosShiftState {
  return {
    status: 'closed',
    openedAt: new Date().toISOString(),
    openingCash: 0,
    salesTotal: 0,
    saleCount: 0,
  };
}

export function resolvePosShiftState(value: unknown): PosShiftState {
  if (!value || typeof value !== 'object') {
    return createDefaultPosShift();
  }

  const shift = value as Record<string, unknown>;
  if (shift.status !== 'open' && shift.status !== 'closed') {
    return createDefaultPosShift();
  }

  return {
    status: shift.status === 'closed' ? 'closed' : 'open',
    openedAt: String(shift.opened_at || new Date().toISOString()),
    openingCash: Number(shift.opening_cash) || 0,
    closedAt: typeof shift.closed_at === 'string' ? shift.closed_at : undefined,
    closingCash: shift.closing_cash == null ? undefined : Number(shift.closing_cash) || 0,
    salesTotal: Number(shift.sales_total) || 0,
    saleCount: Number(shift.sale_count) || 0,
  };
}
