import React, { useState } from 'react';
import { X, Server, Copy, CheckCircle, AlertCircle } from 'lucide-react';
import Button from '../common/Button';
import Input from '../common/Input';
import apiClient from '../../services/api/client';

interface SimpleAddHostFormProps {
  onClose: () => void;
  onSuccess: () => void;
}

const SimpleAddHostForm: React.FC<SimpleAddHostFormProps> = ({ onClose, onSuccess }) => {
  const [formData, setFormData] = useState({
    hostname: '',
    ip: '',
    description: '',
  });
  const [targetOS, setTargetOS] = useState('ubuntu');

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [installCommand, setInstallCommand] = useState('');

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    setError('');
  };

  const handleOSChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setTargetOS(e.target.value);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.hostname) {
      setError('Hostname is required');
      return;
    }
    
    if (!formData.ip) {
      setError('IP Address is required');
      return;
    }

    setLoading(true);
    setError('');

    try {
      // Generate installation token from your backend
  const tokenData = await apiClient.post('/install/token', { target_os: targetOS }) as any;
      setSuccess(true);
      setInstallCommand(tokenData.command);
      
      // Auto-close after 15 seconds
      setTimeout(() => {
        onSuccess();
        onClose();
      }, 15000);

    } catch (err: any) {
      setError(err.message || 'Failed to generate installation token');
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    // You could add a toast notification here
  };

  if (success) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
        <div className="bg-white rounded-2xl shadow-2xl max-w-2xl w-full">
          {/* Header */}
          <div className="bg-gradient-to-r from-green-600 to-blue-600 text-white p-6 rounded-t-2xl flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 bg-white bg-opacity-20 rounded-lg flex items-center justify-center">
                <CheckCircle className="w-6 h-6" />
              </div>
              <div>
                <h2 className="text-2xl font-bold">Host Added Successfully!</h2>
                <p className="text-green-100 text-sm">Now install the monitoring agent</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="text-white hover:bg-white hover:bg-opacity-20 rounded-lg p-2 transition-colors"
            >
              <X className="w-6 h-6" />
            </button>
          </div>

          {/* Content */}
          <div className="p-6 space-y-6">
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h3 className="font-semibold text-blue-900 mb-2">Step 1: Install KineticOps Agent</h3>
              <p className="text-sm text-blue-800 mb-3">
                Run this command on your target server <strong>{formData.hostname || 'your-server'}</strong>:
              </p>
              
              <div className="bg-gray-900 text-green-400 p-3 rounded font-mono text-sm flex items-center justify-between">
                <code className="flex-1 break-all">{installCommand}</code>
                <button
                  onClick={() => copyToClipboard(installCommand)}
                  className="ml-3 p-2 hover:bg-gray-800 rounded transition-colors"
                  title="Copy to clipboard"
                >
                  <Copy className="w-4 h-4" />
                </button>
              </div>
            </div>

            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <h3 className="font-semibold text-yellow-900 mb-2">Step 2: Automatic Discovery</h3>
              <p className="text-sm text-yellow-800">
                The installation token is embedded in the command and will automatically associate the host with your account. No additional configuration needed!
              </p>
            </div>

            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <h3 className="font-semibold text-green-900 mb-2">What happens next?</h3>
              <ul className="text-sm text-green-800 space-y-1">
                <li>• Agent installs and automatically registers with your account</li>
                <li>• Host appears in your dashboard within 30 seconds</li>
                <li>• System metrics (CPU, memory, disk, network) start collecting immediately</li>
                <li>• No manual host creation or configuration required</li>
                <li>• Token expires in 24 hours for security</li>
              </ul>
            </div>

            <div className="flex gap-3">
              <Button variant="outline" fullWidth onClick={onClose}>
                Close
              </Button>
              <Button variant="primary" fullWidth onClick={() => { onSuccess(); onClose(); }}>
                Go to Dashboard
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-2xl max-w-lg w-full">
        {/* Header */}
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white p-6 rounded-t-2xl flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-white bg-opacity-20 rounded-lg flex items-center justify-center">
              <Server className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-2xl font-bold">Add New Host</h2>
              <p className="text-blue-100 text-sm">Simple host registration</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-white hover:bg-white hover:bg-opacity-20 rounded-lg p-2 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Error Message */}
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <Input
            label="Hostname"
            name="hostname"
            value={formData.hostname}
            onChange={handleChange}
            placeholder="web-server-01"
            required
          />

          <Input
            label="IP Address"
            name="ip"
            value={formData.ip}
            onChange={handleChange}
            placeholder="192.168.1.100"
            required
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Description (Optional)
            </label>
            <textarea
              name="description"
              value={formData.description}
              onChange={handleChange}
              placeholder="Web server, Database, etc."
              rows={3}
              className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Target OS</label>
            <select
              value={targetOS}
              onChange={handleOSChange}
              className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="ubuntu">Ubuntu (recommended)</option>
              <option value="centos">CentOS / RHEL</option>
              <option value="other">Other</option>
            </select>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h4 className="font-medium text-blue-900 mb-2">How it works:</h4>
            <p className="text-sm text-blue-800">
              You'll get a secure installation command that automatically associates the agent with your account. 
              No manual host creation needed - everything is discovered automatically!
            </p>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3 pt-4">
            <Button type="button" variant="outline" fullWidth onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              fullWidth
              disabled={loading}
            >
              {loading ? 'Generating Installation Command...' : 'Generate Installation Command'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default SimpleAddHostForm;