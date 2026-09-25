import { MutationCache, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useEffect, type ReactNode } from 'react';

const queryClient = new QueryClient({
  mutationCache: new MutationCache({
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['reports'] });
      void queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  }),
  defaultOptions: {
    queries: {
      staleTime: 0,
      // The API client owns status-aware retries; avoid multiplying requests.
      retry: false,
      refetchOnWindowFocus: true,
    },
    mutations: {
      retry: false,
    },
  },
});

interface QueryProviderProps {
  children: ReactNode;
}

export function QueryProvider({ children }: QueryProviderProps) {
  useEffect(() => {
    const clearTenantData = () => {
      void queryClient.cancelQueries();
      queryClient.clear();
    };

    window.addEventListener('partflow:auth-invalidated', clearTenantData);
    window.addEventListener('partflow:session-cleared', clearTenantData);
    return () => {
      window.removeEventListener('partflow:auth-invalidated', clearTenantData);
      window.removeEventListener('partflow:session-cleared', clearTenantData);
    };
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
