import { useQuery, UseQueryOptions, UseQueryResult } from 'react-query'
import { apiClient, handleApiError, ApiResponse } from '../../services/api/apiClient'

// Generic hook for GET requests
export function useApiQuery<T>(
  queryKey: string | string[],
  endpoint: string,
  options?: Omit<UseQueryOptions<T>, 'queryKey' | 'queryFn'>
): UseQueryResult<T> {
  return useQuery<T>(
    queryKey,
    async () => {
      const response = await apiClient.get<ApiResponse<T>>(endpoint)
      return response.data.data
    },
    {
      ...options,
      onError: (error) => {
        console.error('Query error:', handleApiError(error))
        options?.onError?.(error as any)
      },
    }
  )
}

// Hook for paginated queries
export function usePaginatedQuery<T>(
  queryKey: string | string[],
  endpoint: string,
  params?: Record<string, any>,
  options?: Omit<UseQueryOptions<T>, 'queryKey' | 'queryFn'>
): UseQueryResult<T> {
  return useQuery<T>(
    [...queryKey, params],
    async () => {
      const response = await apiClient.get<ApiResponse<T>>(endpoint, { params })
      return response.data.data
    },
    {
      ...options,
      enabled: !!params, // Only run when params are provided
      onError: (error) => {
        console.error('Paginated query error:', handleApiError(error))
        options?.onError?.(error as any)
      },
    }
  )
}