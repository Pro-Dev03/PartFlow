import { contextBridge, ipcRenderer } from 'electron';

// The desktop shell always starts the embedded SQLite backend. There is no
// user-selectable online/offline mode; internet availability is enforced by
// the renderer and the backend's cloud-authority middleware.
contextBridge.exposeInMainWorld('partflowDesktop', {
  appVersion: process.versions.electron,
  platform: process.platform,
  printHtml: (html) => ipcRenderer.invoke('document:print-html', html),
  invoice: {
    print: (payload) => ipcRenderer.invoke('invoice:print', payload),
    savePdf: (payload) => ipcRenderer.invoke('invoice:save-pdf', payload),
  },
  productImages: {
    list: () => ipcRenderer.invoke('product-images:list'),
    save: (productId, dataUrl) => ipcRenderer.invoke('product-images:save', productId, dataUrl),
    delete: (productId) => ipcRenderer.invoke('product-images:delete', productId),
  },
    partTypeImages: {
      list: () => ipcRenderer.invoke('part-type-images:list'),
      save: (partTypeId, dataUrl) => ipcRenderer.invoke('part-type-images:save', partTypeId, dataUrl),
      delete: (partTypeId) => ipcRenderer.invoke('part-type-images:delete', partTypeId),
    },
  categoryImages: {
    list: () => ipcRenderer.invoke('category-images:list'),
    save: (categoryId, dataUrl) => ipcRenderer.invoke('category-images:save', categoryId, dataUrl),
    delete: (categoryId) => ipcRenderer.invoke('category-images:delete', categoryId),
  },
});
