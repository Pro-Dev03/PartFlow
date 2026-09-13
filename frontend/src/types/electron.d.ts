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
    };
  }
}
