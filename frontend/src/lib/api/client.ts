import { appConfig } from '../config/app';
import type { ApiResponse } from '../../types/api';
import { getErrorMessage, isAuthError, shouldRetry } from '../error-handling';

interface ApiError {
  code: string;
  message: string;
}

class ApiClient {
  private baseURL: string;
  private token: string | null = null;

  constructor() {
    this.baseURL = appConfig.apiUrl;
    this.loadToken();
  }

  private loadToken() {
    this.token = localStorage.getItem('auth_token');
  }

  private saveToken(token: string) {
    this.token = token;
    localStorage.setItem('auth_token', token);
  }

  private clearToken() {
    this.token = null;
    localStorage.removeItem('auth_token');
  }

  setToken(token: string) {
    this.saveToken(token);
  }

  logout() {
    this.clearToken();
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    return headers;
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        code: 'UNKNOWN_ERROR',
        message: 'An unknown error occurred',
      }));
      
      // Use Arabic error messages
      const arabicMessage = getErrorMessage(error);
      
      // Check for auth errors
      if (response.status === 401 || isAuthError(error)) {
        this.clearToken();
        // Don't redirect here, let the caller handle it
      }
      
      throw new Error(arabicMessage);
    }

    const data: ApiResponse<T> = await response.json();
    return data.data;
  }

  async get<T>(endpoint: string, params?: Record<string, string>, retryCount = 0): Promise<T> {
    const url = new URL(`${this.baseURL}${endpoint}`);
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, value);
      });
    }

    try {
      const response = await fetch(url.toString(), {
        method: 'GET',
        headers: this.getHeaders(),
      });

      return this.handleResponse<T>(response);
    } catch (error: any) {
      // Retry logic for network errors
      if (retryCount < 2 && shouldRetry(error)) {
        await new Promise(resolve => setTimeout(resolve, 1000 * (retryCount + 1)));
        return this.get<T>(endpoint, params, retryCount + 1);
      }
      throw error;
    }
  }

  async post<T>(endpoint: string, data?: any, retryCount = 0): Promise<T> {
    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        method: 'POST',
        headers: this.getHeaders(),
        body: data ? JSON.stringify(data) : undefined,
      });

      return this.handleResponse<T>(response);
    } catch (error: any) {
      if (retryCount < 2 && shouldRetry(error)) {
        await new Promise(resolve => setTimeout(resolve, 1000 * (retryCount + 1)));
        return this.post<T>(endpoint, data, retryCount + 1);
      }
      throw error;
    }
  }

  async put<T>(endpoint: string, data?: any, retryCount = 0): Promise<T> {
    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        method: 'PUT',
        headers: this.getHeaders(),
        body: data ? JSON.stringify(data) : undefined,
      });

      return this.handleResponse<T>(response);
    } catch (error: any) {
      if (retryCount < 2 && shouldRetry(error)) {
        await new Promise(resolve => setTimeout(resolve, 1000 * (retryCount + 1)));
        return this.put<T>(endpoint, data, retryCount + 1);
      }
      throw error;
    }
  }

  async patch<T>(endpoint: string, data?: any, retryCount = 0): Promise<T> {
    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        method: 'PATCH',
        headers: this.getHeaders(),
        body: data ? JSON.stringify(data) : undefined,
      });

      return this.handleResponse<T>(response);
    } catch (error: any) {
      if (retryCount < 2 && shouldRetry(error)) {
        await new Promise(resolve => setTimeout(resolve, 1000 * (retryCount + 1)));
        return this.patch<T>(endpoint, data, retryCount + 1);
      }
      throw error;
    }
  }

  async delete<T>(endpoint: string, retryCount = 0): Promise<T> {
    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        method: 'DELETE',
        headers: this.getHeaders(),
      });

      return this.handleResponse<T>(response);
    } catch (error: any) {
      if (retryCount < 2 && shouldRetry(error)) {
        await new Promise(resolve => setTimeout(resolve, 1000 * (retryCount + 1)));
        return this.delete<T>(endpoint, retryCount + 1);
      }
      throw error;
    }
  }
}

export const apiClient = new ApiClient();