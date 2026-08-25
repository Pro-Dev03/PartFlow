import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { authApi } from '../services/api/endpoints';
import { apiClient } from '../services/api/client';
import { TokenManager } from '../lib/token-manager';

interface AuthState {
  isAuthenticated: boolean;
  user: any;
  token: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => void;
  refreshToken: () => Promise<void>;
}

let refreshInterval: ReturnType<typeof setTimeout> | null = null;

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      isAuthenticated: false,
      user: null,
      token: null,
      isLoading: false,

      login: async (email: string, password: string) => {
        set({ isLoading: true });
        try {
          const response = await authApi.login(email, password);
          const data = response.data as any;
          const { user, token } = data;

          // Use TokenManager for consistent token storage
          TokenManager.setToken(token);
          apiClient.setToken(token);

          set({
            isAuthenticated: true,
            user,
            token,
            isLoading: false
          });

          // Start auto-refresh
          startTokenRefresh();
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: () => {
        // Stop auto-refresh
        stopTokenRefresh();
        TokenManager.clearToken();
        apiClient.logout();
        set({
          isAuthenticated: false,
          user: null,
          token: null
        });
      },

      checkAuth: () => {
        // Use TokenManager for consistent token retrieval
        const token = TokenManager.getToken();
        if (token) {
          apiClient.setToken(token);
          set({
            isAuthenticated: true,
            token
          });
          
          // Start auto-refresh if token exists
          startTokenRefresh();
        }
      },

      refreshToken: async () => {
        try {
          const response = await authApi.refreshToken();
          const data = response.data as any;
          const { user, token } = data;

          apiClient.setToken(token);
          set({
            isAuthenticated: true,
            user,
            token
          });
        } catch (error) {
          console.error('Failed to refresh token:', error);
          // If refresh fails, logout
          get().logout();
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        token: state.token
      }),
    }
  )
);

// Auto-refresh token every 10 minutes
function startTokenRefresh() {
  stopTokenRefresh(); // Clear any existing interval
  
  refreshInterval = setInterval(async () => {
    const { isAuthenticated, token } = useAuthStore.getState();
    if (isAuthenticated && token) {
      try {
        await useAuthStore.getState().refreshToken();
      } catch (error) {
        console.error('Auto-refresh failed:', error);
      }
    }
  }, 10 * 60 * 1000); // 10 minutes
}

function stopTokenRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
}