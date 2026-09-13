export {};

declare global {
  interface Window {
    partflowDesktop?: {
      appVersion: string;
      platform: string;
      productImages?: {
        list: () => Promise<Record<string, string>>;
        save: (productId: string, dataUrl: string) => Promise<string>;
        delete: (productId: string) => Promise<boolean>;
      };
      partTypeImages?: {
        list: () => Promise<Record<string, string>>;
        save: (partTypeId: string, dataUrl: string) => Promise<string>;
        delete: (partTypeId: string) => Promise<boolean>;
      };
      categoryImages?: {
        list: () => Promise<Record<string, string>>;
        save: (categoryId: string, dataUrl: string) => Promise<string>;
        delete: (categoryId: string) => Promise<boolean>;
      };
    };
  }
}
