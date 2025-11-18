import React, { useState, useEffect } from 'react';
import { Activity, Database, Zap, Users, AlertCircle } from 'lucide-react';
import Card from '../common/Card';
import axios from 'axios';

interface HealthData {
  status: string;
  postgres: string;
  redis: string;
  kafka: string;
  ws_clients: number;
  metrics_batcher_queue_len: number;
  metrics_batcher_last_flush: string;
}

const SystemHealth: React.FC = () => {
  const [health, setHealth] = useState<HealthData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchHealth = async () => {
    try {
      const response = await axios.get(`${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/health`);
      setHealth(response.data);
      setError('');
    } catch (err) {
      setError('Failed to fetch health data');
      console.error('Health check error:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
    const interval = setInterval(fetchHealth, 5000); // Refresh every 5 seconds
    return () => clearInterval(interval);
  }, []);

  const getQueueStatus = (queueLength: number) => {
    if (queueLength > 400) return { color: 'text-red-600', bg: 'bg-red-100', status: 'High' };
    if (queueLength > 200) return { color: 'text-yellow-600', bg: 'bg-yellow-100', status: 'Medium' };
    return { color: 'text-green-600', bg: 'bg-green-100', status: 'Low' };
  };

  const getServiceStatus = (service: string) => {
    if (service === 'connected') return { color: 'text-green-600', bg: 'bg-green-100', text: 'Connected' };
    return { color: 'text-red-600', bg: 'bg-red-100', text: 'Disconnected' };
  };

  if (isLoading) {
    return (
      <Card>
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
          <div className="grid grid-cols-2 gap-4">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="p-4 bg-gray-100 rounded-lg">
                <div className="h-4 bg-gray-200 rounded w-1/2 mb-2"></div>
                <div className="h-8 bg-gray-200 rounded w-3/4"></div>
              </div>
            ))}
          </div>
        </div>
      </Card>
    );
  }

  if (error || !health) {
    return (
      <Card>
        <div className="text-center py-8">
          <AlertCircle className="w-12 h-12 text-red-300 mx-auto mb-3" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">Health Check Failed</h3>
          <p className="text-gray-500">{error || 'No health data available'}</p>
        </div>
      </Card>
    );
  }

  const queueStatus = getQueueStatus(health.metrics_batcher_queue_len || 0);
  const postgresStatus = getServiceStatus(health.postgres);
  const redisStatus = getServiceStatus(health.redis);
  const kafkaStatus = getServiceStatus(health.kafka);
  
  const healthMetrics = [
    {
      label: 'PostgreSQL',
      value: postgresStatus.text,
      icon: Database,
      color: postgresStatus.color,
      bgColor: postgresStatus.bg
    },
    {
      label: 'Redis',
      value: redisStatus.text,
      icon: Database,
      color: redisStatus.color,
      bgColor: redisStatus.bg
    },
    {
      label: 'Kafka',
      value: kafkaStatus.text,
      icon: Database,
      color: kafkaStatus.color,
      bgColor: kafkaStatus.bg
    },
    {
      label: 'Batcher Queue',
      value: `${health.metrics_batcher_queue_len || 0} metrics`,
      icon: Zap,
      color: queueStatus.color,
      bgColor: queueStatus.bg,
      subtext: queueStatus.status
    },
    {
      label: 'WebSocket Clients',
      value: `${health.ws_clients || 0} connected`,
      icon: Users,
      color: 'text-blue-600',
      bgColor: 'bg-blue-100'
    },
    {
      label: 'System Status',
      value: health.status === 'ok' ? 'Healthy' : 'Degraded',
      icon: Activity,
      color: health.status === 'ok' ? 'text-green-600' : 'text-red-600',
      bgColor: health.status === 'ok' ? 'bg-green-100' : 'bg-red-100'
    }
  ];

  const lastFlush = health.metrics_batcher_last_flush 
    ? new Date(health.metrics_batcher_last_flush).toLocaleTimeString()
    : 'Never';

  return (
    <Card>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-900">System Health</h2>
        <div className="flex items-center space-x-2">
          <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
          <span className="text-sm text-gray-600">Live</span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {healthMetrics.map((metric, index) => {
          const Icon = metric.icon;
          return (
            <div key={index} className="p-4 bg-gray-50 rounded-lg border border-gray-200">
              <div className="flex items-center space-x-3 mb-3">
                <div className={`p-2 rounded-lg ${metric.bgColor}`}>
                  <Icon className={`w-5 h-5 ${metric.color}`} />
                </div>
                <span className="text-sm font-medium text-gray-700">{metric.label}</span>
              </div>
              <div className="flex flex-col">
                <span className={`text-xl font-bold ${metric.color}`}>
                  {metric.value}
                </span>
                {metric.subtext && (
                  <span className="text-xs text-gray-500 mt-1">{metric.subtext} load</span>
                )}
              </div>
            </div>
          );
        })}
      </div>

      <div className="mt-4 p-4 bg-gray-50 rounded-lg border border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Zap className="w-4 h-4 text-gray-600" />
            <span className="text-sm font-medium text-gray-700">Last Metrics Flush</span>
          </div>
          <span className="text-sm text-gray-600">{lastFlush}</span>
        </div>
      </div>
    </Card>
  );
};

export default SystemHealth;
