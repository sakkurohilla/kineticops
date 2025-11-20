import React, { useState } from 'react';
import { Square, RotateCw, AlertCircle } from 'lucide-react';
import Button from '../common/Button';
import api from '../../services/api/client';

interface Process {
  pid: number;
  name: string;
  cpu: number;
  memory: number;
  status: string;
  user: string;
}

interface ProcessControlProps {
  hostId: number;
  process: Process;
  onActionComplete?: () => void;
}

const ProcessControl: React.FC<ProcessControlProps> = ({ hostId, process, onActionComplete }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');

  const handleAction = async (action: 'start' | 'stop' | 'restart', processName: string) => {
    if (!window.confirm(`Are you sure you want to ${action} ${processName}?`)) {
      return;
    }

    setLoading(true);
    setError('');

    try {
      await api.post(`/hosts/${hostId}/processes/${action}`, {
        process_name: processName,
        pid: process.pid,
      });

      if (onActionComplete) {
        onActionComplete();
      }
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to ${action} process`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => handleAction('restart', process.name)}
        disabled={loading}
        title="Restart process"
      >
        <RotateCw className={`w-3 h-3 ${loading ? 'animate-spin' : ''}`} />
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        onClick={() => handleAction('stop', process.name)}
        disabled={loading}
        title="Stop process"
      >
        <Square className="w-3 h-3" />
      </Button>

      {error && (
        <div className="flex items-center gap-1 text-red-600 text-xs">
          <AlertCircle className="w-3 h-3" />
          <span>{error}</span>
        </div>
      )}
    </div>
  );
};

export default ProcessControl;
