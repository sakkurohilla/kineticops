export const formatServiceStatus = (status: string): string => {
  return status.charAt(0).toUpperCase() + status.slice(1).toLowerCase();
};

export const getActionColor = (action: string): string => {
  const colors = {
    start: 'bg-green-600 hover:bg-green-700',
    stop: 'bg-red-600 hover:bg-red-700',
    restart: 'bg-yellow-600 hover:bg-yellow-700',
    enable: 'bg-blue-600 hover:bg-blue-700',
    disable: 'bg-gray-600 hover:bg-gray-700',
  };
  return colors[action as keyof typeof colors] || 'bg-gray-600 hover:bg-gray-700';
};

export const calculateServiceHealth = (service: any): number => {
  let score = 100;
  
  if (service.status !== 'running') score -= 50;
  if (service.cpu_usage > 80) score -= 20;
  if (service.memory_usage > 80) score -= 20;
  if (service.error_rate && parseFloat(service.error_rate) > 1) score -= 10;
  
  return Math.max(0, score);
};

export const parseCommandOutput = (output: string): { success: boolean; message: string } => {
  const lines = output.split('\n').filter(line => line.trim());
  const lastLine = lines[lines.length - 1] || '';
  
  const success = !lastLine.toLowerCase().includes('error') && 
                  !lastLine.toLowerCase().includes('failed');
  
  return {
    success,
    message: output
  };
};

export const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
};

export const getHealthScoreColor = (score: number): string => {
  if (score >= 90) return 'text-green-600';
  if (score >= 70) return 'text-yellow-600';
  return 'text-red-600';
};

export const formatLatency = (ms: number): string => {
  if (ms < 1) return `${(ms * 1000).toFixed(0)}μs`;
  if (ms < 1000) return `${ms.toFixed(1)}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
};

export const getStatusBadgeColor = (status: string): string => {
  const colors = {
    running: 'bg-green-100 text-green-800',
    stopped: 'bg-red-100 text-red-800',
    starting: 'bg-yellow-100 text-yellow-800',
    error: 'bg-red-100 text-red-800',
    warning: 'bg-yellow-100 text-yellow-800',
    healthy: 'bg-green-100 text-green-800',
  };
  return colors[status.toLowerCase() as keyof typeof colors] || 'bg-gray-100 text-gray-800';
};