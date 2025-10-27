import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { validateEmail } from '../../utils/validation';
import authService from '../../services/auth/authService';

const ForgotPassword: React.FC = () => {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [resetToken, setResetToken] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setResetToken('');

    const validation = validateEmail(email);
    if (!validation.isValid) {
      setError(validation.error || 'Invalid email');
      return;
    }

    setIsLoading(true);

    try {
      const response = await authService.forgotPassword(email);
      setSuccess(response.msg || 'Password reset link sent to your email');
      
      // In development, backend returns token - display it
      if (response.token) {
        setResetToken(response.token);
      }
      
      setEmail('');
    } catch (err: any) {
      setError(err.message || 'Failed to send reset link');
    } finally {
      setIsLoading(false);
    }
  };

  const getResetLink = () => {
    const baseUrl = window.location.origin;
    return `${baseUrl}/reset-password?token=${resetToken}`;
  };

  const copyResetLink = () => {
    navigator.clipboard.writeText(getResetLink());
    alert('Reset link copied to clipboard!');
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-50 flex items-center justify-center p-4">
      <div className="card max-w-md w-full">
        <div className="text-center mb-8">
          <div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-3xl">🔑</span>
          </div>
          <h1 className="text-3xl font-bold text-primary-700 mb-2">Forgot Password?</h1>
          <p className="text-gray-600">Enter your email to reset your password</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="p-3 bg-error/10 border border-error rounded-lg">
              <p className="text-sm text-error">{error}</p>
            </div>
          )}

          {success && (
            <div className="p-3 bg-success/10 border border-success rounded-lg">
              <p className="text-sm text-green-700 font-medium">{success}</p>
              
              {/* Development Only: Show reset link */}
              {resetToken && (
                <div className="mt-3 p-3 bg-yellow-50 border border-yellow-200 rounded">
                  <p className="text-xs font-semibold text-yellow-800 mb-2">
                    🚧 DEVELOPMENT MODE - Reset Link:
  </p>
                  <div className="bg-white p-2 rounded border border-yellow-300 mb-2">
                    <p className="text-xs text-gray-700 break-all font-mono">
                      {getResetLink()}
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={copyResetLink}
                    className="text-xs bg-yellow-600 hover:bg-yellow-700 text-white px-3 py-1 rounded"
                  >
                    Copy Link
                  </button>
                  <p className="text-xs text-yellow-700 mt-2">
                    Click the link above or copy it to reset your password.
                  </p>
                </div>
              )}
            </div>
          )}

          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              Email Address
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="input"
              placeholder="you@example.com"
              disabled={isLoading}
              required
            />
          </div>

          <button
            type="submit"
            className="btn btn-primary w-full"
            disabled={isLoading}
          >
            {isLoading ? 'Sending...' : 'Send Reset Link'}
          </button>
        </form>

        <div className="mt-6 pt-6 border-t border-gray-200 text-center">
          <p className="text-sm text-gray-600">
            Remember your password?{' '}
            <Link to="/login" className="text-primary-600 hover:text-primary-700 font-medium">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
};

export default ForgotPassword;