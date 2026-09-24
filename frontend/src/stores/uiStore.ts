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

function readStoredTheme(): UIState['theme'] {
  if (typeof window === 'undefined') return 'light';
  try {
    const storedTheme = window.localStorage.getItem('theme');
    return storedTheme === 'dark' || storedTheme === 'system' ? storedTheme : 'light';
  } catch {
    return 'light';
  }
}

function applyTheme(theme: UIState['theme']) {
  if (typeof document === 'undefined') return;

  const root = document.documentElement;
  const isDark = theme !== 'light';
  root.classList.add('theme-switching');
  root.classList.toggle('dark', isDark);
  root.classList.toggle('light', !isDark);
  root.style.colorScheme = isDark ? 'dark' : 'light';
  window.requestAnimationFrame(() => root.classList.remove('theme-switching'));
}

const initialTheme = readStoredTheme();
applyTheme(initialTheme);

export const useUIStore = create<UIState>((set) => ({
  sidebarCollapsed: false,
  checkoutMode: false,
  theme: initialTheme,
  language: 'ar',
  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
  setCheckoutMode: (checkoutMode) => set({ checkoutMode }),
  setTheme: (theme) => {
    applyTheme(theme);
    try {
      window.localStorage.setItem('theme', theme);
    } catch {
      // Keep the in-memory theme usable when storage is unavailable.
    }
    set({ theme });
  },
  setLanguage: (language) => set({ language }),
}));
