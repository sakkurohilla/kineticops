import React, { useState } from 'react';
import { 
  Cpu, 
  Activity, 
  Clock, 
  RefreshCw,
  AlertCircle
} from 'lucide-react';
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
    <div className="space-y-6 p-6 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 rounded-2xl">

      {/* Header with Refresh - Dark Theme */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-white">Server Performance Overview</h2>
        <div className="flex items-center gap-4">
          <label className="flex items-center gap-2 text-sm text-gray-300">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded border-gray-600 bg-slate-700 text-cyan-500 focus:ring-cyan-500"
            />
            Auto-refresh (30s)
          </label>
          <button
            onClick={() => refetch()}
            className="p-2 hover:bg-slate-700/50 rounded-lg transition-colors backdrop-blur-sm"
            title="Refresh now"
          >
            <RefreshCw className="w-5 h-5 text-gray-300" />
          </button>
        </div>
      </div>

      {/* Last Updated - Dark Theme */}
      <div className="text-sm text-gray-400">
        Last updated: {m?.timestamp ? new Date(m.timestamp).toLocaleString() : 'N/A'}
      </div>

      {/* Metric Cards Grid - Dark Glassmorphism */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-2 gap-4">
        {/* CPU Usage - Dark Glass with Chart */}
        <div className="relative overflow-hidden rounded-xl bg-slate-800/40 backdrop-blur-xl border border-slate-700/50 p-5">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-cyan-500/20 rounded-lg flex items-center justify-center">
                <Cpu className="w-4 h-4 text-cyan-400" />
              </div>
              <h3 className="text-sm font-medium text-gray-300">CPU Usage</h3>
            </div>
            <button className="p-1.5 hover:bg-slate-700/50 rounded-lg transition-colors">
              <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </button>
          </div>
          
          {/* CPU Chart Area */}
          <div className="h-40 -mx-2">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={series.slice(-20)} margin={{ top: 5, right: 5, left: 5, bottom: 5 }}>
                <defs>
                  <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#06B6D4" stopOpacity={0.8}/>
                    <stop offset="95%" stopColor="#06B6D4" stopOpacity={0.1}/>
                  </linearGradient>
                </defs>
                <XAxis dataKey="timestamp" hide />
                <Tooltip 
                  contentStyle={{ 
                    background: 'rgba(30, 41, 59, 0.95)', 
                    border: '1px solid rgba(71, 85, 105, 0.5)', 
                    borderRadius: '8px', 
                    backdropFilter: 'blur(10px)',
                    color: '#fff'
                  }}
                  labelStyle={{ color: '#94a3b8' }}
                  itemStyle={{ color: '#06B6D4' }}
                  formatter={(value: any) => [`${value}%`, 'CPU']} 
                />
                <Area 
                  type="monotone" 
                  dataKey="cpu_usage" 
                  stroke="#06B6D4" 
                  strokeWidth={2}
                  fill="url(#cpuGradient)"
                  dot={false}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
          
          <div className="mt-2 flex items-center justify-between text-xs">
            <span className="text-gray-400">15:38</span>
            <span className="text-gray-400">16:11</span>
            <span className="text-gray-400">16:33</span>
          </div>
        </div>

        {/* System Load Average - Dark Glass with Circular Gauge */}
        <div className="relative overflow-hidden rounded-xl bg-slate-800/40 backdrop-blur-xl border border-slate-700/50 p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-gray-300">System Load Average</h3>
            <button className="p-1.5 hover:bg-slate-700/50 rounded-lg transition-colors">
              <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </button>
          </div>
          
          {/* Circular Gauge */}
          <div className="flex items-center justify-center py-4">
            <div className="relative w-40 h-40">
              <svg className="transform -rotate-90 w-40 h-40">
                {/* Background Circle */}
                <circle
                  cx="80"
                  cy="80"
                  r="70"
                  stroke="rgba(71, 85, 105, 0.3)"
                  strokeWidth="12"
                  fill="none"
                />
                {/* Progress Circle - Cyan to Green gradient */}
                <circle
                  cx="80"
                  cy="80"
                  r="70"
                  stroke="url(#loadGradient)"
                  strokeWidth="12"
                  fill="none"
                  strokeLinecap="round"
                  strokeDasharray={`${2 * Math.PI * 70 * 0.75}`}
                  strokeDashoffset={`${2 * Math.PI * 70 * 0.75 * (1 - (cpu ?? 0) / 100)}`}
                  className="transition-all duration-1000"
                />
                <defs>
                  <linearGradient id="loadGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stopColor="#06B6D4" />
                    <stop offset="100%" stopColor="#10B981" />
                  </linearGradient>
                </defs>
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-4xl font-bold text-white">
                  {typeof cpu === 'number' ? (cpu / 100).toFixed(2) : '0.00'}
                </span>
              </div>
            </div>
          </div>
        </div>

      </div>

      {/* Second Row - Memory, Network, Disk IOPS */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Memory Utilization */}
        <div className="relative overflow-hidden rounded-xl bg-slate-800/40 backdrop-blur-xl border border-slate-700/50 p-5">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-purple-500/20 rounded-lg flex items-center justify-center">
                <Activity className="w-4 h-4 text-purple-400" />
              </div>
              <h3 className="text-sm font-medium text-gray-300">Memory Utilization</h3>
            </div>
            <button className="p-1.5 hover:bg-slate-700/50 rounded-lg transition-colors">
              <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </button>
          </div>
          
          <div className="h-32 -mx-2">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={series.slice(-20)} margin={{ top: 5, right: 5, left: 5, bottom: 5 }}>
                <defs>
                  <linearGradient id="memGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#A855F7" stopOpacity={0.8}/>
                    <stop offset="95%" stopColor="#A855F7" stopOpacity={0.1}/>
                  </linearGradient>
                </defs>
                <XAxis dataKey="timestamp" hide />
                <Tooltip 
                  contentStyle={{ background: 'rgba(30, 41, 59, 0.95)', border: '1px solid rgba(71, 85, 105, 0.5)', borderRadius: '8px', backdropFilter: 'blur(10px)', color: '#fff' }}
                  formatter={(value: any) => [`${value}%`, 'Memory']} 
                />
                <Area type="monotone" dataKey="memory_usage" stroke="#A855F7" strokeWidth={2} fill="url(#memGradient)" dot={false} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Network In/Out */}
        <div className="relative overflow-hidden rounded-xl bg-slate-800/40 backdrop-blur-xl border border-slate-700/50 p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-gray-300">Network In/Out</h3>
            <button className="p-1.5 hover:bg-slate-700/50 rounded-lg transition-colors">
              <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </button>
          </div>
          
          <div className="h-32 flex items-end gap-1">
            {series.slice(-12).map((item, idx) => {
              const maxVal = Math.max(...series.map(s => Math.max(s.network_in || 0, s.network_out || 0)));
              const inHeight = maxVal > 0 ? ((item.network_in || 0) / maxVal) * 100 : 0;
              const outHeight = maxVal > 0 ? ((item.network_out || 0) / maxVal) * 100 : 0;
              return (
                <div key={idx} className="flex-1 flex gap-0.5 items-end">
                  <div className="flex-1 bg-gradient-to-t from-cyan-500 to-cyan-400 rounded-t" style={{ height: `${inHeight}%`, minHeight: '4px' }} />
                  <div className="flex-1 bg-gradient-to-t from-green-500 to-green-400 rounded-t" style={{ height: `${outHeight}%`, minHeight: '4px' }} />
                </div>
              );
            })}
          </div>
        </div>

        {/* Disk IOPS */}
        <div className="relative overflow-hidden rounded-xl bg-slate-800/40 backdrop-blur-xl border border-slate-700/50 p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-gray-300">Disk IOPS</h3>
            <button className="p-1.5 hover:bg-slate-700/50 rounded-lg transition-colors">
              <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </button>
          </div>
          
          <div className="h-32 flex items-end gap-1">
            {series.slice(-12).map((item, idx) => {
              const maxDisk = Math.max(...series.map(s => s.disk_usage || 0));
              const height = maxDisk > 0 ? ((item.disk_usage || 0) / maxDisk) * 100 : 0;
              return (
                <div key={idx} className="flex-1 bg-gradient-to-t from-orange-500 to-orange-400 rounded-t" style={{ height: `${height}%`, minHeight: '4px' }} />
              );
            })}
          </div>
        </div>
      </div>


export default HostDashboard;    </div>
  );
};

export default HostDashboard;
