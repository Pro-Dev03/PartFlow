import { create } from 'zustand';

interface UIState {
  sidebarCollapsed: boolean;
  checkoutMode: boolean;
  theme: 'light' | 'dark' | 'system';
  language: string;
  toggleSidebar: () => void;
  setCheckoutMode: (checkoutMode: boolean) => void;
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  setLanguage: (language: string) => void;
}

const storedTheme = typeof window !== 'undefined' ? localStorage.getItem('theme') : null;
const initialTheme: UIState['theme'] =
  storedTheme === 'light' || storedTheme === 'dark' ? storedTheme : 'system';

export const useUIStore = create<UIState>((set) => ({
  sidebarCollapsed: false,
  checkoutMode: false,
  theme: initialTheme,
  language: 'ar',
  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
  setCheckoutMode: (checkoutMode) => set({ checkoutMode }),
  setTheme: (theme) => set({ theme }),
  setLanguage: (language) => set({ language }),
}));