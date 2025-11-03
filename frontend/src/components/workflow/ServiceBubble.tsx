import React from 'react';
import { CheckCircle, XCircle, Clock } from 'lucide-react';

interface ServiceBubbleProps {
  service: {
    id: number;
    name: string;
    status: string;
    port?: number;
    cpu_usage?: number;
    memory_usage?: number;
    latency?: string;
    error_rate?: string;
  };
  onClick: () => void;
  size?: 'sm' | 'md' | 'lg';
}

const ServiceBubble: React.FC<ServiceBubbleProps> = ({ service, onClick, size = 'md' }) => {
  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
      case 'active':
        return 'from-green-400 to-green-600';
      case 'stopped':
      case 'inactive':
        return 'from-red-400 to-red-600';
      case 'starting':
      case 'stopping':
        return 'from-yellow-400 to-yellow-600';
      default:
        return 'from-gray-400 to-gray-600';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
      case 'active':
        return <CheckCircle className="w-4 h-4" />;
      case 'stopped':
      case 'inactive':
        return <XCircle className="w-4 h-4" />;
      case 'starting':
      case 'stopping':
        return <Clock className="w-4 h-4 animate-spin" />;
      default:
        return <XCircle className="w-4 h-4" />;
    }
  };

  const sizeClasses = {
    sm: 'w-24 h-24',
    md: 'w-32 h-32',
    lg: 'w-40 h-40'
  };

  const textSizes = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base'
  };

  return (
    <div className="relative group">
      <div
        onClick={onClick}
        className={`${sizeClasses[size]} rounded-full mx-auto flex items-center justify-center cursor-pointer transition-all duration-300 transform hover:scale-110 hover:-translate-y-2 shadow-xl bg-gradient-to-br ${getStatusColor(service.status)} relative overflow-hidden`}
      >
        {/* Pulse animation for running services */}
        {service.status === 'running' && (
          <div className="absolute inset-0 rounded-full bg-white opacity-20 animate-ping"></div>
        )}
        
        <div className="text-center text-white z-10">
          <div className="mb-1">
            {getStatusIcon(service.status)}
          </div>
          <p className={`font-bold ${textSizes[size]} truncate max-w-full px-2`}>
            {service.name}
          </p>
          {service.port && (
            <p className={`${size === 'sm' ? 'text-xs' : 'text-xs'} opacity-90`}>
              :{service.port}
            </p>
          )}
        </div>

        {/* Health indicator */}
        <div className="absolute top-2 right-2">
          <div className={`w-3 h-3 rounded-full ${service.status === 'running' ? 'bg-green-300 animate-pulse' : 'bg-red-300'}`}></div>
        </div>

        {/* Performance metrics overlay */}
        {(service.cpu_usage !== undefined || service.memory_usage !== undefined) && (
          <div className="absolute bottom-1 left-1 right-1 bg-black bg-opacity-50 rounded text-xs text-white p-1">
            <div className="flex justify-between">
              {service.cpu_usage !== undefined && (
                <span>CPU: {service.cpu_usage}%</span>
              )}
              {service.memory_usage !== undefined && (
                <span>MEM: {service.memory_usage}%</span>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Hover tooltip */}
      <div className="absolute bottom-full mb-2 left-1/2 transform -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 z-50">
        <div className="bg-gray-900 text-white text-xs rounded-lg px-3 py-2 whitespace-nowrap shadow-xl">
          <div className="font-semibold mb-1">{service.name}</div>
          <div className="space-y-1">
            <div>Status: <span className={`font-medium ${service.status === 'running' ? 'text-green-400' : 'text-red-400'}`}>{service.status}</span></div>
            {service.port && <div>Port: {service.port}</div>}
            {service.latency && <div>Latency: {service.latency}</div>}
            {service.error_rate && <div>Error Rate: {service.error_rate}</div>}
          </div>
          {/* Arrow */}
          <div className="absolute top-full left-1/2 transform -translate-x-1/2 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900"></div>
        </div>
      </div>
    </div>
  );
};

export default ServiceBubble;