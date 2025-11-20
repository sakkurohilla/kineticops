import axios from 'axios';
import authService from '../auth/authService';
import { apiRateLimiter } from '../../utils/rateLimiter';
import { apiCircuitBreaker } from '../../utils/circuitBreaker';

// Force detect network IP
const getBaseURL = (): string => {
  const hostname = window.location.hostname;
  const protocol = window.location.protocol;
  
  // If not localhost, use the same hostname for API
  if (hostname !== 'localhost' && hostname !== '127.0.0.1') {
    return `${protocol}//${hostname}:8080/api/v1`;
  }
  
  // Default localhost
  return 'http://localhost:8080/api/v1';
};

const BASE_URL = getBaseURL();

const apiClient = axios.create({
  baseURL: BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor with rate limiting and circuit breaker
apiClient.interceptors.request.use(
  async (config) => {
    // Rate limiting check
    if (!apiRateLimiter.canMakeRequest()) {
      const waitTime = apiRateLimiter.getWaitTime();
      console.warn(`[API] Rate limit exceeded. Wait ${waitTime}ms`);
      throw new Error(`Rate limit exceeded. Please wait ${Math.ceil(waitTime / 1000)}s`);
    }

    // CSRF token for unsafe methods
    if (config.method && ['post', 'put', 'delete', 'patch'].includes(config.method.toLowerCase())) {
      const csrfToken = getCookie('csrf_token');
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken;
      }
    }

    const token = authService.getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor with circuit breaker
apiClient.interceptors.response.use(
  (response) => {
    return response.data;
  },
  async (error) => {
    // Handle network errors (server down) - DON'T logout, just show error
    if (error.code === 'ECONNREFUSED' || error.code === 'ERR_NETWORK' || !error.response) {
      console.warn('[API] Server connection error, will retry...');
      
      // Circuit breaker pattern
      if (apiCircuitBreaker.getState() === 'OPEN') {
        return Promise.reject({
          message: 'Service temporarily unavailable. Please try again later.',
          status: 503,
          isNetworkError: true,
        });
      }

      return Promise.reject({
        message: 'Server is unavailable. Retrying...',
        status: 503,
        isNetworkError: true,
      });
    }

    // Handle 401
    if (error.response?.status === 401 && error.config && !error.config._retry) {
      error.config._retry = true;
      try {
        const newToken = await authService.refreshToken();
        error.config.headers.Authorization = `Bearer ${newToken}`;
        return apiClient(error.config);
      } catch (refreshError) {
        authService.logout();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject({
      message: error.response?.data?.error || error.message,
      status: error.response?.status,
    });
  }
);

// Helper to get cookie value
function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(';').shift() || null;
  return null;
}

export { BASE_URL };
export default apiClient;
