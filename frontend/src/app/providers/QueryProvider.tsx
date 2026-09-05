import { MutationCache, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';

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
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}