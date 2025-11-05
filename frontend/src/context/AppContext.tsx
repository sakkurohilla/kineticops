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
  const [sessionTimeout, setSessionTimeout] = useState<NodeJS.Timeout | null>(null);

  useEffect(() => {
    // Check if user is already logged in
    const token = authService.getToken();
    if (token) {
      fetchUser();
      startSessionTimeout();
    } else {
      setIsLoading(false);
    }

    return () => {
      clearSessionTimeout();
    };
  }, []);

  const startSessionTimeout = () => {
    clearSessionTimeout();
    const timeout = setTimeout(() => {
      logout();
      alert('Session expired. Please login again.');
    }, 60 * 60 * 1000); // 1 hour
    setSessionTimeout(timeout);
  };

  const clearSessionTimeout = () => {
    if (sessionTimeout) {
      clearTimeout(sessionTimeout);
      setSessionTimeout(null);
    }
  };

  const fetchUser = async () => {
    const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));
    const maxRetries = 4;
    let attempt = 0;
    setIsLoading(true);
    while (attempt <= maxRetries) {
      try {
        const userData = await apiClient.get<User>('/auth/me');
        setUser(userData as any);
        setIsLoading(false);
        return;
      } catch (err: any) {
        console.error('Failed to fetch user:', err);
        const status = err?.status || err?.response?.status;
        // If rate limited, retry with exponential backoff
        if (status === 429 && attempt < maxRetries) {
          const delay = Math.pow(2, attempt) * 1000; // 1s,2s,4s,8s...
          attempt++;
          await sleep(delay);
          continue;
        }

        // Only logout if it's an auth error (401)
        if (status === 401 || (err?.message && err.message.includes('Invalid'))) {
          authService.logout();
          setUser(null);
        }

        setIsLoading(false);
        return;
      }
    }
  };

  const login = async (username: string, password: string) => {
    try {
      await authService.login(username, password);
      
      // After login, fetch user details
      await fetchUser();
      startSessionTimeout();
    } catch (error) {
      throw error;
    }
  };

  const logout = () => {
    clearSessionTimeout();
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
