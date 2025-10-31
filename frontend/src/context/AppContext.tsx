import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User } from '../types';
import authService from '../services/auth/authService';
import apiClient from '../services/api/client'; // ADD THIS IMPORT

interface AppContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  register: (username: string, email: string, password: string) => Promise<string>;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

interface AppProviderProps {
  children: ReactNode;
}

export const AppProvider: React.FC<AppProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    // Check if user is already logged in
    const token = authService.getToken();
    if (token) {
      fetchUser();
    } else {
      setIsLoading(false);
    }
  }, []);

  const fetchUser = async () => {
    try {
      const userData = await apiClient.get<User>('/auth/me');
      // apiClient may return the raw user object or an AxiosResponse depending on interceptors
      setUser(userData as any);
    } catch (error: any) {
      console.error('Failed to fetch user:', error);
      
      // Only logout if it's an auth error (401)
      // Don't logout for network errors or 429 rate limits
      if (error.message?.includes('401') || error.message?.includes('Invalid')) {
        authService.logout();
        setUser(null);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (username: string, password: string) => {
    try {
      await authService.login(username, password);
      
      // After login, fetch user details
      await fetchUser();
    } catch (error) {
      throw error;
    }
  };

  const logout = () => {
    authService.logout();
    setUser(null);
  };

  const register = async (username: string, email: string, password: string): Promise<string> => {
    try {
      const response = await authService.register(username, email, password);
      // Backend returns { msg: "User registered..." }
      return response.msg || 'Registration successful';
    } catch (error) {
      throw error;
    }
  };

  const value: AppContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    logout,
    register,
  };

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};

export const useApp = (): AppContextType => {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useApp must be used within an AppProvider');
  }
  return context;
};
