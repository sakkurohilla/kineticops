import React from 'react';
import { Server, Activity, HardDrive, Cpu, Eye, Trash2 } from 'lucide-react';
import { Host } from '../../types';
import Badge from '../common/Badge';
import Button from '../common/Button';
import { useNavigate } from 'react-router-dom';

interface HostCardProps {
  host: Host;
  onDelete: (id: number) => void;
}

const HostCard: React.FC<HostCardProps> = ({ host, onDelete }) => {
  const navigate = useNavigate();

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'online':
        return 'success';
      case 'offline':
        return 'error';
      case 'warning':
        return 'warning';
      default:
        return 'info';
    }
  };

  const getStatusDot = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'online':
        return 'bg-green-500';
      case 'offline':
        return 'bg-red-500';
      case 'warning':
        return 'bg-orange-500';
      default:
        return 'bg-gray-500';
    }
  };

  const formatDate = (dateString: string | undefined) => {
    if (!dateString) return 'Never';
    const date = new Date(dateString);
    const now = new Date();
    const diff = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diff < 60) return 'Just now';
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    return `${Math.floor(diff / 86400)}d ago`;
  };

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-xl transition-all duration-300 group overflow-hidden">
      {/* Status Bar */}
      <div className={`h-1 ${host.agent_status === 'online' ? 'bg-green-500' : 'bg-gray-300'}`}></div>

      <div className="p-6">
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-start gap-4">
            <div className="w-14 h-14 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform">
              <Server className="w-7 h-7 text-white" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-gray-900 mb-1">{host.hostname || 'Unnamed Host'}</h3>
              <p className="text-sm text-gray-500 flex items-center gap-2">
                <span className={`w-2 h-2 rounded-full ${getStatusDot(host.agent_status || 'offline')}`}></span>
                {host.ip}
              </p>
            </div>
          </div>

          <Badge variant={getStatusColor(host.agent_status || 'offline')} size="sm">
            {host.agent_status || 'offline'}
          </Badge>
        </div>

        {/* Host Info */}
        <div className="space-y-2 mb-4">
          <div className="flex items-center justify-between text-sm">
            <span className="text-gray-600">OS</span>
            <span className="font-medium text-gray-900">{host.os || 'Linux'}</span>
          </div>
          <div className="flex items-center justify-between text-sm">
            <span className="text-gray-600">Group</span>
            <span className="font-medium text-gray-900">{host.group || 'default'}</span>
          </div>
          <div className="flex items-center justify-between text-sm">
            <span className="text-gray-600">Last Seen</span>
            <span className="font-medium text-gray-900">{formatDate(host.last_seen)}</span>
          </div>
        </div>

        {/* Tags */}
        {host.tags && (
          <div className="flex flex-wrap gap-2 mb-4">
            {host.tags.split(',').map((tag, idx) => (
              <span
                key={idx}
                className="px-2 py-1 bg-blue-50 text-blue-700 text-xs font-medium rounded-md"
              >
                {tag.trim()}
              </span>
            ))}
          </div>
        )}

        {/* Quick Stats Placeholder */}
        <div className="grid grid-cols-3 gap-3 mb-4 pt-4 border-t border-gray-100">
          <div className="text-center">
            <Cpu className="w-5 h-5 text-blue-600 mx-auto mb-1" />
            <p className="text-xs text-gray-500">CPU</p>
          </div>
          <div className="text-center">
            <Activity className="w-5 h-5 text-green-600 mx-auto mb-1" />
            <p className="text-xs text-gray-500">Memory</p>
          </div>
          <div className="text-center">
            <HardDrive className="w-5 h-5 text-purple-600 mx-auto mb-1" />
            <p className="text-xs text-gray-500">Disk</p>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          <Button
            variant="primary"
            size="sm"
            fullWidth
            onClick={() => navigate(`/hosts/${host.id}`)}
          >
            <Eye className="w-4 h-4" />
            View Details
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onDelete(host.id)}
            className="text-red-600 hover:bg-red-50 border-red-200"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  );
};

export default HostCard;