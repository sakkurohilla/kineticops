import React, { useState } from 'react';
import { 
  Cpu, 
  HardDrive, 
  Activity, 
  Network, 
  Clock, 
  TrendingUp,
  RefreshCw,
  AlertCircle
} from 'lucide-react';
import Card from '../common/Card';
import useHostMetrics from '../../hooks/useHostMetrics';
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  Tooltip,
} from 'recharts';

interface HostDashboardProps {
  hostId: number;
}

interface MetricData {
  cpu_usage: number;
  memory_usage: number;
  memory_total: number;
  memory_used: number;
  disk_usage: number;
  disk_total: number;
  disk_used: number;
  network_in: number;
  network_out: number;
  uptime: number;
  load_average: string;
  timestamp: string;
}

const HostDashboard: React.FC<HostDashboardProps> = ({ hostId }) => {
  const [autoRefresh, setAutoRefresh] = useState(true);
  const { metrics, series, loading, error, refetch } = useHostMetrics(hostId, autoRefresh);
  const m = metrics as MetricData | null;
  const cpu = m?.cpu_usage ?? 0;
  const memory = m?.memory_usage ?? 0;
  const memory_used = m?.memory_used ?? 0;
  const memory_total = m?.memory_total ?? 0;
  const disk = m?.disk_usage ?? 0;
  const disk_used = m?.disk_used ?? 0;
  const disk_total = m?.disk_total ?? 0;
  const uptimeVal = m?.uptime ?? 0;
  const loadAvg = m?.load_average ?? '';
  const netIn = m?.network_in ?? 0;
  const netOut = m?.network_out ?? 0;

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (days > 0) return `${days}d ${hours}h ${minutes}m`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  const getUsageColor = (percentage: number) => {
    if (percentage >= 90) return 'text-red-600';
    if (percentage >= 70) return 'text-orange-500';
    return 'text-green-600';
  };

  const getUsageBg = (percentage: number) => {
    if (percentage >= 90) return 'bg-red-600';
    if (percentage >= 70) return 'bg-orange-500';
    return 'bg-green-600';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading metrics...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6">
        <div className="flex items-center gap-3">
          <AlertCircle className="w-6 h-6 text-red-600" />
          <div>
            <h3 className="font-semibold text-red-900">Failed to Load Metrics</h3>
            <p className="text-sm text-red-700 mt-1">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-8 text-center">
        <Clock className="w-12 h-12 text-yellow-600 mx-auto mb-3" />
        <h3 className="text-lg font-semibold text-yellow-900 mb-2">No Metrics Yet</h3>
        <p className="text-sm text-yellow-700">
          Metrics collection is in progress. Please wait for the first data collection cycle (up to 60 seconds).
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header with Refresh */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Real-time Metrics</h2>
        <div className="flex items-center gap-4">
          <label className="flex items-center gap-2 text-sm text-gray-600">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            Auto-refresh (30s)
          </label>
          <button
            onClick={() => refetch()}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
            title="Refresh now"
          >
            <RefreshCw className="w-5 h-5 text-gray-600" />
          </button>
        </div>
      </div>

      {/* Last Updated */}
      <div className="text-sm text-gray-500">
        Last updated: {m?.timestamp ? new Date(m.timestamp).toLocaleString() : 'N/A'}
      </div>

      {/* Metric Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* CPU Usage */}
        <Card className="hover:shadow-lg transition-shadow">
          <div className="flex items-start justify-between mb-4">
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">CPU Usage</p>
              <p className={`text-3xl font-bold ${getUsageColor(cpu)}`}>
                {cpu.toFixed(1)}%
              </p>
            </div>
            <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
              <Cpu className="w-6 h-6 text-blue-600" />
            </div>
          </div>
          {/* Progress Bar */}
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full transition-all duration-500 ${getUsageBg(cpu)}`}
              style={{ width: `${Math.min(cpu, 100)}%` }}
            ></div>
          </div>
          {/* Sparkline */}
          <div className="mt-3 h-16" style={{ minHeight: 48 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={series} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                <XAxis dataKey="timestamp" hide />
                <Tooltip formatter={(value: any) => [value, 'CPU']} />
                <Line type="monotone" dataKey="cpu_usage" stroke="#1D4ED8" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Card>

        {/* Memory Usage */}
        <Card className="hover:shadow-lg transition-shadow">
          <div className="flex items-start justify-between mb-4">
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Memory Usage</p>
              <p className={`text-3xl font-bold ${getUsageColor(memory)}`}>
                {memory.toFixed(1)}%
              </p>
            </div>
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
              <Activity className="w-6 h-6 text-green-600" />
            </div>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2 mb-2">
            <div
              className={`h-2 rounded-full transition-all duration-500 ${getUsageBg(memory)}`}
              style={{ width: `${Math.min(memory, 100)}%` }}
            ></div>
          </div>
          <p className="text-xs text-gray-500">
            {memory_used.toFixed(1)} MB / {memory_total.toFixed(1)} MB
          </p>
          <div className="mt-3 h-16" style={{ minHeight: 48 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={series} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                <XAxis dataKey="timestamp" hide />
                <Tooltip formatter={(value: any) => [value, 'Mem']} />
                <Line type="monotone" dataKey="memory_usage" stroke="#10B981" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Card>

        {/* Disk Usage */}
        <Card className="hover:shadow-lg transition-shadow">
          <div className="flex items-start justify-between mb-4">
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Disk Usage</p>
              <p className={`text-3xl font-bold ${getUsageColor(disk)}`}>
                {disk.toFixed(1)}%
              </p>
            </div>
            <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center">
              <HardDrive className="w-6 h-6 text-purple-600" />
            </div>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2 mb-2">
            <div
              className={`h-2 rounded-full transition-all duration-500 ${getUsageBg(disk)}`}
              style={{ width: `${Math.min(disk, 100)}%` }}
            ></div>
          </div>
          <p className="text-xs text-gray-500">
            {disk_used.toFixed(1)} GB / {disk_total.toFixed(1)} GB
          </p>
          <div className="mt-3 h-16" style={{ minHeight: 48 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={series} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                <XAxis dataKey="timestamp" hide />
                <Tooltip formatter={(value: any) => [value, 'Disk']} />
                <Line type="monotone" dataKey="disk_usage" stroke="#7C3AED" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Card>

        {/* Uptime */}
        <Card className="hover:shadow-lg transition-shadow">
          <div className="flex items-start justify-between mb-4">
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Uptime</p>
              <p className="text-3xl font-bold text-blue-600">
                {formatUptime(uptimeVal)}
              </p>
            </div>
            <div className="w-12 h-12 bg-indigo-100 rounded-lg flex items-center justify-center">
              <Clock className="w-6 h-6 text-indigo-600" />
            </div>
          </div>
          <p className="text-xs text-gray-500">
            Load: {loadAvg || 'N/A'}
          </p>
          <div className="mt-3 h-16" style={{ minHeight: 48 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={series} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                <XAxis dataKey="timestamp" hide />
                <Tooltip formatter={(value: any) => [value, 'Uptime']} />
                <Line type="monotone" dataKey="uptime" stroke="#0EA5A4" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </div>

      {/* Network Statistics */}
      <Card>
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 bg-cyan-100 rounded-lg flex items-center justify-center">
            <Network className="w-5 h-5 text-cyan-600" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-gray-900">Network I/O</h3>
            <p className="text-sm text-gray-500">Real-time network statistics</p>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-6">
          <div className="p-4 bg-green-50 rounded-lg border border-green-200">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp className="w-5 h-5 text-green-600" />
              <p className="text-sm font-medium text-green-900">Incoming</p>
            </div>
            <p className="text-2xl font-bold text-green-600">
              {netIn.toFixed(2)} MB
            </p>
            <p className="text-xs text-green-700 mt-1">Total received</p>
          </div>

          <div className="p-4 bg-blue-50 rounded-lg border border-blue-200">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp className="w-5 h-5 text-blue-600 transform rotate-180" />
              <p className="text-sm font-medium text-blue-900">Outgoing</p>
            </div>
            <p className="text-2xl font-bold text-blue-600">
              {netOut.toFixed(2)} MB
            </p>
            <p className="text-xs text-blue-700 mt-1">Total sent</p>
          </div>
        </div>
      </Card>

      {/* System Health Summary */}
      <Card className="bg-gradient-to-r from-blue-50 to-purple-50 border-blue-200">
        <div className="flex items-start gap-4">
          <div className="w-12 h-12 bg-white rounded-lg flex items-center justify-center shadow-sm">
            <Activity className="w-6 h-6 text-blue-600" />
          </div>
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">System Health</h3>
            <div className="space-y-2">
              {cpu < 70 && memory < 70 && disk < 70 ? (
                <p className="text-sm text-green-700 font-medium">
                  ✓ All systems operating normally
                </p>
              ) : (
                <>
                  {cpu >= 90 && (
                    <p className="text-sm text-red-700 font-medium">⚠ High CPU usage detected</p>
                  )}
                  {memory >= 90 && (
                    <p className="text-sm text-red-700 font-medium">⚠ High memory usage detected</p>
                  )}
                  {disk >= 90 && (
                    <p className="text-sm text-red-700 font-medium">⚠ Low disk space warning</p>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default HostDashboard;