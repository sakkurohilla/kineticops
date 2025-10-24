import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import authService from '../../services/auth/authService';

const MFASetup: React.FC = () => {
  const navigate = useNavigate();
  const [qrCode, setQrCode] = useState('');
  const [secret, setSecret] = useState('');
  const [verificationCode, setVerificationCode] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSetupLoading, setIsSetupLoading] = useState(true);

  useEffect(() => {
    setupMFA();
  }, []);

  const setupMFA = async () => {
    try {
      const response = await authService.setupMFA();
      setQrCode(response.qrCode);
      setSecret(response.secret);
    } catch (err: any) {
      setError('Failed to setup MFA. Please try again.');
    } finally {
      setIsSetupLoading(false);
    }
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!verificationCode || verificationCode.length !== 6) {
      setError('Please enter a valid 6-digit code');
      return;
    }

    setIsLoading(true);

    try {
      await authService.verifyMFA(verificationCode);
      setSuccess('MFA setup successful!');
      
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err: any) {
      setError('Invalid verification code. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  if (isSetupLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-50 flex items-center justify-center p-4">
        <div className="card max-w-md w-full text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Setting up MFA...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-50 flex items-center justify-center p-4">
      <div className="card max-w-md w-full">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-primary-700 mb-2">Setup MFA</h1>
          <p className="text-gray-600">Secure your account with two-factor authentication</p>
        </div>

        {error && (
          <div className="p-3 bg-error/10 border border-error rounded-lg mb-4">
            <p className="text-sm text-error">{error}</p>
          </div>
        )}

        {success && (
          <div className="p-3 bg-success/10 border border-success rounded-lg mb-4">
            <p className="text-sm text-green-700">{success}</p>
            <p className="text-xs text-green-600 mt-1">Redirecting to dashboard...</p>
          </div>
        )}

        <div className="space-y-6">
          <div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Step 1: Scan QR Code</h3>
            <p className="text-sm text-gray-600 mb-4">
              Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
            </p>
            
            {qrCode && (
              <div className="bg-white p-4 rounded-lg border text-center">
                <img src={qrCode} alt="MFA QR Code" className="mx-auto" />
              </div>
            )}
          </div>

          <div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Step 2: Manual Entry (Optional)</h3>
            <p className="text-sm text-gray-600 mb-2">
              Or manually enter this secret key in your authenticator app:
            </p>
            <div className="bg-gray-100 p-3 rounded font-mono text-sm break-all">
              {secret}
            </div>
          </div>

          <div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Step 3: Verify Setup</h3>
            <form onSubmit={handleVerify} className="space-y-4">
              <div>
                <label htmlFor="code" className="block text-sm font-medium text-gray-700 mb-1">
                  Enter 6-digit code from your authenticator app
                </label>
                <input
                  id="code"
                  type="text"
                  value={verificationCode}
                  onChange={(e) => setVerificationCode(e.target.value.replace(/[^0-9]/g, '').slice(0, 6))}
                  className="input text-center text-2xl tracking-widest"
                  placeholder="000000"
                  maxLength={6}
                  disabled={isLoading}
                />
              </div>

              <button
                type="submit"
                className="btn btn-primary w-full"
                disabled={isLoading || verificationCode.length !== 6}
              >
                {isLoading ? 'Verifying...' : 'Verify & Complete Setup'}
              </button>
            </form>
          </div>
        </div>

        <div className="mt-6 pt-6 border-t border-gray-200 text-center">
          <button
            onClick={() => navigate('/dashboard')}
            className="text-sm text-gray-600 hover:text-gray-800"
          >
            Skip for now (not recommended)
          </button>
        </div>
      </div>
    </div>
  );
};

export default MFASetup;
