const STORAGE_KEY = 'partflow.local.product-images';

type ProductImages = Record<string, string>;
let cachedImages: ProductImages = {};
let initialized = false;

function readImages(): ProductImages {
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    return value ? (JSON.parse(value) as ProductImages) : {};
  } catch {
    return {};
  }
}

function isDesktopStorageAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean(window.partflowDesktop?.productImages);
}

export async function initializeProductImages(): Promise<void> {
  if (initialized) return;
  initialized = true;
  const legacyImages = readImages();
  cachedImages = legacyImages;

  if (!isDesktopStorageAvailable()) return;

  try {
    const storedImages = await window.partflowDesktop!.productImages!.list();
    cachedImages = { ...legacyImages, ...(storedImages || {}) };
    for (const [productId, image] of Object.entries(legacyImages)) {
      if (!storedImages?.[productId]) {
        await window.partflowDesktop!.productImages!.save(productId, image);
      }
    }
    localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Keep the legacy browser storage available if the desktop bridge is unavailable.
  }
}

export function getLocalProductImage(productId: string): string | undefined {
  const key = String(productId);
  if (!initialized) cachedImages = { ...readImages(), ...cachedImages };
  return cachedImages[key];
}

export function setLocalProductImage(productId: string, image: string | null): void {
  const key = String(productId);
  if (image) cachedImages[key] = image;
  else delete cachedImages[key];

  if (isDesktopStorageAvailable()) {
    void (image
      ? window.partflowDesktop!.productImages!.save(key, image)
      : window.partflowDesktop!.productImages!.delete(key));
    return;
  }

  localStorage.setItem(STORAGE_KEY, JSON.stringify(cachedImages));
}

export function compressProductImage(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('تعذر قراءة الصورة'));
    reader.onload = () => {
      const image = new Image();
      image.onerror = () => reject(new Error('الصورة غير صالحة'));
      image.onload = () => {
        const maxSize = 640;
        const scale = Math.min(1, maxSize / Math.max(image.width, image.height));
        const canvas = document.createElement('canvas');
        canvas.width = Math.max(1, Math.round(image.width * scale));
        canvas.height = Math.max(1, Math.round(image.height * scale));
        const context = canvas.getContext('2d');
        if (!context) {
          reject(new Error('تعذر تجهيز الصورة'));
          return;
        }
        context.drawImage(image, 0, 0, canvas.width, canvas.height);
        resolve(canvas.toDataURL('image/jpeg', 0.82));
      };
      image.src = String(reader.result);
    };
    reader.readAsDataURL(file);
  });
}
