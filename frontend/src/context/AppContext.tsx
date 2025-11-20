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
  const [lastActivity, setLastActivity] = useState<number>(Date.now());

  useEffect(() => {
    // Check if user is already logged in
    const token = authService.getToken();
    if (token) {
      fetchUser();
      startSessionTimeout();
      initializeActivityTracking();
    } else {
      setIsLoading(false);
    }

    return () => {
      clearSessionTimeout();
      removeActivityTracking();
    };
  }, []);

  // Track user activity (mouse, keyboard, scroll)
  const initializeActivityTracking = () => {
    const updateActivity = () => {
      setLastActivity(Date.now());
    };
    
    window.addEventListener('mousemove', updateActivity);
    window.addEventListener('keydown', updateActivity);
    window.addEventListener('scroll', updateActivity);
    window.addEventListener('click', updateActivity);
  };

  const removeActivityTracking = () => {
    const updateActivity = () => setLastActivity(Date.now());
    window.removeEventListener('mousemove', updateActivity);
    window.removeEventListener('keydown', updateActivity);
    window.removeEventListener('scroll', updateActivity);
    window.removeEventListener('click', updateActivity);
  };

  // Check for inactivity every minute
  useEffect(() => {
    if (!user) return;
    
    const checkInactivity = setInterval(() => {
      const inactiveTime = Date.now() - lastActivity;
      const INACTIVITY_LIMIT = 10 * 60 * 1000; // 10 minutes
      
      if (inactiveTime > INACTIVITY_LIMIT) {
        logout();
        alert('Session expired due to inactivity. Please login again.');
      }
    }, 60 * 1000); // Check every minute
    
    return () => clearInterval(checkInactivity);
  }, [user, lastActivity]);

  const startSessionTimeout = () => {
    clearSessionTimeout();
    // Session expires after 10 minutes of inactivity
    const timeout = setTimeout(() => {
      logout();
      alert('Session expired. Please login again.');
    }, 10 * 60 * 1000); // 10 minutes
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

        // Only logout if it's an auth error (401), not on network errors
        if (status === 401 || (err?.message && err.message.includes('Invalid'))) {
          authService.logout();
          setUser(null);
        } else if (err?.isNetworkError) {
          // Don't logout on network errors, just set loading to false
          console.warn('[AppContext] Network error during fetchUser, keeping session');
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
