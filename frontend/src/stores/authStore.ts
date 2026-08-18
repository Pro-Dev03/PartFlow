import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { authApi } from '../services/api/endpoints';
import type { UserRole } from '../lib/permissions';

interface AuthState {
  isAuthenticated: boolean;
  user: any;
  token: string | null;
  isLoading: boolean;
  role: UserRole;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      user: null,
      token: null,
      isLoading: false,
      role: 'employee',

      login: async (email: string, password: string) => {
        set({ isLoading: true });
        try {
          const response = await authApi.login(email, password);
          const data = response.data as any;
          const { user, token } = data;
          set({
            isAuthenticated: true,
            user,
            token,
            role: user?.role || 'employee',
            isLoading: false
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: () => {
        set({
          isAuthenticated: false,
          user: null,
          token: null,
          role: 'employee'
        });
      },

      checkAuth: () => {
        const token = localStorage.getItem('auth_token');
        if (token) {
          set({
            isAuthenticated: true,
            token
          });
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        token: state.token,
        role: state.role
      }),
    }
  )
);