import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { validatePassword } from '../../utils/validation';
import authService from '../../services/auth/authService';
import PasswordStrength from '../../components/auth/PasswordStrength';

const ResetPassword: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  
  const [token, setToken] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isVerifying, setIsVerifying] = useState(true);
  const [tokenValid, setTokenValid] = useState(false);
  const [userEmail, setUserEmail] = useState('');

  // Get token from URL query parameter
  useEffect(() => {
    const tokenParam = searchParams.get('token');
    if (!tokenParam) {
      setError('Invalid reset link. Please request a new password reset.');
      setIsVerifying(false);
      return;
    }
    
    setToken(tokenParam);
    verifyToken(tokenParam);
  }, [searchParams]);

  const verifyToken = async (tokenToVerify: string) => {
    try {
      const response = await authService.verifyResetToken(tokenToVerify);
      setTokenValid(response.valid);
      setUserEmail(response.email);
    } catch (err: any) {
      setError(err.message || 'Invalid or expired reset token');
      setTokenValid(false);
    } finally {
      setIsVerifying(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    // Validate password
    const validation = validatePassword(newPassword);
    if (!validation.isValid) {
      setError(validation.error || 'Invalid password');
      return;
    }

    // Check if passwords match
    if (newPassword !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setIsLoading(true);

    try {
      const response = await authService.resetPassword(token, newPassword);
      setSuccess(response.msg || 'Password reset successful!');
      
      // Redirect to login after 2 seconds
      setTimeout(() => {
        navigate('/login');
      }, 2000);
    } catch (err: any) {
      setError(err.message || 'Failed to reset password');
    } finally {
      setIsLoading(false);
    }
  };

  // Loading state while verifying token
  if (isVerifying) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-50 flex items-center justify-center p-4">
        <div className="card max-w-md w-full text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Verifying reset token...</p>
        </div>
      </div>
    );
  }

  // Invalid token state
  if (!tokenValid) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-50 flex items-center justify-center p-4">
        <div className="card max-w-md w-full">
          <div className="text-center mb-6">
            <div className="w-16 h-16 bg-error/10 rounded-full flex items-center justify-center mx-auto mb-4">
              <span className="text-3xl">❌</span>
            </div>
            <h1 className="text-2xl font-bold text-gray-800 mb-2">Invalid Reset Link</h1>
            <p className="text-gray-600">{error || 'This password reset link is invalid or has expired.'}</p>
          </div>
          
          <div className="space-y-3">
            <Link to="/forgot-password" className="btn btn-primary w-full block text-center">
              Request New Reset Link
            </Link>
            <Link to="/login" className="btn bg-gray-200 hover:bg-gray-300 text-gray-700 w-full block text-center">
              Back to Login
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-50 flex items-center justify-center p-4">
      <div className="card max-w-md w-full">
        <div className="text-center mb-8">
          <div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-3xl">🔐</span>
          </div>
          <h1 className="text-3xl font-bold text-primary-700 mb-2">Reset Password</h1>
          <p className="text-gray-600">Enter your new password for</p>
          <p className="text-sm font-medium text-primary-600">{userEmail}</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="p-3 bg-error/10 border border-error rounded-lg">
              <p className="text-sm text-error">{error}</p>
            </div>
          )}

          {success && (
            <div className="p-3 bg-success/10 border border-success rounded-lg">
              <p className="text-sm text-green-700">{success}</p>
              <p className="text-xs text-green-600 mt-1">Redirecting to login...</p>
            </div>
          )}

          {/* New Password */}
          <div>
            <label htmlFor="newPassword" className="block text-sm font-medium text-gray-700 mb-1">
              New Password
            </label>
            <div className="relative">
              <input
                id="newPassword"
                type={showPassword ? 'text' : 'password'}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="input pr-10"
                placeholder="Enter new password"
                disabled={isLoading || success !== ''}
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
                disabled={isLoading || success !== ''}
              >
                {showPassword ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
            {newPassword && <PasswordStrength password={newPassword} />}
          </div>

          {/* Confirm Password */}
          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
              Confirm Password
            </label>
            <div className="relative">
              <input
                id="confirmPassword"
                type={showConfirmPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="input pr-10"
                placeholder="Confirm new password"
                disabled={isLoading || success !== ''}
                required
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
                disabled={isLoading || success !== ''}
              >
                {showConfirmPassword ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
            {confirmPassword && newPassword !== confirmPassword && (
              <p className="text-xs text-error mt-1">Passwords do not match</p>
            )}
            {confirmPassword && newPassword === confirmPassword && (
              <p className="text-xs text-success mt-1">✓ Passwords match</p>
            )}
          </div>

          <button
            type="submit"
            className="btn btn-primary w-full"
            disabled={isLoading || success !== '' || !newPassword || !confirmPassword}
          >
            {isLoading ? 'Resetting Password...' : 'Reset Password'}
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

export default ResetPassword;
