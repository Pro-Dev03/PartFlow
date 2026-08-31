import { contextBridge } from 'electron';

// The desktop shell always starts the embedded SQLite backend. There is no
// user-selectable online/offline mode; internet availability is enforced by
// the renderer and the backend's cloud-authority middleware.
contextBridge.exposeInMainWorld('partflowDesktop', {
  appVersion: process.versions.electron,
  platform: process.platform,
});
