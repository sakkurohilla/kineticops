import React, { useState, useEffect } from 'react';
import { Activity, Cpu, HardDrive, Wifi, TrendingUp, TrendingDown } from 'lucide-react';
import Card from '../common/Card';
import { useMetrics } from '../../hooks/useMetrics';

interface LiveMetricsProps {
  hostId?: number;
}

const LiveMetrics: React.FC<LiveMetricsProps> = ({ hostId }) => {
  const { data, isLoading } = useMetrics('1h', hostId, true);
  const [trends, setTrends] = useState<Record<string, 'up' | 'down' | 'stable'>>({});

  // Calculate trends based on recent data
  useEffect(() => {
    const calculateTrend = (metricData: any[]) => {
      if (metricData.length < 2) return 'stable';
      const recent = metricData.slice(-5);
      const avg1 = recent.slice(0, Math.ceil(recent.length / 2)).reduce((sum, item) => sum + item.value, 0) / Math.ceil(recent.length / 2);
      const avg2 = recent.slice(Math.ceil(recent.length / 2)).reduce((sum, item) => sum + item.value, 0) / Math.floor(recent.length / 2);
      
      if (avg2 > avg1 + 5) return 'up';
      if (avg2 < avg1 - 5) return 'down';
      return 'stable';
    };

    setTrends({
      cpu: calculateTrend(data.cpu),
      memory: calculateTrend(data.memory),
      disk: calculateTrend(data.disk),
      network: calculateTrend(data.network),
    });
  }, [data]);

  const getLatestValue = (metricData: any[]) => {
    if (metricData.length === 0) return 0;
    return metricData[metricData.length - 1]?.value || 0;
  };

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up': return <TrendingUp className="w-4 h-4 text-red-500" />;
      case 'down': return <TrendingDown className="w-4 h-4 text-green-500" />;
      default: return <Activity className="w-4 h-4 text-gray-500" />;
    }
  };

  const getValueColor = (value: number, type: string) => {
    if (type === 'network') return 'text-purple-600';
    if (value >= 90) return 'text-red-600';
    if (value >= 70) return 'text-yellow-600';
    return 'text-green-600';
  };

  const metrics = [
    {
      id: 'cpu',
      label: 'CPU Usage',
      icon: Cpu,
      value: getLatestValue(data.cpu),
      unit: '%',
      color: 'bg-blue-100 text-blue-600'
    },
    {
      id: 'memory',
      label: 'Memory Usage',
      icon: Activity,
      value: getLatestValue(data.memory),
      unit: '%',
      color: 'bg-green-100 text-green-600'
    },
    {
      id: 'disk',
      label: 'Disk Usage',
      icon: HardDrive,
      value: getLatestValue(data.disk),
      unit: '%',
      color: 'bg-yellow-100 text-yellow-600'
    },
    {
      id: 'network',
      label: 'Network I/O',
      icon: Wifi,
      value: getLatestValue(data.network),
      unit: ' MB/s',
      color: 'bg-purple-100 text-purple-600'
    }
  ];

  const hasData = data.cpu.length > 0 || data.memory.length > 0 || data.disk.length > 0 || data.network.length > 0;

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

  if (!hasData) {
    return (
      <Card>
        <div className="text-center py-8">
          <Activity className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No Metrics Available</h3>
          <p className="text-gray-500">Metrics will appear here once hosts start reporting data</p>
        </div>
      </Card>
    );
  }

  return (
    <Card>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-900">Live Metrics</h2>
        <div className="flex items-center space-x-2">
          <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
          <span className="text-sm text-gray-600">Real-time</span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {metrics.map((metric) => {
          const Icon = metric.icon;
          const trend = trends[metric.id] || 'stable';
          
          return (
            <div key={metric.id} className="p-4 bg-gray-50 rounded-lg border border-gray-200 hover:shadow-md transition-shadow">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center space-x-3">
                  <div className={`p-2 rounded-lg ${metric.color}`}>
                    <Icon className="w-5 h-5" />
                  </div>
                  <span className="text-sm font-medium text-gray-700">{metric.label}</span>
                </div>
                {getTrendIcon(trend)}
              </div>
              
              <div className="flex items-baseline justify-between">
                <span className={`text-2xl font-bold ${getValueColor(metric.value, metric.id)}`}>
                  {metric.value.toFixed(1)}{metric.unit}
                </span>
                <div className="text-right">
                  <div className={`text-xs px-2 py-1 rounded-full ${
                    metric.value >= 90 ? 'bg-red-100 text-red-700' :
                    metric.value >= 70 ? 'bg-yellow-100 text-yellow-700' :
                    'bg-green-100 text-green-700'
                  }`}>
                    {metric.value >= 90 ? 'Critical' :
                     metric.value >= 70 ? 'Warning' : 'Normal'}
                  </div>
                </div>
              </div>

              {/* Mini progress bar */}
              <div className="mt-3 w-full bg-gray-200 rounded-full h-2">
                <div 
                  className={`h-2 rounded-full transition-all duration-500 ${
                    metric.value >= 90 ? 'bg-red-500' :
                    metric.value >= 70 ? 'bg-yellow-500' : 'bg-green-500'
                  }`}
                  style={{ width: `${Math.min(metric.value, 100)}%` }}
                ></div>
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
};

export default LiveMetrics;