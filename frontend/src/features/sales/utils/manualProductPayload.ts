export type ManualProductInput = {
  name: string;
  price: string;
  quantity: string;
  barcode?: string;
  sku?: string;
};

export function buildManualProductPayload(input: ManualProductInput) {
  const price = Number(input.price) || 0;
  const quantity = Number(input.quantity) || 1;
  const rawBarcode = (input.barcode ?? '').trim();
  const generatedSku = input.sku?.trim() || `SKU-${Date.now()}`;

  return {
    name: input.name.trim(),
    sku: generatedSku,
    barcode: rawBarcode || undefined,
    selling_price: price,
    cost_price: Number((price * 0.7).toFixed(2)),
    stock: quantity,
    condition: 'new',
    min_stock_level: 1,
  };
}
