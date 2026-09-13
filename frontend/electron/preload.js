import { contextBridge, ipcRenderer } from 'electron';

// The desktop shell always starts the embedded SQLite backend. There is no
// user-selectable online/offline mode; internet availability is enforced by
// the renderer and the backend's cloud-authority middleware.
contextBridge.exposeInMainWorld('partflowDesktop', {
  appVersion: process.versions.electron,
  platform: process.platform,
  productImages: {
    list: () => ipcRenderer.invoke('product-images:list'),
    save: (productId, dataUrl) => ipcRenderer.invoke('product-images:save', productId, dataUrl),
    delete: (productId) => ipcRenderer.invoke('product-images:delete', productId),
  },
});
