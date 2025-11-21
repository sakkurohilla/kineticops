import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

const LoginForm: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validation
    if (!username || !password) {
      setError('Please enter both username and password');
      return;
    }

    setIsLoading(true);

    try {
      await login(username, password);
      
      // Small delay to allow WebSocket manager to detect token change
      // and establish connection before navigating to dashboard
      await new Promise(resolve => setTimeout(resolve, 300));
      
      navigate('/dashboard');
    } catch (err: any) {
      setError(err.message || 'Invalid username or password');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="p-3 bg-error/10 border border-error rounded-lg">
          <p className="text-sm text-error">{error}</p>
        </div>
      )}

      <div>
        <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
          Username
        </label>
        <input
          id="username"
          name="username"
          autoComplete="username"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          className="input"
          placeholder="Enter your username"
          disabled={isLoading}
        />
      </div>

      <div>
        <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
          Password
        </label>
        <input
          id="password"
          name="password"
          autoComplete="current-password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="input"
          placeholder="Enter your password"
          disabled={isLoading}
        />
      </div>

      <div className="flex items-center justify-between">
        <label className="flex items-center">
          <input type="checkbox" className="mr-2" />
          <span className="text-sm text-gray-600">Remember me</span>
        </label>
        <a href="/forgot-password" className="text-sm text-primary-600 hover:text-primary-700">
          Forgot password?
        </a>
      </div>

      <button
        type="submit"
        className="btn btn-primary w-full"
        disabled={isLoading}
      >
        {isLoading ? 'Signing in...' : 'Sign In'}
      </button>
    </form>
  );
};

export default LoginForm;
