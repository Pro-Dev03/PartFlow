import { useEffect, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { settingsApi } from '../services/api/endpoints';
import { toast } from 'sonner';

interface InitialDataSyncState {
  isLoading: boolean;
  isError: boolean;
  error: Error | null;
  progress: number; // 0-100
}

export function useInitialDataSync(
  enabled: boolean = false,
  onSuccess?: () => void,
  onError?: (error: Error) => void,
  userId?: string
): InitialDataSyncState {
  const [progress, setProgress] = useState(0);
  const queryClient = useQueryClient();

  const { isLoading, isError, error } = useQuery({
    queryKey: ['sync', 'initial-data', userId ?? 'default'],
    queryFn: async () => {
      setProgress(10);

      try {
        // Ask the local API to perform the authenticated cloud -> SQLite sync.
        // The browser must not keep a second, stale copy of the operational
        // database in localStorage.
        const response = await settingsApi.syncCloudData();
        setProgress(50);

        if (!response.success) {
          throw new Error(response.error || 'Failed to fetch initial data');
        }

        const data = response.data || {};
        setProgress(80);

        // Mark sync as complete
        localStorage.setItem(initialSyncStorageKey(userId), 'true');
        setProgress(100);

        // The sync mutates SQLite behind React Query's back. Invalidate every
        // operational query so dashboard, customers, debts and reports read
        // the newly downloaded snapshot instead of an older cached response.
        // Refresh active operational queries in the background. Do not await
        // this promise: waiting here keeps the full-screen sync modal open at
        // 100% while those requests perform their own cloud checks, which can
        // leave navigation blocked indefinitely.
        void queryClient.invalidateQueries();

        // Show success message
        const tables = (data as { tables?: Record<string, number> }).tables;
        const totalItems = tables
          ? Object.values(tables).reduce((sum, count) => sum + (Number(count) || 0), 0)
          : Object.values(data as Record<string, unknown>)
              .filter(Array.isArray)
              .reduce((sum, arr) => sum + (arr as unknown[]).length, 0);
        
        toast.success(`✓ تم تحميل ${totalItems} عنصر من الخادم`);

        // Call onSuccess callback if provided
        if (onSuccess) {
          onSuccess();
        }

        return data;
      } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));
        setProgress(0);
        
        toast.error(`خطأ في التحميل: ${err.message}`);

        // Call onError callback if provided
        if (onError) {
          onError(err);
        }

        throw err;
      }
    },
    enabled,
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    staleTime: Infinity, // Don't refetch initial data automatically
    gcTime: Infinity,    // Don't garbage collect this data
  });

  // Reset progress when query state changes
  useEffect(() => {
    if (!isLoading && progress > 0 && progress < 100) {
      setProgress(0);
    }
  }, [isLoading, progress]);

  return {
    isLoading,
    isError,
    error: error instanceof Error ? error : null,
    progress,
  };
}

// Helper function to check if initial sync is needed
function initialSyncStorageKey(userId?: string): string {
  return userId ? `partflow-initial-sync-complete:${userId}` : 'partflow-initial-sync-complete';
}

export function isInitialSyncNeeded(userId?: string): boolean {
  const syncComplete = localStorage.getItem(initialSyncStorageKey(userId));

  // The desktop business database is SQLite in both operating modes.  The
  // first authenticated launch must therefore hydrate it from the cloud
  // regardless of the UI's legacy mode flag.
  return syncComplete !== 'true';
}

// Helper function to get cached initial sync data
export function getCachedInitialSyncData(): any {
  // Kept as a compatibility no-op for older callers. Operational rows live
  // in SQLite and are never reconstructed from browser storage.
  return null;
}

// Helper function to clear initial sync data
export function clearInitialSyncData(userId?: string): void {
  localStorage.removeItem(initialSyncStorageKey(userId));
}
