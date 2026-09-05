const STORAGE_KEY = 'partflow.local.product-images';

type ProductImages = Record<string, string>;

function readImages(): ProductImages {
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    return value ? (JSON.parse(value) as ProductImages) : {};
  } catch {
    return {};
  }
}

export function getLocalProductImage(productId: string): string | undefined {
  return readImages()[String(productId)];
}

export function setLocalProductImage(productId: string, image: string | null): void {
  const images = readImages();
  if (image) images[String(productId)] = image;
  else delete images[String(productId)];
  localStorage.setItem(STORAGE_KEY, JSON.stringify(images));
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
