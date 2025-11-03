import React, { useState, useEffect } from 'react';
import { Cpu, HardDrive, Activity, Zap } from 'lucide-react';

interface RealTimeMetricsProps {
  hostId?: number;
  sessionToken?: string;
}

interface MetricGaugeProps {
  label: string;
  value: number;
  icon: React.ReactNode;
  color: string;
  unit?: string;
}

const MetricGauge: React.FC<MetricGaugeProps> = ({ label, value, icon, color, unit = '%' }) => {
  const circumference = 2 * Math.PI * 40;
  const strokeDasharray = circumference;
  const strokeDashoffset = circumference - (value / 100) * circumference;

  return (
    <div className="bg-white rounded-xl p-6 shadow-lg border border-gray-200">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">{label}</h3>
        <div className={`p-2 rounded-lg ${color}`}>
          {icon}
        </div>
      </div>
      
      <div className="relative w-32 h-32 mx-auto mb-4">
        <svg className="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
          {/* Background circle */}
          <circle
            cx="50"
            cy="50"
            r="40"
            fill="none"
            stroke="#e5e7eb"
            strokeWidth="8"
          />
          {/* Progress circle */}
          <circle
            cx="50"
            cy="50"
            r="40"
            fill="none"
            stroke="currentColor"
            strokeWidth="8"
            strokeLinecap="round"
            strokeDasharray={strokeDasharray}
            strokeDashoffset={strokeDashoffset}
            className={`transition-all duration-1000 ease-out ${
              color.includes('orange') ? 'text-orange-500' :
              color.includes('blue') ? 'text-blue-500' :
              color.includes('green') ? 'text-green-500' :
              'text-purple-500'
            }`}
          />
        </svg>
        
        {/* Center value */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center">
            <div className={`text-3xl font-bold ${
              color.includes('orange') ? 'text-orange-600' :
              color.includes('blue') ? 'text-blue-600' :
              color.includes('green') ? 'text-green-600' :
              'text-purple-600'
            }`}>
              {value.toFixed(1)}
            </div>
            <div className="text-sm text-gray-500">{unit}</div>
          </div>
        </div>
      </div>

      {/* Status indicator */}
      <div className="flex items-center justify-center">
        <div className={`w-3 h-3 rounded-full mr-2 ${
          value < 50 ? 'bg-green-400 animate-pulse' :
          value < 80 ? 'bg-yellow-400 animate-pulse' :
          'bg-red-400 animate-pulse'
        }`}></div>
        <span className={`text-sm font-medium ${
          value < 50 ? 'text-green-600' :
          value < 80 ? 'text-yellow-600' :
          'text-red-600'
        }`}>
          {value < 50 ? 'Optimal' : value < 80 ? 'Warning' : 'Critical'}
        </span>
      </div>
    </div>
  );
};

const RealTimeMetrics: React.FC<RealTimeMetricsProps> = () => {
  const [metrics, setMetrics] = useState({
    cpu: 0,
    memory: 0,
    disk: 0,
    network: 0
  });

  // Simulate real-time metrics updates
  useEffect(() => {
    const interval = setInterval(() => {
      setMetrics(prev => ({
        cpu: Math.max(0, Math.min(100, prev.cpu + (Math.random() - 0.5) * 10)),
        memory: Math.max(0, Math.min(100, prev.memory + (Math.random() - 0.5) * 5)),
        disk: Math.max(0, Math.min(100, prev.disk + (Math.random() - 0.5) * 2)),
        network: Math.max(0, Math.min(100, prev.network + (Math.random() - 0.5) * 15))
      }));
    }, 2000);

    // Initialize with random values
    setMetrics({
      cpu: 20 + Math.random() * 60,
      memory: 30 + Math.random() * 50,
      disk: 40 + Math.random() * 40,
      network: 10 + Math.random() * 30
    });

    return () => clearInterval(interval);
  }, []);

  const metricsConfig = [
    {
      label: 'CPU Usage',
      value: metrics.cpu,
      icon: <Cpu className="w-5 h-5 text-white" />,
      color: 'bg-orange-500',
      unit: '%'
    },
    {
      label: 'Memory Usage',
      value: metrics.memory,
      icon: <Activity className="w-5 h-5 text-white" />,
      color: 'bg-blue-500',
      unit: '%'
    },
    {
      label: 'Disk Usage',
      value: metrics.disk,
      icon: <HardDrive className="w-5 h-5 text-white" />,
      color: 'bg-green-500',
      unit: '%'
    },
    {
      label: 'Network I/O',
      value: metrics.network,
      icon: <Zap className="w-5 h-5 text-white" />,
      color: 'bg-purple-500',
      unit: '%'
    }
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Real-Time Metrics</h2>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-green-400 rounded-full animate-pulse"></div>
          <span className="text-sm text-gray-600">Live Updates</span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {metricsConfig.map((metric, index) => (
          <MetricGauge
            key={index}
            label={metric.label}
            value={metric.value}
            icon={metric.icon}
            color={metric.color}
            unit={metric.unit}
          />
        ))}
      </div>

      {/* Performance Summary */}
      <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-xl p-6 border border-blue-200">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Performance Summary</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600">
              {((100 - metrics.cpu + 100 - metrics.memory + 100 - metrics.disk) / 3).toFixed(0)}%
            </div>
            <div className="text-sm text-gray-600">Health Score</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600">
              {(Math.random() * 5 + 2).toFixed(1)}ms
            </div>
            <div className="text-sm text-gray-600">Avg Latency</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-purple-600">
              {(Math.random() * 2 + 1).toFixed(2)}K
            </div>
            <div className="text-sm text-gray-600">Req/sec</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-orange-600">
              {(Math.random() * 0.1).toFixed(3)}%
            </div>
            <div className="text-sm text-gray-600">Error Rate</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RealTimeMetrics;