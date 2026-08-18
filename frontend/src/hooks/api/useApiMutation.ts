import { useMutation, UseMutationOptions, UseMutationResult } from '@tanstack/react-query'
import { apiClient, handleApiError, ApiResponse } from '../../services/api/apiClient'

// Generic hook for POST requests
export function useApiMutation<TData, TVariables>(
  endpoint: string,
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE' = 'POST',
  options?: Omit<UseMutationOptions<TData, Error, TVariables>, 'mutationFn'>
): UseMutationResult<TData, Error, TVariables> {
  return useMutation<TData, Error, TVariables>(
    async (variables) => {
      const response = await apiClient.request<ApiResponse<TData>>({
        url: endpoint,
        method,
        data: variables,
      })
      return response.data.data
    },
    {
      ...options,
      onError: (error) => {
        console.error('Mutation error:', handleApiError(error))
        options?.onError?.(error)
      },
    }
  )
}

// Hook for DELETE requests
export function useApiDelete<T>(
  endpoint: string,
  options?: Omit<UseMutationOptions<T, Error, string>, 'mutationFn'>
): UseMutationResult<T, Error, string> {
  return useMutation<T, Error, string>(
    async (id) => {
      const response = await apiClient.delete<ApiResponse<T>>(`${endpoint}/${id}`)
      return response.data.data
    },
    {
      ...options,
      onError: (error) => {
        console.error('Delete error:', handleApiError(error))
        options?.onError?.(error)
      },
    }
  )
}

// Hook for file upload
export function useFileUpload<T>(
  endpoint: string,
  options?: Omit<UseMutationOptions<T, Error, File>, 'mutationFn'>
): UseMutationResult<T, Error, File> {
  return useMutation<T, Error, File>(
    async (file) => {
      const formData = new FormData()
      formData.append('file', file)

      const response = await apiClient.post<ApiResponse<T>>(endpoint, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
      return response.data.data
    },
    {
      ...options,
      onError: (error) => {
        console.error('Upload error:', handleApiError(error))
        options?.onError?.(error)
      },
    }
  )
}