import { barcodeApi } from '../services/api/endpoints';

export type ProductBarcodeLookupSource = 'cache' | 'local-catalog' | 'upcitemdb' | 'openfoodfacts' | 'searxng-web';

export interface ProductBarcodeLookupResult {
  barcode: string;
  source: ProductBarcodeLookupSource;
  name: string;
  brand?: string;
  category?: string;
  description?: string;
  size?: string;
  specifications?: string[];
  image?: string;
  model?: string;
  confidence?: number;
}

export function clearBarcodeLookupFields<T extends { name: string }>(row: T): T {
  const next = { ...row, name: '' } as T & Record<string, unknown>;
  for (const field of ['description', 'image', 'image_url', 'brand', 'category', 'model', 'size', 'specifications']) {
    if (field in row) next[field] = Array.isArray(row[field]) ? [] : '';
  }
  return next;
}

const STORAGE_KEY = 'partflow-product-barcode-cache';

function cleanText(value: unknown): string {
  if (typeof value !== 'string') return '';
  return value.trim();
}

export function normalizeBarcode(raw: string): string | null {
  const normalized = raw.replace(/[\s-]/g, '');
  if (!/^\d{8,14}$/.test(normalized)) return null;
  return normalized;
}

function readLookupCache(): Record<string, ProductBarcodeLookupResult> {
  try {
    if (typeof localStorage === 'undefined') return {};
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch {
    return {};
  }
}

function writeLookupCache(cache: Record<string, ProductBarcodeLookupResult>) {
  try {
    if (typeof localStorage === 'undefined') return;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(cache));
  } catch {
    // Ignore storage restrictions.
  }
}

export function normalizeBarcodeLookupResult(raw: Partial<ProductBarcodeLookupResult> & { barcode?: string; name?: string; source?: string }): ProductBarcodeLookupResult {
  const barcode = normalizeBarcode(cleanText(raw.barcode || '')) || '';
  const name = cleanText(raw.name) || 'منتج غير مسمى';
  const category = cleanText(raw.category) || 'General';
  const description = cleanText(raw.description) || '';
  const size = cleanText(raw.size) || '';
  const model = cleanText(raw.model) || '';
  const validSources: ProductBarcodeLookupSource[] = ['cache', 'local-catalog', 'upcitemdb', 'openfoodfacts', 'searxng-web'];
  const source = validSources.includes(raw.source as ProductBarcodeLookupSource)
    ? raw.source as ProductBarcodeLookupSource
    : 'cache';

  return {
    barcode,
    source,
    name,
    brand: cleanText(raw.brand),
    category,
    description,
    size,
    specifications: Array.isArray(raw.specifications) ? raw.specifications.map((spec) => cleanText(spec)).filter(Boolean) : [],
    image: cleanText(raw.image),
    model,
    confidence: Number.isFinite(raw.confidence) ? Number(raw.confidence) : 0,
  };
}

export async function lookupProductByBarcode(barcode: string): Promise<ProductBarcodeLookupResult | null> {
  const normalized = normalizeBarcode(cleanText(barcode));
  if (!normalized) return null;

  const cache = readLookupCache();
  const cached = cache[normalized];
  if (cached) {
    return normalizeBarcodeLookupResult({ ...cached, source: 'cache' });
  }

  let response;
  try {
    response = await barcodeApi.lookupProductMetadata(normalized);
  } catch (error: any) {
    if (error?.status === 404 || error?.code === 'BARCODE_NO_RESULT') return null;
    throw error;
  }
  const payload = (response as any)?.data ?? response;
  if (!payload?.name) return null;

  const merged = normalizeBarcodeLookupResult(payload);

  const nextCache = readLookupCache();
  nextCache[normalized] = merged;
  writeLookupCache(nextCache);

  return merged;
}
