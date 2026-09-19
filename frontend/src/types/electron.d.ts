export {};

declare global {
  interface Window {
    partflowDesktop?: {
      appVersion: string;
      platform: string;
      printHtml?: (html: string) => Promise<boolean>;
      invoice?: {
        print: (payload: unknown) => Promise<boolean>;
        savePdf: (payload: unknown) => Promise<{ canceled: boolean; filePath?: string }>;
      };
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
