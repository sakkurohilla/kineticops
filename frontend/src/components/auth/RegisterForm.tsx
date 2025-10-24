import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { validateEmail, validatePassword, validateName, validateConfirmPassword } from '../../utils/validation';
import PasswordStrength from './PasswordStrength';

const RegisterForm: React.FC = () => {
  const navigate = useNavigate();
  const { register } = useAuth();
  
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const validateField = (field: string, value: string) => {
    let validation;
    
    switch (field) {
      case 'username':
        validation = validateName(value);
        break;
      case 'email':
        validation = validateEmail(value);
        break;
      case 'password':
        validation = validatePassword(value);
        break;
      case 'confirmPassword':
        validation = validateConfirmPassword(password, value);
        break;
      default:
        return;
    }

    setFieldErrors(prev => ({
      ...prev,
      [field]: validation.isValid ? '' : (validation.error || ''),
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    // Validate all fields
    const usernameValidation = validateName(username);
    const emailValidation = validateEmail(email);
    const passwordValidation = validatePassword(password);
    const confirmPasswordValidation = validateConfirmPassword(password, confirmPassword);

    const errors: Record<string, string> = {};
    
    if (!usernameValidation.isValid) errors.username = usernameValidation.error || '';
    if (!emailValidation.isValid) errors.email = emailValidation.error || '';
    if (!passwordValidation.isValid) errors.password = passwordValidation.error || '';
    if (!confirmPasswordValidation.isValid) errors.confirmPassword = confirmPasswordValidation.error || '';

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      setError('Please fix the errors above');
      return;
    }

    setIsLoading(true);

    try {
      const message = await register(username, email, password);
      setSuccess(message);
      
      // Redirect to login after 2 seconds
      setTimeout(() => {
        navigate('/login');
      }, 2000);
    } catch (err: any) {
      console.error('Registration error:', err);
      setError(err.message || 'Registration failed. Please try again.');
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

      {success && (
        <div className="p-3 bg-success/10 border border-success rounded-lg">
          <p className="text-sm text-green-700">{success}</p>
          <p className="text-xs text-green-600 mt-1">Redirecting to login...</p>
        </div>
      )}

      <div>
        <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
          Username
        </label>
        <input
          id="username"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          onBlur={(e) => validateField('username', e.target.value)}
          className={`input ${fieldErrors.username ? 'border-error' : ''}`}
          placeholder="Choose a username"
          disabled={isLoading}
        />
        {fieldErrors.username && (
          <p className="text-xs text-error mt-1">{fieldErrors.username}</p>
        )}
      </div>

      <div>
        <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
          Email
        </label>
        <input
          id="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          onBlur={(e) => validateField('email', e.target.value)}
          className={`input ${fieldErrors.email ? 'border-error' : ''}`}
          placeholder="you@example.com"
          disabled={isLoading}
        />
        {fieldErrors.email && (
          <p className="text-xs text-error mt-1">{fieldErrors.email}</p>
        )}
      </div>

      <div>
        <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
          Password
        </label>
        <div className="relative">
          <input
            id="password"
            type={showPassword ? 'text' : 'password'}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            onBlur={(e) => validateField('password', e.target.value)}
            className={`input pr-10 ${fieldErrors.password ? 'border-error' : ''}`}
            placeholder="Create a strong password"
            disabled={isLoading}
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
          >
            {showPassword ? (
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
              </svg>
            ) : (
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
            )}
          </button>
        </div>
        {fieldErrors.password && (
          <p className="text-xs text-error mt-1">{fieldErrors.password}</p>
        )}
        <PasswordStrength password={password} />
      </div>

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
            onBlur={(e) => validateField('confirmPassword', e.target.value)}
            className={`input pr-10 ${fieldErrors.confirmPassword ? 'border-error' : ''}`}
            placeholder="Re-enter your password"
            disabled={isLoading}
          />
          <button
            type="button"
            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
          >
            {showConfirmPassword ? (
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
              </svg>
            ) : (
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
            )}
          </button>
        </div>
        {fieldErrors.confirmPassword && (
          <p className="text-xs text-error mt-1">{fieldErrors.confirmPassword}</p>
        )}
      </div>

      <button
        type="submit"
        className="btn btn-secondary w-full"
        disabled={isLoading}
      >
        {isLoading ? 'Creating account...' : 'Create Account'}
      </button>
    </form>
  );
};

export default RegisterForm;
