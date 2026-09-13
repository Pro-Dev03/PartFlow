const STORAGE_KEY = 'partflow.local.category-images';

type CategoryImages = Record<string, string>;
let cachedImages: CategoryImages = {};
let initialized = false;

function readImages(): CategoryImages {
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    return value ? (JSON.parse(value) as CategoryImages) : {};
  } catch {
    return {};
  }
}

function isDesktopStorageAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean(window.partflowDesktop?.categoryImages);
}

export async function initializeCategoryImages(): Promise<void> {
  if (initialized) return;
  initialized = true;
  const legacyImages = readImages();
  cachedImages = legacyImages;

  if (!isDesktopStorageAvailable()) return;

  try {
    const storedImages = await window.partflowDesktop!.categoryImages!.list();
    cachedImages = { ...legacyImages, ...(storedImages || {}) };
    for (const [categoryId, image] of Object.entries(legacyImages)) {
      if (!storedImages?.[categoryId]) {
        await window.partflowDesktop!.categoryImages!.save(categoryId, image);
      }
    }
    localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Keep browser storage available if the desktop bridge is unavailable.
  }
}

export function getCategoryImage(categoryId: string): string | undefined {
  const key = String(categoryId);
  if (!initialized) cachedImages = { ...readImages(), ...cachedImages };
  return cachedImages[key];
}

export function setCategoryImage(categoryId: string, image: string | null): void {
  const key = String(categoryId);
  if (image) cachedImages[key] = image;
  else delete cachedImages[key];

  if (isDesktopStorageAvailable()) {
    void (image
      ? window.partflowDesktop!.categoryImages!.save(key, image)
      : window.partflowDesktop!.categoryImages!.delete(key));
    return;
  }

  localStorage.setItem(STORAGE_KEY, JSON.stringify(cachedImages));
}

export { compressProductImage as compressCategoryImage } from './localProductImages';
