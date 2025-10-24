import { useApp } from '../context/AppContext';

export const useAuth = () => {
  const { user, isAuthenticated, isLoading, login, logout, register } = useApp();

  return {
    user,
    isAuthenticated,
    isLoading,
    login,
    logout,
    register,
  };
};
