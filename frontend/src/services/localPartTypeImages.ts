const STORAGE_KEY = 'partflow.local.part-type-images';

type PartTypeImages = Record<string, string>;
let cachedImages: PartTypeImages = {};
let initialized = false;

function readImages(): PartTypeImages {
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    return value ? (JSON.parse(value) as PartTypeImages) : {};
  } catch {
    return {};
  }
}

function isDesktopStorageAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean(window.partflowDesktop?.partTypeImages);
}

export async function initializePartTypeImages(): Promise<void> {
  if (initialized) return;
  initialized = true;
  const legacyImages = readImages();
  cachedImages = legacyImages;

  if (!isDesktopStorageAvailable()) return;

  try {
    const storedImages = await window.partflowDesktop!.partTypeImages!.list();
    cachedImages = { ...legacyImages, ...(storedImages || {}) };
    for (const [partTypeId, image] of Object.entries(legacyImages)) {
      if (!storedImages?.[partTypeId]) {
        await window.partflowDesktop!.partTypeImages!.save(partTypeId, image);
      }
    }
    localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Keep browser storage available if the desktop bridge is unavailable.
  }
}

export function getPartTypeImage(partTypeId: string): string | undefined {
  const key = String(partTypeId);
  if (!initialized) cachedImages = { ...readImages(), ...cachedImages };
  return cachedImages[key];
}

export function setPartTypeImage(partTypeId: string, image: string | null): void {
  const key = String(partTypeId);
  if (image) cachedImages[key] = image;
  else delete cachedImages[key];

  if (isDesktopStorageAvailable()) {
    void (image
      ? window.partflowDesktop!.partTypeImages!.save(key, image)
      : window.partflowDesktop!.partTypeImages!.delete(key));
    return;
  }

  localStorage.setItem(STORAGE_KEY, JSON.stringify(cachedImages));
}

export { compressProductImage as compressPartTypeImage } from './localProductImages';
