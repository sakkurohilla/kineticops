export interface ApiError {
  message: string;
  status?: number;
  code?: string;
}

export const handleApiError = (error: any): ApiError => {
  if (error.response) {
    return {
      message: error.response.data?.error || error.response.data?.message || 'Server error',
      status: error.response.status,
      code: error.response.data?.code
    };
  }
  
  if (error.request) {
    return {
      message: 'Network error - please check your connection',
      status: 0
    };
  }
  
  return {
    message: error.message || 'An unexpected error occurred'
  };
};

export const isNetworkError = (error: ApiError): boolean => {
  return error.status === 0 || error.status === undefined;
};

export const isAuthError = (error: ApiError): boolean => {
  return error.status === 401 || error.status === 403;
};