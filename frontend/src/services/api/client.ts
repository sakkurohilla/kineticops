import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import authService from '../auth/authService';

// Determine the API base URL based on environment
const getBaseURL = (): string => {
  // If VITE_API_URL is set in environment, use it
  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }

  // For production build
  if (import.meta.env.PROD) {
    return '/api/v1';
  }

  // For development - detect if accessing from network
  const hostname = window.location.hostname;
  
  // If accessing from network IP (not localhost/127.0.0.1)
  if (hostname !== 'localhost' && hostname !== '127.0.0.1') {
    // Use the same host but port 8080 for API
    return `http://${hostname}:8080/api/v1`;
  }

  // Default to localhost for local development
  return 'http://localhost:8080/api/v1';
};

const BASE_URL = getBaseURL();

console.log('[API Client] Using base URL:', BASE_URL);

const apiClient: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = authService.getToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// Define error response type
interface ApiErrorResponse {
  error?: string;
  message?: string;
}

// Response interceptor
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response.data;
  },
  async (error: AxiosError<ApiErrorResponse>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Handle 401 errors (unauthorized)
    if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // Try to refresh the token
        const newToken = await authService.refreshToken();
        
        // Update the failed request with new token
        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
        }
        
        // Retry the original request
        return apiClient(originalRequest);
      } catch (refreshError) {
        // Refresh failed, logout user
        authService.logout();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // Handle network errors
    if (error.message === 'Network Error' || !error.response) {
      return Promise.reject({
        message: 'No response from server. Please check your connection.',
        status: 0,
      });
    }

    // Return structured error
    return Promise.reject({
      message: error.response?.data?.error || error.response?.data?.message || error.message,
      status: error.response?.status,
    });
  }
);

export default apiClient;
