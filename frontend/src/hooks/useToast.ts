import { create } from 'zustand';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

interface ToastMessage {
  id: string;
  type: ToastType;
  title?: string;
  message: string;
  duration?: number;
}

interface ToastStore {
  toasts: ToastMessage[];
  addToast: (type: ToastType, message: string, title?: string, duration?: number) => void;
  removeToast: (id: string) => void;
  clearToasts: () => void;
}

export const useToastStore = create<ToastStore>((set) => ({
  toasts: [],
  addToast: (type, message, title, duration = 3000) => {
    const id = Date.now().toString();
    set((state) => ({
      toasts: [...state.toasts, { id, type, message, title, duration }],
    }));

    if (duration > 0) {
      setTimeout(() => {
        set((state) => ({
          toasts: state.toasts.filter((t) => t.id !== id),
        }));
      }, duration);
    }
  },
  removeToast: (id) => {
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    }));
  },
  clearToasts: () => set({ toasts: [] }),
}));

export function useToast() {
  const { addToast, removeToast, clearToasts } = useToastStore();

  return {
    success: (message: string, title?: string, duration?: number) => 
      addToast('success', message, title, duration),
    error: (message: string, title?: string, duration?: number) => 
      addToast('error', message, title, duration),
    warning: (message: string, title?: string, duration?: number) => 
      addToast('warning', message, title, duration),
    info: (message: string, title?: string, duration?: number) => 
      addToast('info', message, title, duration),
    remove: removeToast,
    clear: clearToasts,
  };
}
