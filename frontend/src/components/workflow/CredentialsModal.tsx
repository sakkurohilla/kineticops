import React, { useState } from 'react';
import { X, Key, Loader, Shield } from 'lucide-react';
import Button from '../common/Button';
import Input from '../common/Input';

interface CredentialsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (credentials: { username: string; password?: string; ssh_key?: string }) => Promise<void>;
  hostName: string;
  hostIP: string;
  loading?: boolean;
}

const CredentialsModal: React.FC<CredentialsModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  hostName,
  hostIP,
  loading = false
}) => {
  const [credentials, setCredentials] = useState({
    username: 'root',
    password: '',
    ssh_key: '',
    auth_method: 'password' as 'password' | 'key'
  });
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!credentials.username) {
      setError('Username is required');
      return;
    }

    if (credentials.auth_method === 'password' && !credentials.password) {
      setError('Password is required');
      return;
    }

    if (credentials.auth_method === 'key' && !credentials.ssh_key) {
      setError('SSH key is required');
      return;
    }

    try {
      await onSubmit({
        username: credentials.username,
        password: credentials.auth_method === 'password' ? credentials.password : undefined,
        ssh_key: credentials.auth_method === 'key' ? credentials.ssh_key : undefined
      });
    } catch (err: any) {
      setError(err.message || 'Authentication failed');
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full">
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white p-6 rounded-t-2xl flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-white bg-opacity-20 rounded-lg flex items-center justify-center">
              <Shield className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-xl font-bold">Secure Connection</h2>
              <p className="text-blue-100 text-sm">Enter credentials for workflow session</p>
            </div>
          </div>
          <button onClick={onClose} className="text-white hover:bg-white hover:bg-opacity-20 rounded-lg p-2 transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-center gap-2 mb-2">
              <div className="w-2 h-2 bg-blue-600 rounded-full"></div>
              <span className="text-sm font-medium text-blue-900">Target Host</span>
            </div>
            <p className="text-sm text-blue-800 font-mono">{hostName} ({hostIP})</p>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <Input
            label="Username"
            name="username"
            value={credentials.username}
            onChange={(e) => setCredentials(prev => ({ ...prev, username: e.target.value }))}
            placeholder="root"
            required
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-3">Authentication Method</label>
            <div className="flex gap-4">
              <label className="flex items-center">
                <input
                  type="radio"
                  name="auth_method"
                  value="password"
                  checked={credentials.auth_method === 'password'}
                  onChange={(e) => setCredentials(prev => ({ ...prev, auth_method: e.target.value as 'password' | 'key' }))}
                  className="mr-2"
                />
                <span className="text-sm text-gray-700">Password</span>
              </label>
              <label className="flex items-center">
                <input
                  type="radio"
                  name="auth_method"
                  value="key"
                  checked={credentials.auth_method === 'key'}
                  onChange={(e) => setCredentials(prev => ({ ...prev, auth_method: e.target.value as 'password' | 'key' }))}
                  className="mr-2"
                />
                <span className="text-sm text-gray-700">SSH Key</span>
              </label>
            </div>
          </div>

          {credentials.auth_method === 'password' ? (
            <Input
              label="Password"
              name="password"
              type="password"
              value={credentials.password}
              onChange={(e) => setCredentials(prev => ({ ...prev, password: e.target.value }))}
              placeholder="Enter password"
              required
            />
          ) : (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">SSH Private Key</label>
              <textarea
                name="ssh_key"
                value={credentials.ssh_key}
                onChange={(e) => setCredentials(prev => ({ ...prev, ssh_key: e.target.value }))}
                placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                rows={6}
                className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm resize-none"
                required
              />
            </div>
          )}

          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
            <div className="flex items-start gap-2">
              <Key className="w-4 h-4 text-yellow-600 mt-0.5 flex-shrink-0" />
              <div className="text-xs text-yellow-800">
                <p className="font-medium mb-1">Security Notice</p>
                <p>Credentials are used only for this session (1 hour) and are never stored permanently.</p>
              </div>
            </div>
          </div>

          <div className="flex gap-3 pt-4">
            <Button type="button" variant="outline" fullWidth onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" fullWidth disabled={loading}>
              {loading ? (
                <>
                  <Loader className="w-4 h-4 animate-spin" />
                  Connecting...
                </>
              ) : (
                <>
                  <Shield className="w-4 h-4" />
                  Connect Securely
                </>
              )}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CredentialsModal;