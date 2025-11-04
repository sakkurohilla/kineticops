import axios from 'axios';
import authService from '../auth/authService';

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

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
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

// Response interceptor
apiClient.interceptors.response.use(
  (response) => {
    return response.data;
  },
  async (error) => {


    // Handle network errors (server down)
    if (error.code === 'ECONNREFUSED' || error.code === 'ERR_NETWORK' || !error.response) {
      authService.logout();
      window.location.href = '/login';
      return Promise.reject({
        message: 'Server is unavailable. Please try again later.',
        status: 503,
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

export { BASE_URL };
export default apiClient;
