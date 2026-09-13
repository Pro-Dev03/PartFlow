export const DEFAULT_PROFIT_MARGIN = 30;

export function calculateSuggestedSellingPrice(costPrice: number, profitMargin: number): number {
  if (!Number.isFinite(costPrice) || costPrice <= 0 || !Number.isFinite(profitMargin) || profitMargin < 0 || profitMargin >= 100) {
    return 0;
  }

  return Math.round(costPrice / (1 - profitMargin / 100));
}
