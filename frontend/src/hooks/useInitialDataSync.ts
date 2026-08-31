import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { syncApi } from '../services/api/endpoints';
import { getCloudApiUrl } from '../lib/config/app';
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
  onError?: (error: Error) => void
): InitialDataSyncState {
  const [progress, setProgress] = useState(0);

  const { isLoading, isError, error } = useQuery({
    queryKey: ['sync', 'initial-data'],
    queryFn: async () => {
      setProgress(10);

      try {
        // Fetch initial data from cloud
        const response = await syncApi.getInitialData();
        setProgress(50);

        if (!response.success) {
          throw new Error(response.error || 'Failed to fetch initial data');
        }

        const data = response.data;
        
        // Store data in localStorage (or IndexedDB for large datasets)
        const syncData = {
          timestamp: new Date().toISOString(),
          data,
        };

        localStorage.setItem('partflow-initial-sync-data', JSON.stringify(syncData));
        setProgress(80);

        // Mark sync as complete
        localStorage.setItem('partflow-initial-sync-complete', 'true');
        setProgress(100);

        // Show success message
        const totalItems = Object.values(data)
          .filter(Array.isArray)
          .reduce((sum, arr) => sum + arr.length, 0);
        
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
export function isInitialSyncNeeded(): boolean {
  const mode = localStorage.getItem('partflow-operating-mode');
  const syncComplete = localStorage.getItem('partflow-initial-sync-complete');
  
  return mode === 'online' && syncComplete !== 'true';
}

// Helper function to get cached initial sync data
export function getCachedInitialSyncData(): any {
  try {
    const data = localStorage.getItem('partflow-initial-sync-data');
    if (data) {
      return JSON.parse(data);
    }
  } catch (error) {
    console.error('Failed to parse cached initial sync data:', error);
  }
  return null;
}

// Helper function to clear initial sync data
export function clearInitialSyncData(): void {
  localStorage.removeItem('partflow-initial-sync-data');
  localStorage.removeItem('partflow-initial-sync-complete');
}
