import React, { useState, useEffect } from 'react';
import { 
  Cpu, 
  HardDrive, 
  Activity, 
  Network, 
  Clock, 
  TrendingUp,
  RefreshCw,
  AlertCircle,
  Server,
  MapPin,
  Calendar
} from 'lucide-react';
import useHostMetrics from '../../hooks/useHostMetrics';
import {
  ResponsiveContainer,
  XAxis,
  Tooltip,
  AreaChart,
  Area,
  LineChart,
  Line
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
  const [hostInfo, setHostInfo] = useState<any>(null);
  const { metrics, series, loading, error, refetch } = useHostMetrics(hostId, autoRefresh);
  const m = metrics as MetricData | null;
  
  // Fetch host information
  useEffect(() => {
    const fetchHostInfo = async () => {
      try {
        const response = await fetch(`http://localhost:8080/api/v1/hosts/${hostId}`);
        if (response.ok) {
          const data = await response.json();
          setHostInfo(data);
        }
      } catch (err) {
        console.error('Failed to fetch host info:', err);
      }
    };
    if (hostId) {
      fetchHostInfo();
    }
  }, [hostId]);
  
  const cpu = m?.cpu_usage;
  const memory = m?.memory_usage;
  const memory_used = m?.memory_used ?? 0;
  const memory_total = m?.memory_total ?? 0;
  const disk = m?.disk_usage;
  const disk_used = m?.disk_used ?? 0;
  const uptimeVal = m?.uptime ?? null;
  const loadAvg = m?.load_average ?? '';
  const netIn = m?.network_in !== undefined && m.network_in !== null ? m.network_in : 0;
  const netOut = m?.network_out !== undefined && m.network_out !== null ? m.network_out : 0;

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (days > 0) return `${days}h ${hours}h ${minutes}m`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600 mx-auto mb-3"></div>
          <p className="text-sm text-gray-600">Loading metrics...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-xl p-4">
        <div className="flex items-center gap-2">
          <AlertCircle className="w-5 h-5 text-red-600" />
          <div>
            <h3 className="text-sm font-semibold text-red-900">Failed to Load Metrics</h3>
            <p className="text-xs text-red-700 mt-0.5">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-6 text-center">
        <Clock className="w-10 h-10 text-yellow-600 mx-auto mb-2" />
        <h3 className="text-sm font-semibold text-yellow-900 mb-1">No Metrics Yet</h3>
        <p className="text-xs text-yellow-700">Metrics collection in progress...</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">

      {/* Modern Compact Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-cyan-500 rounded-xl flex items-center justify-center shadow-lg">
            <TrendingUp className="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-gray-900">Performance Overview</h2>
            <p className="text-xs text-gray-500">
              {m?.timestamp ? new Date(m.timestamp).toLocaleTimeString() : 'N/A'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-1.5 px-3 py-1.5 bg-white rounded-lg border border-gray-200 shadow-sm text-xs text-gray-700 hover:bg-gray-50 transition-colors">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="w-3.5 h-3.5 rounded border-gray-300 text-blue-600"
            />
            Auto (30s)
          </label>
          <button
            onClick={() => refetch()}
            className="p-2 bg-white hover:bg-gray-50 rounded-lg border border-gray-200 shadow-sm transition-colors"
            title="Refresh now"
          >
            <RefreshCw className="w-4 h-4 text-gray-600" />
          </button>
        </div>
      </div>

      {/* Compact Metric Cards - Paytm Style */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3">
        
        {/* CPU Usage */}
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-blue-50 to-cyan-50 border border-blue-100 p-4 hover:shadow-lg transition-shadow">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-gray-600 uppercase tracking-wider">CPU Usage</span>
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-cyan-500 rounded-lg flex items-center justify-center shadow-md">
              <Cpu className="w-4 h-4 text-white" />
            </div>
          </div>
          
          <div className="flex items-end justify-center h-20">
            <div className="relative w-20 h-20">
              <svg className="transform -rotate-90 w-20 h-20">
                <circle cx="40" cy="40" r="32" stroke="#E0E7FF" strokeWidth="5" fill="none" />
                <circle
                  cx="40" cy="40" r="32"
                  stroke="url(#cpuGrad)"
                  strokeWidth="5"
                  fill="none"
                  strokeLinecap="round"
                  strokeDasharray={`${2 * Math.PI * 32}`}
                  strokeDashoffset={`${2 * Math.PI * 32 * (1 - (cpu ?? 0) / 100)}`}
                  className="transition-all duration-700"
                />
                <defs>
                  <linearGradient id="cpuGrad">
                    <stop offset="0%" stopColor="#3B82F6" />
                    <stop offset="100%" stopColor="#06B6D4" />
                  </linearGradient>
                </defs>
              </svg>
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-xl font-bold bg-gradient-to-br from-blue-600 to-cyan-600 bg-clip-text text-transparent">
                  {typeof cpu === 'number' ? `${cpu.toFixed(1)}%` : 'N/A'}
                </span>
              </div>
            </div>
          </div>
          <p className="text-[10px] text-center text-gray-500 mt-1">Time</p>
        </div>

        {/* Memory Usage */}
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-emerald-50 to-teal-50 border border-emerald-100 p-4 hover:shadow-lg transition-shadow">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-gray-600 uppercase tracking-wider">Memory</span>
            <div className="w-8 h-8 bg-gradient-to-br from-emerald-500 to-teal-500 rounded-lg flex items-center justify-center shadow-md">
              <Activity className="w-4 h-4 text-white" />
            </div>
          </div>
          
          <div className="h-16 mb-2">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={series.slice(-12)} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="memGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#10B981" stopOpacity={0.5}/>
                    <stop offset="100%" stopColor="#14B8A6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <XAxis dataKey="timestamp" hide />
                <Tooltip 
                  contentStyle={{ background: 'white', border: 'none', borderRadius: '8px', fontSize: '11px', padding: '6px 10px', boxShadow: '0 4px 6px rgba(0,0,0,0.1)' }}
                  formatter={(value: any) => [`${value}%`, 'Memory']} 
                />
                <Area type="monotone" dataKey="memory_usage" stroke="#10B981" strokeWidth={2} fill="url(#memGrad)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
          
          <div className="flex items-center justify-between">
            <div>
              <p className="text-[10px] text-gray-500">3 Day</p>
              <p className="text-sm font-bold bg-gradient-to-r from-emerald-600 to-teal-600 bg-clip-text text-transparent">
                {typeof memory === 'number' ? `${memory.toFixed(1)}%` : 'N/A'}
              </p>
            </div>
            <div className="text-right">
              <p className="text-[10px] text-gray-500">Available</p>
              <p className="text-sm font-bold text-emerald-600">
                {(memory_total - memory_used).toFixed(0)} MB
              </p>
            </div>
          </div>
        </div>

        {/* Disk Usage */}
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-purple-50 to-pink-50 border border-purple-100 p-4 hover:shadow-lg transition-shadow">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-gray-600 uppercase tracking-wider">Disk</span>
            <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center shadow-md">
              <HardDrive className="w-4 h-4 text-white" />
            </div>
          </div>
          
          <div className="flex items-end justify-center h-20">
            <div className="relative w-20 h-20">
              <svg className="transform -rotate-90 w-20 h-20">
                <circle cx="40" cy="40" r="32" stroke="#F3E8FF" strokeWidth="5" fill="none" />
                <circle
                  cx="40" cy="40" r="32"
                  stroke="url(#diskGrad)"
                  strokeWidth="5"
                  fill="none"
                  strokeLinecap="round"
                  strokeDasharray={`${2 * Math.PI * 32}`}
                  strokeDashoffset={`${2 * Math.PI * 32 * (1 - (disk ?? 0) / 100)}`}
                  className="transition-all duration-700"
                />
                <defs>
                  <linearGradient id="diskGrad">
                    <stop offset="0%" stopColor="#A855F7" />
                    <stop offset="100%" stopColor="#EC4899" />
                  </linearGradient>
                </defs>
              </svg>
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-xl font-bold bg-gradient-to-br from-purple-600 to-pink-600 bg-clip-text text-transparent">
                  {typeof disk === 'number' ? `${disk.toFixed(0)}%` : 'N/A'}
                </span>
              </div>
            </div>
          </div>
          <p className="text-[10px] text-center text-gray-500 mt-1">{disk_used.toFixed(0)}GB</p>
        </div>

        {/* Network I/O */}
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-orange-50 to-red-50 border border-orange-100 p-4 hover:shadow-lg transition-shadow">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-gray-600 uppercase tracking-wider">Network I/O</span>
            <div className="w-8 h-8 bg-gradient-to-br from-orange-500 to-red-500 rounded-lg flex items-center justify-center shadow-md">
              <Network className="w-4 h-4 text-white" />
            </div>
          </div>
          
          <div className="space-y-2.5">
            <div>
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-1">
                  <div className="w-1.5 h-1.5 bg-orange-500 rounded-full" />
                  <span className="text-[10px] text-gray-600 uppercase">Incoming</span>
                </div>
                <span className="text-xs font-bold text-gray-900">{netIn.toFixed(1)} MB</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-1.5">
                <div
                  className="h-1.5 rounded-full bg-gradient-to-r from-orange-500 to-red-500"
                  style={{ width: `${Math.min((netIn / 2574.47) * 100, 100)}%` }}
                />
              </div>
            </div>
            
            <div>
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-1">
                  <div className="w-1.5 h-1.5 bg-pink-500 rounded-full" />
                  <span className="text-[10px] text-gray-600 uppercase">Outgoing</span>
                </div>
                <span className="text-xs font-bold text-gray-900">{netOut.toFixed(1)} MB</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-1.5">
                <div
                  className="h-1.5 rounded-full bg-gradient-to-r from-red-500 to-pink-500"
                  style={{ width: `${Math.min((netOut / 737.0) * 100, 100)}%` }}
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Performance Trend & Uptime - Compact */}
      <div className="grid grid-cols-1 lg:grid-cols-6 gap-3">
        
        {/* System Uptime */}
        <div className="lg:col-span-2 rounded-2xl bg-gradient-to-br from-indigo-50 to-blue-50 border border-indigo-100 p-4 hover:shadow-lg transition-shadow">
          <div className="flex items-center gap-2 mb-3">
            <div className="w-7 h-7 bg-gradient-to-br from-indigo-500 to-blue-500 rounded-lg flex items-center justify-center">
              <Clock className="w-3.5 h-3.5 text-white" />
            </div>
            <span className="text-xs font-semibold text-gray-600 uppercase tracking-wider">System Uptime</span>
          </div>
          
          <div className="text-center">
            <div className="text-3xl font-bold bg-gradient-to-r from-indigo-600 to-blue-600 bg-clip-text text-transparent mb-1">
              {uptimeVal !== null ? formatUptime(uptimeVal) : 'N/A'}
            </div>
            <div className="text-xs text-gray-500">
              Load: <span className="font-semibold text-gray-700">{loadAvg || 'N/A'}</span>
            </div>
          </div>
        </div>

        {/* Performance Trend */}
        <div className="lg:col-span-4 rounded-2xl bg-gradient-to-br from-cyan-50 to-blue-50 border border-cyan-100 p-4 hover:shadow-lg transition-shadow">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <div className="w-7 h-7 bg-gradient-to-br from-cyan-500 to-blue-500 rounded-lg flex items-center justify-center">
                <TrendingUp className="w-3.5 h-3.5 text-white" />
              </div>
              <span className="text-xs font-semibold text-gray-600 uppercase tracking-wider">Performance Trend</span>
            </div>
          </div>
          
          <div className="h-16">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={series.slice(-20)} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="trendGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#06B6D4" stopOpacity={0.3}/>
                    <stop offset="100%" stopColor="#3B82F6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <XAxis dataKey="timestamp" hide />
                <Tooltip 
                  contentStyle={{ background: 'white', border: 'none', borderRadius: '8px', fontSize: '11px', padding: '6px 10px', boxShadow: '0 4px 6px rgba(0,0,0,0.1)' }}
                />
                <Line type="monotone" dataKey="cpu_usage" stroke="#06B6D4" strokeWidth={2} dot={false} name="CPU" />
                <Line type="monotone" dataKey="memory_usage" stroke="#3B82F6" strokeWidth={2} dot={false} name="Memory" />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Network Statistics - Full Width */}
      <div className="rounded-2xl bg-gradient-to-br from-teal-50 to-emerald-50 border border-teal-100 p-4 hover:shadow-lg transition-shadow">
        <div className="flex items-center gap-2 mb-4">
          <div className="w-7 h-7 bg-gradient-to-br from-teal-500 to-emerald-500 rounded-lg flex items-center justify-center">
            <Network className="w-3.5 h-3.5 text-white" />
          </div>
          <h3 className="text-xs font-semibold text-gray-700 uppercase tracking-wider">Network I/O</h3>
          <span className="text-[10px] text-gray-500">Real-time network statistics</span>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-white/60 backdrop-blur-sm rounded-xl p-4 border border-teal-100">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-teal-500 rounded-full animate-pulse" />
                <span className="text-xs font-semibold text-gray-700">Incoming</span>
              </div>
              <span className="text-sm font-bold bg-gradient-to-r from-teal-600 to-emerald-600 bg-clip-text text-transparent">
                {netIn.toFixed(2)} MB
              </span>
            </div>
            <p className="text-[10px] text-gray-500">Total received</p>
          </div>
          
          <div className="bg-white/60 backdrop-blur-sm rounded-xl p-4 border border-teal-100">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
                <span className="text-xs font-semibold text-gray-700">Outgoing</span>
              </div>
              <span className="text-sm font-bold bg-gradient-to-r from-emerald-600 to-teal-600 bg-clip-text text-transparent">
                {netOut.toFixed(2)} MB
              </span>
            </div>
            <p className="text-[10px] text-gray-500">Total sent</p>
          </div>
        </div>
      </div>

      {/* Host Information - Paytm Style Card */}
      <div className="rounded-2xl bg-gradient-to-br from-slate-50 to-gray-50 border border-slate-200 overflow-hidden hover:shadow-lg transition-shadow">
        <div className="bg-gradient-to-r from-slate-700 to-gray-700 p-3">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 bg-white/20 backdrop-blur-sm rounded-lg flex items-center justify-center">
              <Server className="w-3.5 h-3.5 text-white" />
            </div>
            <h3 className="text-sm font-bold text-white">Host Information</h3>
          </div>
        </div>
        
        <div className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="bg-white rounded-xl p-3 border border-slate-200">
              <div className="flex items-center gap-2 mb-1.5">
                <Server className="w-3.5 h-3.5 text-blue-600" />
                <span className="text-[10px] font-semibold text-gray-500 uppercase tracking-wide">Hostname</span>
              </div>
              <p className="text-sm font-bold text-gray-900">{hostInfo?.hostname || 'N/A'}</p>
            </div>
            
            <div className="bg-white rounded-xl p-3 border border-slate-200">
              <div className="flex items-center gap-2 mb-1.5">
                <MapPin className="w-3.5 h-3.5 text-green-600" />
                <span className="text-[10px] font-semibold text-gray-500 uppercase tracking-wide">IP Address</span>
              </div>
              <p className="text-sm font-bold text-gray-900">{hostInfo?.ip || 'N/A'}</p>
            </div>
            
            <div className="bg-white rounded-xl p-3 border border-slate-200">
              <div className="flex items-center gap-2 mb-1.5">
                <Activity className="w-3.5 h-3.5 text-purple-600" />
                <span className="text-[10px] font-semibold text-gray-500 uppercase tracking-wide">Operating System</span>
              </div>
              <p className="text-sm font-bold text-gray-900">{hostInfo?.os || 'linux'}</p>
            </div>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-3">
            <div className="bg-white rounded-xl p-3 border border-slate-200">
              <div className="flex items-center gap-2 mb-1.5">
                <Server className="w-3.5 h-3.5 text-orange-600" />
                <span className="text-[10px] font-semibold text-gray-500 uppercase tracking-wide">Group</span>
              </div>
              <p className="text-sm font-bold text-gray-900">{hostInfo?.group || 'auto-discovered'}</p>
            </div>
            
            <div className="bg-white rounded-xl p-3 border border-slate-200">
              <div className="flex items-center gap-2 mb-1.5">
                <Clock className="w-3.5 h-3.5 text-cyan-600" />
                <span className="text-[10px] font-semibold text-gray-500 uppercase tracking-wide">Last Seen</span>
              </div>
              <p className="text-sm font-bold text-gray-900">
                {hostInfo?.last_seen ? new Date(hostInfo.last_seen).toLocaleString() : 'N/A'}
              </p>
            </div>
            
            <div className="bg-white rounded-xl p-3 border border-slate-200">
              <div className="flex items-center gap-2 mb-1.5">
                <Calendar className="w-3.5 h-3.5 text-pink-600" />
                <span className="text-[10px] font-semibold text-gray-500 uppercase tracking-wide">Created At</span>
              </div>
              <p className="text-sm font-bold text-gray-900">
                {hostInfo?.created_at ? new Date(hostInfo.created_at).toLocaleString() : 'N/A'}
              </p>
            </div>
          </div>
          
          {hostInfo?.description && (
            <div className="mt-3 bg-blue-50 border border-blue-100 rounded-xl p-3">
              <p className="text-[10px] font-semibold text-blue-600 uppercase tracking-wide mb-1">Description</p>
              <p className="text-xs text-gray-700">{hostInfo.description}</p>
            </div>
          )}
          
          <div className="mt-3 bg-gray-50 border border-gray-200 rounded-xl p-3">
            <p className="text-[10px] font-semibold text-gray-600 uppercase tracking-wide mb-1">Tenant ID</p>
            <p className="text-xs font-mono text-gray-900">{hostInfo?.tenant_id || '2'}</p>
          </div>
        </div>
      </div>

    </div>
  );
};

export default HostDashboard;
