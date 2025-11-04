import React, { useState, useEffect } from 'react';
import { X, Loader, CheckCircle, AlertCircle, Server } from 'lucide-react';
import Button from '../common/Button';
import Input from '../common/Input';
import hostService, { CreateHostRequest } from '../../services/api/hostService';

interface AddHostFormProps {
  onClose: () => void;
  onSuccess: () => void;
  mode?: 'create' | 'edit';
  hostId?: number;
  initialData?: Partial<CreateHostRequest & { ssh_key?: string }>;
}

interface AgentSetupResponse {
  host_id: number;
  agent_id: number;
  token: string;
  setup_method: string;
  install_script?: string;
  status: string;
  message: string;
}

const AddHostForm: React.FC<AddHostFormProps> = ({ onClose, onSuccess, mode = 'create', hostId, initialData }) => {
  const [formData, setFormData] = useState<CreateHostRequest & { ssh_key?: string; auth_method?: 'password' | 'key'; setup_method?: 'automatic' | 'manual' | 'none' }>({
    hostname: '',
    ip: '',
    ssh_user: 'root',
    ssh_password: '',
    ssh_key: '',
    ssh_port: 22,
    os: 'linux',
    group: 'default',
    tags: '',
    description: '',
    auth_method: 'password',
    setup_method: 'automatic',
  });

  const [testing, setTesting] = useState(false);
  const [creating, setCreating] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [error, setError] = useState<string>('');
  const [setupResult, setSetupResult] = useState<AgentSetupResponse | null>(null);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: name === 'ssh_port' ? parseInt(value) || 22 : value,
    }));
    setTestResult(null); // Reset test result when form changes
  };

  // populate initialData when in edit mode
  useEffect(() => {
    if (mode === 'edit' && initialData) {
      setFormData((prev) => ({ ...prev, ...(initialData as any) }));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleTestConnection = async () => {
    if (!formData.ip || !formData.ssh_user) {
      setError('Please fill in IP and Username to test connection');
      return;
    }
    
    if (formData.auth_method === 'password' && !formData.ssh_password) {
      setError('Please provide SSH password');
      return;
    }
    
    if (formData.auth_method === 'key' && !formData.ssh_key) {
      setError('Please provide SSH private key');
      return;
    }

    setTesting(true);
    setError('');
    setTestResult(null);

    try {
      const result = await hostService.testSSHConnection({
        ip: formData.ip,
        port: formData.ssh_port,
        username: formData.ssh_user,
        password: formData.auth_method === 'password' ? formData.ssh_password : '',
        private_key: formData.auth_method === 'key' ? formData.ssh_key : '',
      });

      if (result.success) {
        setTestResult({ success: true, message: 'SSH connection successful!' });
      } else {
        setTestResult({ success: false, message: result.error || 'Connection failed' });
      }
    } catch (err: any) {
      setTestResult({ success: false, message: err.message || 'Connection test failed' });
    } finally {
      setTesting(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();


    // For create mode require a successful SSH test. Edit mode doesn't need SSH test.
    if (mode === 'create' && (!testResult || !testResult.success)) {
      setError('Please test the SSH connection first');
      return;
    }

    if (!formData.hostname) {
      setError('Hostname is required');
      return;
    }

    setCreating(true);
    setError('');

    try {
      if (mode === 'edit' && hostId) {
        // Only send basic fields for edit - no SSH credentials
        const payload = {
          hostname: formData.hostname,
          ip: formData.ip,
          os: formData.os,
          group: formData.group,
          tags: formData.tags,
          description: formData.description
        };
        await hostService.updateHost(hostId, payload);
        onSuccess();
        onClose();
      } else {
        // Create host with agent setup
        const setupPayload = {
          setup_method: formData.setup_method || 'automatic',
          hostname: formData.hostname,
          ip: formData.ip,
          username: formData.ssh_user,
          password: formData.auth_method === 'password' ? formData.ssh_password : '',
          ssh_key: formData.auth_method === 'key' ? formData.ssh_key : '',
          port: formData.ssh_port,
        };

        const response = await hostService.createHostWithAgent(setupPayload);
        setSetupResult(response);
        
        if (response.setup_method === 'automatic') {
          // Automatic setup - show success and close
          setTimeout(() => {
            onSuccess();
            onClose();
          }, 2000);
        }
      }
    } catch (err: any) {
      setError(err.message || 'Failed to create host');
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="sticky top-0 bg-gradient-to-r from-blue-600 to-purple-600 text-white p-6 rounded-t-2xl flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-white bg-opacity-20 rounded-lg flex items-center justify-center">
              <Server className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-2xl font-bold">{mode === 'edit' ? 'Edit Host' : 'Add New Host'}</h2>
              <p className="text-blue-100 text-sm">
                {mode === 'edit' ? 'Update host information and settings' : 'Configure SSH connection to monitor your server'}
              </p>
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
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* Error Message */}
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Test Result */}
          {testResult && (
            <div
              className={`border rounded-lg p-4 flex items-start gap-3 ${
                testResult.success
                  ? 'bg-green-50 border-green-200'
                  : 'bg-red-50 border-red-200'
              }`}
            >
              {testResult.success ? (
                <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
              ) : (
                <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
              )}
              <p
                className={`text-sm ${
                  testResult.success ? 'text-green-800' : 'text-red-800'
                }`}
              >
                {testResult.message}
              </p>
            </div>
          )}

          {/* Setup Method Selection */}
          {mode === 'create' && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                <div className="w-2 h-2 bg-green-600 rounded-full"></div>
                Setup Method
              </h3>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div 
                  className={`p-4 border-2 rounded-lg cursor-pointer transition-all ${
                    formData.setup_method === 'automatic' 
                      ? 'border-blue-500 bg-blue-50' 
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                  onClick={() => setFormData(prev => ({ ...prev, setup_method: 'automatic' }))}
                >
                  <div className="flex items-center gap-3">
                    <input
                      type="radio"
                      name="setup_method"
                      value="automatic"
                      checked={formData.setup_method === 'automatic'}
                      onChange={handleChange}
                      className="text-blue-600"
                    />
                    <div>
                      <h4 className="font-semibold text-gray-900">⚡ Automatic</h4>
                      <p className="text-sm text-gray-600">SSH auto-install (just enter credentials)</p>
                    </div>
                  </div>
                </div>
                
                <div 
                  className={`p-4 border-2 rounded-lg cursor-pointer transition-all ${
                    formData.setup_method === 'manual' 
                      ? 'border-blue-500 bg-blue-50' 
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                  onClick={() => setFormData(prev => ({ ...prev, setup_method: 'manual' }))}
                >
                  <div className="flex items-center gap-3">
                    <input
                      type="radio"
                      name="setup_method"
                      value="manual"
                      checked={formData.setup_method === 'manual'}
                      onChange={handleChange}
                      className="text-blue-600"
                    />
                    <div>
                      <h4 className="font-semibold text-gray-900">📋 Manual</h4>
                      <p className="text-sm text-gray-600">Download script and run manually</p>
                    </div>
                  </div>
                </div>
                
                <div 
                  className={`p-4 border-2 rounded-lg cursor-pointer transition-all ${
                    formData.setup_method === 'none' 
                      ? 'border-blue-500 bg-blue-50' 
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                  onClick={() => setFormData(prev => ({ ...prev, setup_method: 'none' }))}
                >
                  <div className="flex items-center gap-3">
                    <input
                      type="radio"
                      name="setup_method"
                      value="none"
                      checked={formData.setup_method === 'none'}
                      onChange={handleChange}
                      className="text-blue-600"
                    />
                    <div>
                      <h4 className="font-semibold text-gray-900">🚫 No Agent</h4>
                      <p className="text-sm text-gray-600">Create host without monitoring agent</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Host Information */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
              <div className="w-2 h-2 bg-blue-600 rounded-full"></div>
              Host Information
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
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
                disabled={mode === 'edit'}
              />

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Operating System
                </label>
                <select
                  name="os"
                  value={formData.os}
                  onChange={handleChange}
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="linux">Linux</option>
                  <option value="ubuntu">Ubuntu</option>
                  <option value="centos">CentOS</option>
                  <option value="debian">Debian</option>
                  <option value="redhat">Red Hat</option>
                </select>
              </div>

              <Input
                label="Group"
                name="group"
                value={formData.group}
                onChange={handleChange}
                placeholder="production"
              />
            </div>

            <Input
              label="Tags (comma-separated)"
              name="tags"
              value={formData.tags}
              onChange={handleChange}
              placeholder="web, production, critical"
            />

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Description
              </label>
              <textarea
                name="description"
                value={formData.description}
                onChange={handleChange}
                placeholder="Brief description of this host..."
                rows={3}
                className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              />
            </div>
          </div>

          {/* SSH Configuration - Only for create mode */}
          {mode === 'create' && formData.setup_method !== undefined && (
            <div className="space-y-4 pt-6 border-t border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                <div className="w-2 h-2 bg-purple-600 rounded-full"></div>
                SSH Configuration
              </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input
                label="SSH Username"
                name="ssh_user"
                value={formData.ssh_user}
                onChange={handleChange}
                placeholder="root"
                required
              />

              <Input
                label="SSH Port"
                name="ssh_port"
                type="number"
                value={formData.ssh_port}
                onChange={handleChange}
                placeholder="22"
                required
              />
            </div>

            {/* Authentication Method */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-3">
                Authentication Method
              </label>
              <div className="flex gap-4">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="auth_method"
                    value="password"
                    checked={formData.auth_method === 'password'}
                    onChange={handleChange}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">Password</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="auth_method"
                    value="key"
                    checked={formData.auth_method === 'key'}
                    onChange={handleChange}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">SSH Key</span>
                </label>
              </div>
            </div>

            {/* Password or SSH Key based on selection */}
            {formData.auth_method === 'password' ? (
              <Input
                label="SSH Password"
                name="ssh_password"
                type="password"
                value={formData.ssh_password}
                onChange={handleChange}
                placeholder="Enter SSH password"
                required
              />
            ) : (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  SSH Private Key
                </label>
                <textarea
                  name="ssh_key"
                  value={formData.ssh_key}
                  onChange={handleChange}
                  placeholder="-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----"
                  rows={6}
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm resize-none"
                  required
                />
                <p className="text-xs text-gray-500 mt-1">
                  <strong>Private SSH key required</strong> (not public key). Generate with: <code className="bg-gray-100 px-1 rounded">ssh-keygen -t ed25519</code><br/>
                  Copy content from: <code className="bg-gray-100 px-1 rounded">~/.ssh/id_ed25519</code> (private key file)
                </p>
              </div>
            )}

              {/* Test Connection Button */}
              <Button
                type="button"
                variant="outline"
                fullWidth
                onClick={handleTestConnection}
                disabled={testing || !formData.ip || !formData.ssh_user || (formData.auth_method === 'password' && !formData.ssh_password) || (formData.auth_method === 'key' && !formData.ssh_key)}
              >
                {testing ? (
                  <>
                    <Loader className="w-4 h-4 animate-spin" />
                    Testing Connection...
                  </>
                ) : (
                  <>
                    <CheckCircle className="w-4 h-4" />
                    Test SSH Connection
                  </>
                )}
              </Button>
            </div>
          )}

          {/* Setup Result */}
          {setupResult && (
            <div className="space-y-4 pt-6 border-t border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                <div className="w-2 h-2 bg-green-600 rounded-full"></div>
                Setup Result
              </h3>
              
              <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                <div className="flex items-start gap-3">
                  <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <p className="text-sm font-medium text-green-800">{setupResult.message}</p>
                    <p className="text-xs text-green-600 mt-1">Host ID: {setupResult.host_id} | Agent ID: {setupResult.agent_id}</p>
                  </div>
                </div>
              </div>
              
              {setupResult.setup_method === 'manual' && setupResult.install_script && (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <h4 className="font-medium text-gray-900">Installation Script</h4>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        navigator.clipboard.writeText(setupResult.install_script!);
                        alert('Script copied to clipboard!');
                      }}
                    >
                      Copy Script
                    </Button>
                  </div>
                  
                  <div className="bg-gray-900 text-green-400 p-4 rounded-lg font-mono text-sm max-h-64 overflow-y-auto">
                    <pre>{setupResult.install_script}</pre>
                  </div>
                  
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <h5 className="font-medium text-blue-900 mb-2">Manual Installation Steps:</h5>
                    <ol className="text-sm text-blue-800 space-y-1 list-decimal list-inside">
                      <li>Copy the script above</li>
                      <li>SSH into your host: <code className="bg-blue-100 px-1 rounded">ssh {formData.ssh_user}@{formData.ip}</code></li>
                      <li>Save script: <code className="bg-blue-100 px-1 rounded">nano ~/.kineticops-install.sh</code></li>
                      <li>Run script: <code className="bg-blue-100 px-1 rounded">bash ~/.kineticops-install.sh</code></li>
                      <li>Agent will start automatically and appear online in dashboard</li>
                    </ol>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex gap-3 pt-6 border-t border-gray-200">
            {setupResult && setupResult.setup_method === 'manual' ? (
              <Button type="button" variant="primary" fullWidth onClick={() => { onSuccess(); onClose(); }}>
                Done - Go to Dashboard
              </Button>
            ) : (
              <>
                <Button type="button" variant="outline" fullWidth onClick={onClose}>
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  fullWidth
                  disabled={creating || (mode === 'create' && formData.setup_method !== 'none' && !testResult?.success)}
                >
                  {creating ? (
                    <>
                      <Loader className="w-4 h-4 animate-spin" />
                      {mode === 'edit' ? 'Saving Changes...' : formData.setup_method === 'automatic' ? 'Installing Agent...' : 'Creating Host...'}
                    </>
                  ) : (
                    <>
                      <Server className="w-4 h-4" />
                      {mode === 'edit' ? 'Save Changes' : formData.setup_method === 'automatic' ? 'Install Agent' : 'Create Host'}
                    </>
                  )}
                </Button>
              </>
            )}
          </div>
        </form>
      </div>
    </div>
  );
};

export default AddHostForm;