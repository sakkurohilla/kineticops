import React, { useState } from 'react';
import { Play, Square, RotateCcw, CheckCircle, XCircle, Eye, Loader, Terminal } from 'lucide-react';
import Button from '../common/Button';

interface ServiceControlPanelProps {
  service: {
    id: number;
    name: string;
    status: string;
    port?: number;
  };
  onAction: (action: 'start' | 'stop' | 'restart' | 'enable' | 'disable') => Promise<any>;
  onClose: () => void;
  loading?: boolean;
}

const ServiceControlPanel: React.FC<ServiceControlPanelProps> = ({
  service,
  onAction,
  onClose,
  loading = false
}) => {
  const [actionLoading, setActionLoading] = useState<string>('');
  const [output, setOutput] = useState<string>('');
  const [showOutput, setShowOutput] = useState(false);

  const handleAction = async (action: 'start' | 'stop' | 'restart' | 'enable' | 'disable') => {
    setActionLoading(action);
    setOutput('');
    setShowOutput(true);
    
    try {
      const result = await onAction(action);
      setOutput(result.output || `${action} command executed successfully`);
    } catch (error: any) {
      setOutput(`Error: ${error.message || 'Command failed'}`);
    } finally {
      setActionLoading('');
    }
  };

  const actions = [
    { 
      key: 'start', 
      label: 'Start', 
      icon: <Play className="w-4 h-4" />, 
      color: 'bg-green-600 hover:bg-green-700',
      disabled: service.status === 'running'
    },
    { 
      key: 'stop', 
      label: 'Stop', 
      icon: <Square className="w-4 h-4" />, 
      color: 'bg-red-600 hover:bg-red-700',
      disabled: service.status === 'stopped'
    },
    { 
      key: 'restart', 
      label: 'Restart', 
      icon: <RotateCcw className="w-4 h-4" />, 
      color: 'bg-yellow-600 hover:bg-yellow-700',
      disabled: false
    },
    { 
      key: 'enable', 
      label: 'Enable', 
      icon: <CheckCircle className="w-4 h-4" />, 
      color: 'bg-blue-600 hover:bg-blue-700',
      disabled: false
    },
    { 
      key: 'disable', 
      label: 'Disable', 
      icon: <XCircle className="w-4 h-4" />, 
      color: 'bg-gray-600 hover:bg-gray-700',
      disabled: false
    },
    { 
      key: 'logs', 
      label: 'View Logs', 
      icon: <Eye className="w-4 h-4" />, 
      color: 'bg-purple-600 hover:bg-purple-700',
      disabled: false,
      onClick: () => setShowOutput(!showOutput)
    }
  ];

  return (
    <div className="fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="bg-gradient-to-r from-slate-800 to-slate-900 text-white p-6 rounded-t-2xl">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-bold mb-2">Service Control</h2>
              <div className="flex items-center gap-4">
                <span className="text-lg font-mono">{service.name}</span>
                <div className={`px-3 py-1 rounded-full text-sm font-medium ${
                  service.status === 'running' 
                    ? 'bg-green-500 bg-opacity-20 text-green-300' 
                    : 'bg-red-500 bg-opacity-20 text-red-300'
                }`}>
                  {service.status}
                </div>
                {service.port && (
                  <span className="text-sm text-gray-300">Port: {service.port}</span>
                )}
              </div>
            </div>
            <button
              onClick={onClose}
              className="text-white hover:bg-white hover:bg-opacity-20 rounded-lg p-2 transition-colors"
            >
              <XCircle className="w-6 h-6" />
            </button>
          </div>
        </div>

        <div className="p-6">
          {/* Action Buttons */}
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
            {actions.map((action) => (
              <Button
                key={action.key}
                onClick={action.onClick || (() => handleAction(action.key as any))}
                disabled={action.disabled || loading || actionLoading !== ''}
                className={`${action.color} text-white font-semibold py-3 px-4 rounded-lg transition-all duration-200 flex items-center justify-center gap-2 ${
                  action.disabled ? 'opacity-50 cursor-not-allowed' : 'hover:scale-105'
                }`}
              >
                {actionLoading === action.key ? (
                  <Loader className="w-4 h-4 animate-spin" />
                ) : (
                  action.icon
                )}
                {action.label}
              </Button>
            ))}
          </div>

          {/* Command Output */}
          {showOutput && (
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Terminal className="w-5 h-5 text-gray-600" />
                <h3 className="text-lg font-semibold text-gray-900">Command Output</h3>
              </div>
              
              <div className="bg-gray-900 text-green-400 p-4 rounded-lg font-mono text-sm max-h-64 overflow-y-auto">
                {actionLoading ? (
                  <div className="flex items-center gap-2">
                    <Loader className="w-4 h-4 animate-spin" />
                    <span>Executing {actionLoading} command...</span>
                  </div>
                ) : output ? (
                  <pre className="whitespace-pre-wrap">{output}</pre>
                ) : (
                  <span className="text-gray-500">No output yet. Execute a command to see results.</span>
                )}
              </div>
            </div>
          )}

          {/* Service Information */}
          <div className="mt-6 bg-gray-50 rounded-lg p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Service Information</h3>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-gray-600">Service Name:</span>
                <span className="ml-2 font-mono font-medium">{service.name}</span>
              </div>
              <div>
                <span className="text-gray-600">Current Status:</span>
                <span className={`ml-2 font-medium ${
                  service.status === 'running' ? 'text-green-600' : 'text-red-600'
                }`}>
                  {service.status}
                </span>
              </div>
              {service.port && (
                <div>
                  <span className="text-gray-600">Port:</span>
                  <span className="ml-2 font-mono font-medium">{service.port}</span>
                </div>
              )}
              <div>
                <span className="text-gray-600">Service ID:</span>
                <span className="ml-2 font-mono font-medium">{service.id}</span>
              </div>
            </div>
          </div>

          {/* Close Button */}
          <div className="mt-6 flex justify-end">
            <Button
              onClick={onClose}
              variant="outline"
              className="px-6 py-2"
            >
              Close Panel
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ServiceControlPanel;