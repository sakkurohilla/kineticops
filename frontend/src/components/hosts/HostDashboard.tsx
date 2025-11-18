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
import GlassCard from '../common/GlassCard';
// Removed SimpleAnalytics per user request: simplify hosts tab to original 3-card summary
import useHostMetrics from '../../hooks/useHostMetrics';
import {
  ResponsiveContainer,
  XAxis,
  Tooltip,
  AreaChart,
  Area,
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
  // Use undefined-ish defaults so we can show N/A when data is missing
  const cpu = m?.cpu_usage;
  const memory = m?.memory_usage;
  // backend stores memory totals in MB (not bytes) for host_metrics snapshots
  const memory_used = m?.memory_used ?? 0; // MB
  const memory_total = m?.memory_total ?? 0; // MB
  const disk = m?.disk_usage;
  // backend stores disk totals in GB (host_metrics snapshot conversion)
  const disk_used = m?.disk_used ?? 0; // GB
  const uptimeVal = m?.uptime ?? null;
  const loadAvg = m?.load_average ?? '';
  // network values are already in MB from backend (host_metrics table)
  const netIn = m?.network_in !== undefined && m.network_in !== null ? m.network_in : 0;
  const netOut = m?.network_out !== undefined && m.network_out !== null ? m.network_out : 0;

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (days > 0) return `${days}d ${hours}h ${minutes}m`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
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

      {/* Metric Cards Grid - Glassmorphism Style */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        {/* CPU Usage - Glassmorphism Card */}
        <GlassCard gradient="from-blue-500/20 to-cyan-500/20">
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-gray-700">CPU Usage</h3>
              <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-cyan-500 rounded-xl flex items-center justify-center shadow-lg">
                <Cpu className="w-5 h-5 text-white" />
              </div>
            </div>
            
            {/* Circular Progress */}
            <div className="flex items-center justify-center mb-4">
              <div className="relative w-32 h-32">
                <svg className="transform -rotate-90 w-32 h-32">
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="currentColor"
                    strokeWidth="8"
                    fill="none"
                    className="text-gray-200"
                  />
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="url(#cpuGradient)"
                    strokeWidth="8"
                    fill="none"
                    strokeLinecap="round"
                    strokeDasharray={`${2 * Math.PI * 56}`}
                    strokeDashoffset={`${2 * Math.PI * 56 * (1 - (cpu ?? 0) / 100)}`}
                    className="transition-all duration-500"
                  />
                  <defs>
                    <linearGradient id="cpuGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                      <stop offset="0%" stopColor="#3B82F6" />
                      <stop offset="100%" stopColor="#06B6D4" />
                    </linearGradient>
                  </defs>
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                  <span className="text-3xl font-bold text-gray-900">
                    {typeof cpu === 'number' ? `${cpu.toFixed(0)}%` : 'N/A'}
                  </span>
                  <span className="text-xs text-gray-500">Time</span>
                </div>
              </div>
            </div>
          </div>
        </GlassCard>

        {/* Memory Usage - Glassmorphism Card */}
        <GlassCard gradient="from-emerald-500/20 to-teal-500/20">
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-gray-700">Memory Usage</h3>
              <div className="w-10 h-10 bg-gradient-to-br from-emerald-500 to-teal-500 rounded-xl flex items-center justify-center shadow-lg">
                <Activity className="w-5 h-5 text-white" />
              </div>
            </div>
            
            {/* Area Chart */}
            <div className="h-40">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={series.slice(-7)} margin={{ top: 10, right: 0, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="memoryGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10B981" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="#14B8A6" stopOpacity={0.1}/>
                    </linearGradient>
                  </defs>
                  <XAxis dataKey="timestamp" hide />
                  <Tooltip 
                    contentStyle={{ background: 'rgba(255,255,255,0.9)', border: 'none', borderRadius: '12px', backdropFilter: 'blur(10px)' }}
                    formatter={(value: any) => [`${value}%`, 'Memory']} 
                  />
                  <Area 
                    type="monotone" 
                    dataKey="memory_usage" 
                    stroke="#10B981" 
                    strokeWidth={2}
                    fill="url(#memoryGradient)" 
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
            
            <div className="mt-2 flex items-center justify-between">
              <div>
                <p className="text-xs text-gray-500">3 Day</p>
                <p className="text-sm font-semibold text-gray-700">
                  {typeof memory === 'number' ? `${memory.toFixed(1)}%` : 'N/A'}
                </p>
              </div>
              <div className="text-right">
                <p className="text-xs text-gray-500">Available</p>
                <p className="text-sm font-semibold text-emerald-600">
                  {(memory_total - memory_used).toFixed(0)} MB
                </p>
              </div>
            </div>
          </div>
        </GlassCard>

        {/* Disk Usage - Glassmorphism Card */}
        <GlassCard gradient="from-purple-500/20 to-pink-500/20">
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-gray-700">Disk Usage</h3>
              <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-pink-500 rounded-xl flex items-center justify-center shadow-lg">
                <HardDrive className="w-5 h-5 text-white" />
              </div>
            </div>
            
            {/* Circular Progress */}
            <div className="flex items-center justify-center mb-4">
              <div className="relative w-32 h-32">
                <svg className="transform -rotate-90 w-32 h-32">
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="currentColor"
                    strokeWidth="8"
                    fill="none"
                    className="text-gray-200"
                  />
                  <circle
                    cx="64"
                    cy="64"
                    r="56"
                    stroke="url(#diskGradient)"
                    strokeWidth="8"
                    fill="none"
                    strokeLinecap="round"
                    strokeDasharray={`${2 * Math.PI * 56}`}
                    strokeDashoffset={`${2 * Math.PI * 56 * (1 - (disk ?? 0) / 100)}`}
                    className="transition-all duration-500"
                  />
                  <defs>
                    <linearGradient id="diskGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                      <stop offset="0%" stopColor="#A855F7" />
                      <stop offset="100%" stopColor="#EC4899" />
                    </linearGradient>
                  </defs>
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                  <span className="text-3xl font-bold text-gray-900">
                    {typeof disk === 'number' ? `${disk.toFixed(0)}%` : 'N/A'}
                  </span>
                  <span className="text-xs text-gray-500">{disk_used.toFixed(0)}GB</span>
                </div>
              </div>
            </div>
          </div>
        </GlassCard>

        {/* Network I/O - Glassmorphism Card */}
        <GlassCard gradient="from-orange-500/20 to-red-500/20">
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-gray-700">Network I/O</h3>
              <div className="w-10 h-10 bg-gradient-to-br from-orange-500 to-red-500 rounded-xl flex items-center justify-center shadow-lg">
                <Network className="w-5 h-5 text-white" />
              </div>
            </div>
            
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-gradient-to-r from-orange-500 to-red-500 rounded-full animate-pulse" />
                  <span className="text-xs text-gray-600">Incoming</span>
                </div>
                <span className="text-sm font-semibold text-gray-900">{netIn.toFixed(1)} MB</span>
              </div>
              
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="h-2 rounded-full bg-gradient-to-r from-orange-500 to-red-500 transition-all duration-500"
                  style={{ width: `${Math.min((netIn / (netIn + netOut)) * 100 || 50, 100)}%` }}
                />
              </div>
              
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-gradient-to-r from-red-500 to-pink-500 rounded-full animate-pulse" />
                  <span className="text-xs text-gray-600">Outgoing</span>
                </div>
                <span className="text-sm font-semibold text-gray-900">{netOut.toFixed(1)} MB</span>
              </div>
              
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="h-2 rounded-full bg-gradient-to-r from-red-500 to-pink-500 transition-all duration-500"
                  style={{ width: `${Math.min((netOut / (netIn + netOut)) * 100 || 50, 100)}%` }}
                />
              </div>
            </div>
          </div>
        </GlassCard>
      </div>

      {/* System Information Cards */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6">
        {/* Uptime & Load Average */}
        <GlassCard gradient="from-indigo-500/20 to-blue-500/20">
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-gray-700">System Uptime</h3>
              <div className="w-10 h-10 bg-gradient-to-br from-indigo-500 to-blue-500 rounded-xl flex items-center justify-center shadow-lg">
                <Clock className="w-5 h-5 text-white" />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-3xl font-bold text-gray-900">
                  {uptimeVal !== null ? formatUptime(uptimeVal) : 'N/A'}
                </p>
                <p className="text-xs text-gray-500 mt-1">Load: {loadAvg || 'N/A'}</p>
              </div>
            </div>
          </div>
        </GlassCard>

        {/* Analytics Summary */}
        <GlassCard gradient="from-cyan-500/20 to-blue-500/20">
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-gray-700">Performance Trend</h3>
              <div className="w-10 h-10 bg-gradient-to-br from-cyan-500 to-blue-500 rounded-xl flex items-center justify-center shadow-lg">
                <TrendingUp className="w-5 h-5 text-white" />
              </div>
            </div>
            <div className="h-20">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={series.slice(-10)} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="cpuTrendGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#06B6D4" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="#3B82F6" stopOpacity={0.1}/>
                    </linearGradient>
                  </defs>
                  <XAxis dataKey="timestamp" hide />
                  <Tooltip 
                    contentStyle={{ background: 'rgba(255,255,255,0.9)', border: 'none', borderRadius: '12px' }}
                    formatter={(value: any) => [`${value}%`, 'CPU']} 
                  />
                  <Area 
                    type="monotone" 
                    dataKey="cpu_usage" 
                    stroke="#06B6D4" 
                    strokeWidth={2}
                    fill="url(#cpuTrendGradient)" 
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </GlassCard>
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
              {netIn !== null ? `${netIn.toFixed(2)} MB` : 'N/A'}
            </p>
            <p className="text-xs text-green-700 mt-1">Total received</p>
          </div>

          <div className="p-4 bg-blue-50 rounded-lg border border-blue-200">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp className="w-5 h-5 text-blue-600 transform rotate-180" />
              <p className="text-sm font-medium text-blue-900">Outgoing</p>
            </div>
            <p className="text-2xl font-bold text-blue-600">
              {netOut !== null ? `${netOut.toFixed(2)} MB` : 'N/A'}
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
              {typeof cpu === 'number' && typeof memory === 'number' && typeof disk === 'number' && cpu < 70 && memory < 70 && disk < 70 ? (
                <p className="text-sm text-green-700 font-medium">
                  ✓ All systems operating normally
                </p>
              ) : (
                <>
                  {typeof cpu === 'number' && cpu >= 90 && (
                    <p className="text-sm text-red-700 font-medium">⚠ High CPU usage detected</p>
                  )}
                  {typeof memory === 'number' && memory >= 90 && (
                    <p className="text-sm text-red-700 font-medium">⚠ High memory usage detected</p>
                  )}
                  {typeof disk === 'number' && disk >= 90 && (
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