import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig } from 'axios'

// API Client Configuration
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // Request interceptor
  client.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      // Add auth token if available
      const token = localStorage.getItem('auth_token')
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`
      }
      
      // Add organization ID if available
      const orgId = localStorage.getItem('organization_id')
      if (orgId && config.headers) {
        config.headers['X-Organization-ID'] = orgId
      }

      return config
    },
    (error) => {
      return Promise.reject(error)
    }
  )

  // Response interceptor
  client.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
      if (error.response?.status === 401) {
        // Handle unauthorized - redirect to login
        localStorage.removeItem('auth_token')
        window.location.href = '/login'
      }

      if (error.response?.status === 403) {
        // Handle forbidden - show permission error
        console.error('Permission denied')
      }

      if (error.response?.status === 500) {
        // Handle server error
        console.error('Server error')
      }

      return Promise.reject(error)
    }
  )

  return client
}

export const apiClient = createApiClient()

// Helper function to handle API errors
export const handleApiError = (error: any): string => {
  if (error.response?.data?.message) {
    return error.response.data.message
  }
  
  if (error.response?.data?.error) {
    return error.response.data.error
  }
  
  if (error.message) {
    return error.message
  }
  
  return 'حدث خطأ غير متوقع'
}

// API Response types
export interface ApiResponse<T> {
  success: boolean
  data: T
  meta?: {
    page?: number
    limit?: number
    total?: number
    totalPages?: number
  }
  error?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}