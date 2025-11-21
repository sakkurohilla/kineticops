import React, { useState, useEffect, useRef } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import Card from '../../components/common/Card';
import Button from '../../components/common/Button';
import Badge from '../../components/common/Badge';
import { 
  Server, 
  Activity, 
  AlertTriangle, 
  TrendingUp,
  TrendingDown,
  Cpu,
  Shield,
  ChevronRight,
  Database,
  X,
  RefreshCw
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import hostService from '../../services/api/hostService';
import apiClient from '../../services/api/client';
import authService from '../../services/auth/authService';
// import { useMetrics } from '../../hooks/useMetrics';
import useWebsocket from '../../hooks/useWebsocket';
import { formatTimestamp } from '../../utils/dateUtils';

interface DashboardStats {
  totalHosts: number;
  onlineHosts: number;
  offlineHosts: number;
  criticalAlerts: number;
  warningAlerts: number;
  avgCpuUsage: number | null;
  avgMemoryUsage: number | null;
  avgDiskUsage: number | null;
  systemHealth: number;
}

interface Host {
  id: number;
  hostname?: string;
  ip: string;
  agent_status?: string;
  last_seen?: string;
  os?: string;
}

interface Alert {
  id: number;
  host_id: number;
  host_name?: string;
  message?: string;
  severity?: string;
  created_at: string;
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  // Don't initialize with zeros - causes flash on refresh
  const [stats, setStats] = useState<DashboardStats>({
    totalHosts: 0,
    onlineHosts: 0,
    offlineHosts: 0,
    criticalAlerts: 0,
    warningAlerts: 0,
    avgCpuUsage: null,
    avgMemoryUsage: null,
    avgDiskUsage: null,
    systemHealth: 100,
  });
  const [hosts, setHosts] = useState<Host[]>([]);
  
  // Keep ref in sync with state
  useEffect(() => {
    hostsRef.current = hosts;
  }, [hosts]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  
  // Keep ref in sync with state
  useEffect(() => {
    alertsRef.current = alerts;
  }, [alerts]);
  const [hostMetrics, setHostMetrics] = useState<Record<number, any>>({});
  const [selectedHostId, setSelectedHostId] = useState<number | null>(null);
  const [selectedHostMetrics, setSelectedHostMetrics] = useState<any | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);
  // Derived values for selected host analytics (compute once)
  const selectedMemPercent = (() => {
    const m = selectedHostMetrics;
    if (!m) return null;
    if ((m.memory_total && m.memory_used) || (m.memory_total_bytes && m.memory_used_bytes)) {
      const total = m.memory_total || m.memory_total_bytes;
      const used = m.memory_used || m.memory_used_bytes;
      if (total && used) return (used / total) * 100;
    }
    return m.memory_usage ?? null;
  })();

  const selectedUsedTotalDisplay = (() => {
    const m = selectedHostMetrics;
    if (!m) return '—';
    if ((m.memory_total && m.memory_used) || (m.memory_total_bytes && m.memory_used_bytes)) {
      const total = m.memory_total || m.memory_total_bytes;
      const used = m.memory_used || m.memory_used_bytes;
      const toMB = (v: number) => (v > 1024*1024 ? `${(v / (1024*1024)).toFixed(1)} MB` : `${(v / 1024).toFixed(1)} KB`);
      return `${toMB(used)} / ${toMB(total)}`;
    }
    if (m.memory_total) {
      const used = (m.memory_usage || 0) / 100 * m.memory_total;
      return `${used.toFixed(1)} MB / ${m.memory_total.toFixed(1)} MB`;
    }
    return '—';
  })();
  
  // const { data: metricsData } = useMetrics('1h', undefined, true);

  // Keep a ref of last hosts list to avoid refetching when unchanged
  const lastHostsRef = useRef<number[] | null>(null);
  // pending metrics buffer for debounce
  const pendingMetricsRef = useRef<Record<number, any>>({});
  const debounceTimerRef = useRef<number | null>(null);
  const DEBOUNCE_MS = 200;
  
  // Refs to always access latest state (prevents stale closure in WebSocket callback)
  const hostsRef = useRef<Host[]>([]);
  const alertsRef = useRef<Alert[]>([]);

  // Real-time updates via WebSocket - buffer updates and apply debounced
  useWebsocket((payload: any) => {
    if (!payload || !payload?.host_id) return;

    // Normalize minimal payload
    const normalized = {
      cpu_usage: payload.cpu_usage || 0,
      memory_usage: payload.memory_usage ?? null,
      memory_total: payload.memory_total ?? payload.memory_total_bytes ?? null,
      memory_used: payload.memory_used ?? payload.memory_used_bytes ?? null,
      memory_available: payload.memory_available ?? payload.memory_available_bytes ?? null,
      disk_usage: payload.disk_usage ?? null,
      disk_total: payload.disk_total ?? payload.disk_total_bytes ?? null,
      disk_used: payload.disk_used ?? payload.disk_used_bytes ?? null,
      network_in: payload.network_in || 0,
      network_out: payload.network_out || 0,
      timestamp: payload.timestamp || new Date().toISOString(),
    } as any;

    // Buffer into pending metrics
    pendingMetricsRef.current = {
      ...pendingMetricsRef.current,
      [payload.host_id]: {
        ...(pendingMetricsRef.current[payload.host_id] || {}),
        ...normalized,
      }
    };

    // Schedule debounce flush
    if (debounceTimerRef.current) {
      window.clearTimeout(debounceTimerRef.current);
    }
    debounceTimerRef.current = window.setTimeout(() => {
      const toApply = { ...pendingMetricsRef.current };
      pendingMetricsRef.current = {};
      debounceTimerRef.current = null;

      // merge into hostMetrics and recompute once
      setHostMetrics(prev => {
        const merged = { ...prev, ...toApply };
        // Only recompute if we have hosts loaded (avoid resetting stats on initial WS update)
        // Use refs to get latest values (prevents stale closure)
        const currentHosts = hostsRef.current;
        const currentAlerts = alertsRef.current;
        if (currentHosts && currentHosts.length > 0) {
          recomputeStats(currentHosts, merged, currentAlerts);
        }
        return merged;
      });
    }, DEBOUNCE_MS) as unknown as number;
  });

  useEffect(() => {
    // Wait for auth token before fetching to avoid 403 errors
    const token = authService.getToken();
    if (token) {
      console.log('[Dashboard] Auth token ready, fetching dashboard data');
      fetchDashboardData();
    } else {
      console.warn('[Dashboard] No auth token yet, will retry until available');
      // Retry with longer interval until token is available (after login)
      let retryCount = 0;
      const maxRetries = 10; // Max 10 seconds waiting
      const timer = setInterval(() => {
        const retryToken = authService.getToken();
        if (retryToken) {
          console.log('[Dashboard] Token acquired on retry, fetching dashboard data');
          clearInterval(timer);
          fetchDashboardData();
        } else {
          retryCount++;
          if (retryCount >= maxRetries) {
            console.error('[Dashboard] Still no token after', maxRetries, 'retries');
            clearInterval(timer);
          }
        }
      }, 1000); // Check every second
      return () => clearInterval(timer);
    }
    // Removed 30s interval - rely on WebSocket for real-time updates to prevent flicker
    // Manual refresh available via Refresh button
  }, []);

  const fetchDashboardData = async () => {
    try {
      // Never set loading to prevent values flashing to 0
      setIsRefreshing(true);

      // Fetch hosts - try direct API call if service fails
      let hostsData;
      try {
        hostsData = await hostService.getAllHosts();
      } catch (err) {
        // Direct API call without auth for auto-discovered hosts
        const response = await fetch('http://localhost:8080/api/v1/hosts', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json'
          }
        });
        if (response.ok) {
          hostsData = await response.json();
        } else {
          hostsData = [];
        }
      }
      
      // Always update hosts state (even if empty)
      console.log('[Dashboard] Fetched hosts:', hostsData?.length || 0, 'hosts', hostsData);
      setHosts(hostsData || []);

      // Avoid re-fetching per-host metrics if host list didn't change
      const hostIds = (hostsData || []).map((h: any) => h.id).sort();
      const last = lastHostsRef.current;
      const sameHosts = last && JSON.stringify(last) === JSON.stringify(hostIds);
      lastHostsRef.current = hostIds;

      // Fetch alerts
      try {
        const alertsData = await apiClient.get('/alerts?limit=10');
        setAlerts(Array.isArray(alertsData) ? alertsData : []);
      } catch (err) {
        setAlerts([]);
      }

      // Calculate stats
      const totalHosts = hostsData.length;
      const onlineHosts = hostsData.filter((h: any) => h.agent_status === 'online').length;
      const offlineHosts = totalHosts - onlineHosts;
      
      let metricsResults: Array<{ hostId: number; metrics: any }>; 
      if (sameHosts && Object.keys(hostMetrics).length > 0) {
        // Host list unchanged and we already have metrics from websocket/cache - reuse
        metricsResults = (hostsData || []).map((h: any) => ({ hostId: h.id, metrics: hostMetrics[h.id] || null }));
      } else {
        // Try bulk endpoint first (new backend route)
          try {
          const allLatest: any = await apiClient.get(`/hosts/metrics/latest/all?limit=${hostIds.length}`);
          console.log('[Dashboard] Bulk metrics API response:', allLatest);
          // support both array and map shapes returned by backend
          if (Array.isArray(allLatest)) {
            const map: Record<number, any> = {};
            (allLatest as any[]).forEach((m: any) => { if (m && m.host_id) map[m.host_id] = m; });
            metricsResults = (hostsData || []).map((h: any) => ({ hostId: h.id, metrics: map[h.id] || null }));
          } else {
            metricsResults = (hostsData || []).map((h: any) => ({ hostId: h.id, metrics: allLatest[h.id] || null }));
          }
        } catch (e) {
          console.warn('[Dashboard] Bulk metrics failed, falling back to per-host fetch:', e);
          // Fallback: per-host fetch with timeout
          const timeoutMs = 1500;
          const perHost = (hostsData || []).map((host: any) => {
            const p = (async () => {
              try {
                const metrics = await hostService.getLatestMetrics(host.id);
                console.log('[Dashboard] Per-host metrics for', host.id, ':', metrics);
                return { hostId: host.id, metrics };
              } catch (err) {
                try {
                  const response = await fetch(`http://localhost:8080/api/v1/hosts/${host.id}/metrics/latest`);
                  if (response.ok) {
                    const metrics = await response.json();
                    return { hostId: host.id, metrics };
                  }
                } catch (e) {
                  // ignore
                }
                return { hostId: host.id, metrics: null };
              }
            })();
            // apply timeout
            return Promise.race([
              p,
              new Promise(resolve => setTimeout(() => resolve({ hostId: host.id, metrics: null }), timeoutMs))
            ] as any) as Promise<{ hostId: number; metrics: any }>;
          });

          metricsResults = await Promise.all(perHost);
        }
      }
      console.log('[Dashboard] Got metricsResults:', metricsResults?.length || 0, 'results', metricsResults);
      
      const metricsMap: Record<number, any> = {};
      let totalCpu = 0, totalMemory = 0, totalDisk = 0, validMetrics = 0;

      metricsResults.forEach(({ hostId, metrics }) => {
        console.log('[Dashboard] Processing metric for host', hostId, ':', metrics);
        // Skip placeholder metric frames (early websocket frames with no real values)
        if (metrics && !metrics._placeholder) {
          // compute memory percent from totals when available
          let memPercent = metrics.memory_usage;
          if ((metrics.memory_total && metrics.memory_used) || (metrics.memory_total_bytes && metrics.memory_used_bytes)) {
            const total = metrics.memory_total || metrics.memory_total_bytes;
            const used = metrics.memory_used || metrics.memory_used_bytes;
            if (total && used) memPercent = (used / total) * 100;
          }

          // disk percent fallback
          let diskPercent = metrics.disk_usage;
          if ((metrics.disk_total && metrics.disk_used) || (metrics.disk_total_bytes && metrics.disk_used_bytes)) {
            const dtotal = metrics.disk_total || metrics.disk_total_bytes;
            const dused = metrics.disk_used || metrics.disk_used_bytes;
            if (dtotal && dused) diskPercent = (dused / dtotal) * 100;
          }

          const normalized = {
            ...metrics,
            memory_usage: typeof memPercent === 'number' ? memPercent : (metrics.memory_usage || 0),
            disk_usage: typeof diskPercent === 'number' ? diskPercent : (metrics.disk_usage || 0),
          };

          metricsMap[hostId] = normalized;
          totalCpu += normalized.cpu_usage || 0;
          totalMemory += normalized.memory_usage || 0;
          totalDisk += normalized.disk_usage || 0;
          validMetrics++;
        }
      });

      // Merge metrics into existing hostMetrics state to preserve websocket updates
      setHostMetrics(prev => ({ ...prev, ...metricsMap }));

      // Calculate new averages; if none valid, keep previous non-null values
      const newAvgCpuUsage = validMetrics > 0 ? totalCpu / validMetrics : null;
      const newAvgMemoryUsage = validMetrics > 0 ? totalMemory / validMetrics : null;
      const newAvgDiskUsage = validMetrics > 0 ? totalDisk / validMetrics : null;

      console.log('[Dashboard] Initial fetch complete: validMetrics=' + validMetrics + 
        ' cpu=' + (newAvgCpuUsage?.toFixed(1) || 'null') + '%' +
        ' mem=' + (newAvgMemoryUsage?.toFixed(1) || 'null') + '%' +
        ' disk=' + (newAvgDiskUsage?.toFixed(1) || 'null') + '%');

      const systemHealth = totalHosts > 0 ? Math.round((onlineHosts / totalHosts) * 100) : 100;

      // Functional update so we can preserve last non-null averages on empty responses
      setStats(prev => {
        const cpuFinal = newAvgCpuUsage !== null ? newAvgCpuUsage : prev.avgCpuUsage;
        const memFinal = newAvgMemoryUsage !== null ? newAvgMemoryUsage : prev.avgMemoryUsage;
        const diskFinal = newAvgDiskUsage !== null ? newAvgDiskUsage : prev.avgDiskUsage;
        
        const newStats = {
          totalHosts,
            onlineHosts,
            offlineHosts,
            criticalAlerts: alerts.filter(a => a.severity === 'critical').length,
            warningAlerts: alerts.filter(a => a.severity === 'warning' || a.severity === 'high').length,
            avgCpuUsage: cpuFinal,
            avgMemoryUsage: memFinal,
            avgDiskUsage: diskFinal,
            systemHealth,
        };
        
        console.log('[Dashboard] Setting final stats:', newStats);
        return newStats;
      });

    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      // REMOVED setIsLoading(false) - no longer using loading state
      setIsRefreshing(false);
    }
  };

  // SystemHealth component now handles health endpoint data

  // Recompute stats locally when we get websocket updates (keeps UI responsive)
  const recomputeStats = (hostsList: Host[] | null, metricsMap: Record<number, any>, alertsList: Alert[]) => {
    const hostsArr = hostsList || hosts;
    
    console.log('[Dashboard] recomputeStats called with', hostsArr.length, 'hosts and', Object.keys(metricsMap).length, 'metrics');
    
  // metricsKeys was unused; compute stats directly from hosts list and metricsMap
    let totalCpu = 0, totalMemory = 0, totalDisk = 0, validMetrics = 0;
    hostsArr.forEach((host: any) => {
      const m = metricsMap[host.id];
      console.log('[Dashboard] recomputeStats - host', host.id, 'metric:', m);
      if (!m) return;
      let memPercent = m.memory_usage;
      if ((m.memory_total && m.memory_used) || (m.memory_total_bytes && m.memory_used_bytes)) {
        const total = m.memory_total || m.memory_total_bytes;
        const used = m.memory_used || m.memory_used_bytes;
        if (total && used) memPercent = (used / total) * 100;
      }
      let diskPercent = m.disk_usage;
      if ((m.disk_total && m.disk_used) || (m.disk_total_bytes && m.disk_used_bytes)) {
        const dtotal = m.disk_total || m.disk_total_bytes;
        const dused = m.disk_used || m.disk_used_bytes;
        if (dtotal && dused) diskPercent = (dused / dtotal) * 100;
      }
      totalCpu += m.cpu_usage || 0;
      totalMemory += (typeof memPercent === 'number' ? memPercent : 0);
      totalDisk += (typeof diskPercent === 'number' ? diskPercent : 0);
      validMetrics++;
      console.log('[Dashboard] Added to totals - cpu:', (m.cpu_usage || 0), 'mem:', (typeof memPercent === 'number' ? memPercent : 0), 'disk:', (typeof diskPercent === 'number' ? diskPercent : 0));
    });

    const totalHosts = hostsArr.length;
    const onlineHosts = hostsArr.filter((h: any) => h.agent_status === 'online').length;
    const systemHealth = totalHosts > 0 ? Math.round((onlineHosts / totalHosts) * 100) : 100;

    console.log('[Dashboard] recomputeStats result: validMetrics=' + validMetrics + 
      ' cpu=' + (validMetrics > 0 ? (totalCpu / validMetrics).toFixed(1) : 'null') + '%' +
      ' mem=' + (validMetrics > 0 ? (totalMemory / validMetrics).toFixed(1) : 'null') + '%' +
      ' disk=' + (validMetrics > 0 ? (totalDisk / validMetrics).toFixed(1) : 'null') + '%');

    setStats(prev => ({
      ...prev,
      totalHosts,
      onlineHosts,
      offlineHosts: totalHosts - onlineHosts,
      criticalAlerts: (alertsList || alerts).filter(a => a.severity === 'critical').length,
      warningAlerts: (alertsList || alerts).filter(a => a.severity === 'warning' || a.severity === 'high').length,
      // Only update averages if we have valid metrics, otherwise keep previous values
      avgCpuUsage: validMetrics > 0 ? totalCpu / validMetrics : prev.avgCpuUsage,
      avgMemoryUsage: validMetrics > 0 ? totalMemory / validMetrics : prev.avgMemoryUsage,
      avgDiskUsage: validMetrics > 0 ? totalDisk / validMetrics : prev.avgDiskUsage,
      systemHealth,
    }));
  };





  const getTrendIcon = (current: number | null, previous: number) => {
    if (current === null) return <Activity className="w-4 h-4 text-gray-500" />;
    if (current > previous) return <TrendingUp className="w-4 h-4 text-red-500" />;
    if (current < previous) return <TrendingDown className="w-4 h-4 text-green-500" />;
    return <Activity className="w-4 h-4 text-gray-500" />;
  };

  // REMOVED loading spinner completely - it was causing values to flash to 0
  // Dashboard now shows data immediately, no loading state
  
  // Compact Host Analytics Panel (Dashboard)
  // This panel shows detailed metrics when user clicks on a specific host
  return (
    <MainLayout>
      {selectedHostId && (
        <div className="fixed right-6 top-24 w-72 bg-white/95 backdrop-blur-sm rounded-2xl p-4 shadow-2xl border border-gray-100 z-50">
          <div className="flex items-start justify-between mb-2">
            <div>
              <h3 className="text-sm font-bold text-gray-900">Host Analytics</h3>
              <p className="text-xs text-gray-500">Host ID: {selectedHostId}</p>
            </div>
            <button onClick={() => { setSelectedHostId(null); setSelectedHostMetrics(null); }} className="p-1 rounded hover:bg-gray-100">
              <X className="w-4 h-4 text-gray-600" />
            </button>
            </div>

            <div className="flex items-center justify-between">
              <div className="text-center">
                {/* Memory Circle */}
                <div className="relative w-20 h-20 mx-auto mb-2">
                  <svg className="w-20 h-20 transform -rotate-90" viewBox="0 0 36 36">
                    <path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#eef2ff" strokeWidth="2" />
                    <path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#10b981" strokeWidth="2" strokeDasharray={`${Math.round(selectedMemPercent ?? 0)}, 100`} strokeLinecap="round" className="transition-all duration-500" />
                  </svg>
                  <div className="absolute inset-0 flex items-center justify-center">
                    <span className="text-sm font-bold text-gray-800">{selectedMemPercent ? Math.round(selectedMemPercent) : '—'}%</span>
                  </div>
                </div>
                <p className="text-xs text-gray-500">Memory</p>
              </div>

              <div className="pl-2">
                <p className="text-xs text-gray-600">Used / Total</p>
                <p className="text-sm font-semibold text-gray-900">{selectedUsedTotalDisplay}</p>

                <p className="mt-3 text-xs text-gray-500">Uptime</p>
                <p className="text-sm font-medium text-gray-900">{selectedHostMetrics && selectedHostMetrics.uptime ? `${Math.floor(selectedHostMetrics.uptime / 60)}m` : '—'}</p>
              </div>
            </div>

            <div className="mt-3 text-xs text-gray-500">Last updated: {selectedHostMetrics?.timestamp ? new Date(selectedHostMetrics.timestamp).toLocaleString() : '—'}</div>
          </div>
        )}
      
      {/* Main Dashboard Content */}
      <div className="p-4 space-y-4">
        {/* Dashboard Header */}
        <div className="flex items-center justify-between mb-2">
          <h1 className="text-xl font-bold text-gray-800">Dashboard</h1>
          <button
            onClick={() => fetchDashboardData()}
            disabled={isRefreshing}
            className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
            title="Refresh dashboard data"
          >
            <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
            <span className="text-sm font-medium">{isRefreshing ? 'Refreshing...' : 'Refresh'}</span>
          </button>
        </div>

        {/* Compact Metrics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
          {/* Total Hosts - Blue Gradient */}
          <div 
            className="bg-gradient-to-br from-blue-500 to-blue-700 rounded-lg p-3 text-white cursor-pointer transform hover:scale-105 transition-all duration-300 shadow-md hover:shadow-lg"
            onClick={() => navigate('/hosts')}
            title="Total number of monitored hosts. Green dot shows online hosts, red dot shows offline hosts. Click to view all hosts."
          >
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center space-x-1 mb-1">
                  <Server className="w-3 h-3" />
                  <p className="text-blue-100 font-medium text-xs">Total Hosts</p>
                </div>
                <p className="text-lg font-bold mb-1">{stats.totalHosts}</p>
                <div className="flex items-center space-x-2">
                  <div className="flex items-center space-x-1">
                    <div className="w-1.5 h-1.5 bg-green-400 rounded-full"></div>
                    <span className="text-xs font-medium">{stats.onlineHosts}</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <div className="w-1.5 h-1.5 bg-red-400 rounded-full"></div>
                    <span className="text-xs font-medium">{stats.offlineHosts}</span>
                  </div>
                </div>
              </div>
              <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center backdrop-blur-sm">
                <Server className="w-4 h-4" />
              </div>
            </div>
          </div>

          {/* System Health - Green Gradient */}
          <div 
            className="bg-gradient-to-br from-green-500 to-emerald-700 rounded-lg p-3 text-white transform hover:scale-105 transition-all duration-300 shadow-md hover:shadow-lg"
            title="Overall system health score based on average CPU, memory, and disk usage across all online hosts. Higher is better."
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="flex items-center space-x-1 mb-1">
                  <Shield className="w-3 h-3" />
                  <p className="text-green-100 font-medium text-xs">System Health</p>
                </div>
                <p className="text-xl font-bold mb-1">
                  {stats.totalHosts === 0 ? 'N/A' : `${stats.systemHealth}%`}
                </p>
                <div className="w-full bg-white/20 rounded-full h-1.5 backdrop-blur-sm">
                  <div 
                    className="h-1.5 rounded-full bg-gradient-to-r from-white to-green-200 transition-all duration-500 shadow-sm"
                    style={{ width: `${stats.systemHealth}%` }}
                  ></div>
                </div>
              </div>
              <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center backdrop-blur-sm">
                <Shield className="w-4 h-4" />
              </div>
            </div>
          </div>

          {/* Active Alerts - Orange Gradient */}
          <div 
            className="bg-gradient-to-br from-orange-500 to-red-600 rounded-lg p-3 text-white cursor-pointer transform hover:scale-105 transition-all duration-300 shadow-md hover:shadow-lg"
            onClick={() => navigate('/alerts')}
            title="Total active alerts across all hosts. Red dot shows critical alerts, yellow dot shows warnings. Click to view alert details."
          >
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center space-x-1 mb-1">
                  <AlertTriangle className="w-3 h-3" />
                  <p className="text-orange-100 font-medium text-xs">Active Alerts</p>
                </div>
                <p className="text-lg font-bold mb-1">{stats.criticalAlerts + stats.warningAlerts}</p>
                <div className="flex items-center space-x-2">
                  <div className="flex items-center space-x-1">
                    <div className="w-1 h-1 bg-red-300 rounded-full"></div>
                    <span className="text-xs font-medium">{stats.criticalAlerts}</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <div className="w-1 h-1 bg-yellow-300 rounded-full"></div>
                    <span className="text-xs font-medium">{stats.warningAlerts}</span>
                  </div>
                </div>
              </div>
              <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center backdrop-blur-sm">
                <AlertTriangle className="w-4 h-4" />
              </div>
            </div>
          </div>

          {/* Average CPU - Purple Gradient */}
          <div 
            className="bg-gradient-to-br from-purple-500 to-indigo-700 rounded-lg p-3 text-white transform hover:scale-105 transition-all duration-300 shadow-md hover:shadow-lg"
            title="Average CPU usage across all online hosts. Shows percentage of CPU capacity being used. Trend indicator compares to threshold."
          >
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center space-x-1 mb-1">
                  <Cpu className="w-3 h-3" />
                  <p className="text-purple-100 font-medium text-xs">Avg CPU</p>
                </div>
                <p className="text-lg font-bold mb-1">
                  {stats.avgCpuUsage !== null ? `${stats.avgCpuUsage.toFixed(1)}%` : '—'}
                </p>
                <div className="flex items-center space-x-1">
                  {getTrendIcon(stats.avgCpuUsage, 50)}
                  <span className="text-xs font-medium text-purple-100">vs last</span>
                </div>
              </div>
              <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center backdrop-blur-sm">
                <Cpu className="w-4 h-4" />
              </div>
            </div>
          </div>
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {/* Resource Usage Overview */}
          <div className="lg:col-span-2">
            <div className="bg-white/80 backdrop-blur-sm rounded-xl p-4 shadow-lg border border-white/20">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center space-x-2">
                  <div className="w-6 h-6 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                    <Activity className="w-4 h-4 text-white" />
                  </div>
                  <h2 className="text-lg font-bold bg-gradient-to-r from-gray-800 to-gray-600 bg-clip-text text-transparent">Resource Usage</h2>
                </div>
                <button 
                  onClick={() => navigate('/metrics')}
                  className="px-3 py-1 bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-lg hover:shadow-lg transition-all duration-300 hover:scale-105 text-sm"
                >
                  Details <ChevronRight className="w-3 h-3 ml-1 inline" />
                </button>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {/* CPU Usage - Animated Circle */}
                <div 
                  className="text-center p-3 bg-gradient-to-br from-blue-50 to-blue-100 rounded-lg border border-blue-200"
                  title="Average CPU usage percentage across all monitored hosts. Shows how much processing power is being used."
                >
                  <div className="relative w-16 h-16 mx-auto mb-2">
                    <svg className="w-16 h-16 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#dbeafe"
                        strokeWidth="2"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="url(#blueGradient)"
                        strokeWidth="2"
                        strokeDasharray={`${stats.avgCpuUsage ?? 0}, 100`}
                        strokeLinecap="round"
                        className="transition-all duration-1000 ease-out"
                      />
                      <defs>
                        <linearGradient id="blueGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                          <stop offset="0%" stopColor="#3b82f6" />
                          <stop offset="100%" stopColor="#1d4ed8" />
                        </linearGradient>
                      </defs>
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-base font-bold text-blue-700">
                        {stats.avgCpuUsage !== null ? `${stats.avgCpuUsage.toFixed(1)}%` : '—'}
                      </span>
                    </div>
                  </div>
                  <p className="text-xs font-semibold text-blue-700 flex items-center justify-center space-x-1">
                    <Cpu className="w-3 h-3" />
                    <span>CPU</span>
                  </p>
                </div>

                {/* Memory Usage - Animated Circle */}
                <div 
                  className="text-center p-3 bg-gradient-to-br from-green-50 to-emerald-100 rounded-lg border border-green-200"
                  title="Average memory (RAM) usage percentage across all monitored hosts. Shows how much RAM is currently in use."
                >
                  <div className="relative w-16 h-16 mx-auto mb-2">
                    <svg className="w-16 h-16 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#dcfce7"
                        strokeWidth="2"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="url(#greenGradient)"
                        strokeWidth="2"
                        strokeDasharray={`${stats.avgMemoryUsage ?? 0}, 100`}
                        strokeLinecap="round"
                        className="transition-all duration-1000 ease-out"
                      />
                      <defs>
                        <linearGradient id="greenGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                          <stop offset="0%" stopColor="#10b981" />
                          <stop offset="100%" stopColor="#059669" />
                        </linearGradient>
                      </defs>
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-base font-bold text-green-700">
                        {stats.avgMemoryUsage !== null ? `${stats.avgMemoryUsage.toFixed(1)}%` : '—'}
                      </span>
                    </div>
                  </div>
                  <p className="text-xs font-semibold text-green-700 flex items-center justify-center space-x-1">
                    <Activity className="w-3 h-3" />
                    <span>Memory</span>
                  </p>
                </div>

                {/* Disk Usage - Animated Circle */}
                <div 
                  className="text-center p-3 bg-gradient-to-br from-amber-50 to-orange-100 rounded-lg border border-amber-200"
                  title="Average disk space usage percentage across all monitored hosts. Shows how much storage space is used."
                >
                  <div className="relative w-16 h-16 mx-auto mb-2">
                    <svg className="w-16 h-16 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#fef3c7"
                        strokeWidth="2"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="url(#orangeGradient)"
                        strokeWidth="2"
                        strokeDasharray={`${stats.avgDiskUsage ?? 0}, 100`}
                        strokeLinecap="round"
                        className="transition-all duration-1000 ease-out"
                      />
                      <defs>
                        <linearGradient id="orangeGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                          <stop offset="0%" stopColor="#f59e0b" />
                          <stop offset="100%" stopColor="#d97706" />
                        </linearGradient>
                      </defs>
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-base font-bold text-amber-700">
                        {stats.avgDiskUsage !== null ? `${stats.avgDiskUsage.toFixed(1)}%` : '—'}
                      </span>
                    </div>
                  </div>
                  <p className="text-xs font-semibold text-amber-700 flex items-center justify-center space-x-1">
                    <Database className="w-3 h-3" />
                    <span>Disk</span>
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Actions - Paytm Style */}
          <div>
            <div className="bg-white/80 backdrop-blur-sm rounded-xl p-4 shadow-lg border border-white/20">
              <div className="flex items-center space-x-2 mb-4">
                <div className="w-6 h-6 bg-gradient-to-r from-pink-500 to-rose-600 rounded-lg flex items-center justify-center">
                  <TrendingUp className="w-4 h-4 text-white" />
                </div>
                <h2 className="text-lg font-bold bg-gradient-to-r from-gray-800 to-gray-600 bg-clip-text text-transparent">Quick Actions</h2>
              </div>
              <div className="space-y-2">
                <button 
                  onClick={() => navigate('/hosts')}
                  className="w-full p-2 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-lg hover:shadow-lg transition-all duration-300 hover:scale-105 flex items-center space-x-2 text-sm"
                >
                  <div className="w-6 h-6 bg-white/20 rounded-lg flex items-center justify-center">
                    <Server className="w-3 h-3" />
                  </div>
                  <span className="font-semibold">Add Host</span>
                </button>
                <button 
                  onClick={() => navigate('/metrics')}
                  className="w-full p-2 bg-gradient-to-r from-green-500 to-emerald-600 text-white rounded-lg hover:shadow-lg transition-all duration-300 hover:scale-105 flex items-center space-x-2 text-sm"
                >
                  <div className="w-6 h-6 bg-white/20 rounded-lg flex items-center justify-center">
                    <Activity className="w-3 h-3" />
                  </div>
                  <span className="font-semibold">Metrics</span>
                </button>
                <button 
                  onClick={() => navigate('/alerts')}
                  className="w-full p-2 bg-gradient-to-r from-orange-500 to-red-600 text-white rounded-lg hover:shadow-lg transition-all duration-300 hover:scale-105 flex items-center space-x-2 text-sm"
                >
                  <div className="w-6 h-6 bg-white/20 rounded-lg flex items-center justify-center">
                    <AlertTriangle className="w-3 h-3" />
                  </div>
                  <span className="font-semibold">Alerts</span>
                </button>
                <button 
                  onClick={() => navigate('/logs')}
                  className="w-full p-2 bg-gradient-to-r from-purple-500 to-indigo-600 text-white rounded-lg hover:shadow-lg transition-all duration-300 hover:scale-105 flex items-center space-x-2 text-sm"
                >
                  <div className="w-6 h-6 bg-white/20 rounded-lg flex items-center justify-center">
                    <Database className="w-3 h-3" />
                  </div>
                  <span className="font-semibold">Logs</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Hosts Overview - Grafana Style */}
        <div className="bg-white/80 backdrop-blur-sm rounded-xl p-4 shadow-lg border border-white/20">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-2">
              <div className="w-6 h-6 bg-gradient-to-r from-green-500 to-emerald-600 rounded-lg flex items-center justify-center">
                <Server className="w-4 h-4 text-white" />
              </div>
              <h2 className="text-lg font-bold bg-gradient-to-r from-gray-800 to-gray-600 bg-clip-text text-transparent">Hosts Overview</h2>
            </div>
            <Button variant="ghost" size="sm" onClick={() => navigate('/hosts')}>
              View All <ChevronRight className="w-3 h-3 ml-1" />
            </Button>
          </div>

          {hosts.length === 0 ? (
            <div className="text-center py-12">
              <Server className="w-16 h-16 text-gray-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No Hosts Yet</h3>
              <p className="text-gray-500 mb-6">Add your first host to start monitoring</p>
              <Button variant="primary" onClick={() => navigate('/hosts')}>
                Add Host
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
              {hosts.slice(0, 10).map((host) => {
                const metrics = hostMetrics[host.id];
                const cpuUsage = typeof metrics?.cpu_usage === 'number' ? metrics.cpu_usage : null;
                // prefer memory totals when present
                let memoryUsage: number | null = typeof metrics?.memory_usage === 'number' ? metrics.memory_usage : null;
                if ((metrics?.memory_total && metrics?.memory_used) || (metrics?.memory_total_bytes && metrics?.memory_used_bytes)) {
                  const total = metrics?.memory_total || metrics?.memory_total_bytes;
                  const used = metrics?.memory_used || metrics?.memory_used_bytes;
                  if (total && used) memoryUsage = (used / total) * 100;
                }
                const diskUsage = typeof metrics?.disk_usage === 'number' ? metrics.disk_usage : null;
                const isOnline = host.agent_status === 'online';
                
                return (
                  <div 
                    key={host.id} 
                    className="group relative bg-gradient-to-br from-white via-white to-blue-50/30 rounded-2xl p-4 shadow-card hover:shadow-paytm-lg transition-all duration-300 hover:scale-105 cursor-pointer border border-blue-100/50 overflow-hidden"
                    onClick={() => {
                      setSelectedHostId(host.id);
                      hostService.getLatestMetrics(host.id).then(m => setSelectedHostMetrics(m)).catch(() => setSelectedHostMetrics(null));
                    }}
                    onDoubleClick={() => navigate(`/hosts/${host.id}`)}
                  >
                    {/* Paytm-style gradient overlay on hover */}
                    <div className="absolute inset-0 bg-gradient-to-br from-primary-500/5 to-secondary-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
                    
                    <div className="relative z-10">
                      {/* Host Header - Paytm Style */}
                      <div className="flex items-center justify-between mb-4">
                        <div className="flex items-center space-x-2">
                          <div className={`w-9 h-9 rounded-xl flex items-center justify-center shadow-md ${
                            isOnline ? 'bg-gradient-to-br from-primary-500 to-primary-600' : 'bg-gradient-to-br from-gray-400 to-gray-500'
                          }`}>
                            <Server className="w-4 h-4 text-white" />
                          </div>
                          <div>
                            <h3 className="font-bold text-gray-900 text-sm truncate max-w-[120px]">{host.hostname || host.ip}</h3>
                            <p className="text-xs text-gray-500">{host.ip}</p>
                          </div>
                        </div>
                        <div className={`w-2.5 h-2.5 rounded-full shadow-lg ${
                          isOnline ? 'bg-green-500 shadow-green-500/50 animate-pulse' : 'bg-gray-400'
                        }`}></div>
                      </div>

                      {/* Paytm-style Metrics Grid with Gradients */}
                      <div className="grid grid-cols-3 gap-3 mb-3">
                        {/* CPU Circle - Vibrant Blue Gradient */}
                        <div className="text-center">
                          <div className="relative w-11 h-11 mx-auto mb-1.5">
                            <svg className="w-11 h-11 transform -rotate-90" viewBox="0 0 36 36">
                              <path
                                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                                fill="none"
                                stroke="#e0f2fe"
                                strokeWidth="3"
                              />
                              <path
                                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                                fill="none"
                                stroke="url(#cpuGrad)"
                                strokeWidth="3"
                                strokeDasharray={`${cpuUsage ?? 0}, 100`}
                                strokeLinecap="round"
                                className="transition-all duration-1000 ease-out"
                              />
                              <defs>
                                <linearGradient id="cpuGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                                  <stop offset="0%" stopColor="#007bff" />
                                  <stop offset="100%" stopColor="#00b9f5" />
                                </linearGradient>
                              </defs>
                            </svg>
                            <div className="absolute inset-0 flex items-center justify-center">
                              <span className="text-sm font-extrabold bg-gradient-to-br from-primary-600 to-secondary-500 bg-clip-text text-transparent">
                                {cpuUsage !== null ? `${cpuUsage.toFixed(1)}%` : '—'}
                              </span>
                            </div>
                          </div>
                          <p className="text-xs font-semibold text-gray-600">CPU</p>
                        </div>

                        {/* Memory Circle - Vibrant Green Gradient */}
                        <div className="text-center">
                          <div className="relative w-11 h-11 mx-auto mb-1.5">
                            <svg className="w-11 h-11 transform -rotate-90" viewBox="0 0 36 36">
                              <path
                                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                                fill="none"
                                stroke="#d1fae5"
                                strokeWidth="3"
                              />
                              <path
                                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                                fill="none"
                                stroke="url(#memGrad)"
                                strokeWidth="3"
                                strokeDasharray={`${memoryUsage ?? 0}, 100`}
                                strokeLinecap="round"
                                className="transition-all duration-1000 ease-out"
                              />
                              <defs>
                                <linearGradient id="memGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                                  <stop offset="0%" stopColor="#00d26a" />
                                  <stop offset="100%" stopColor="#10b981" />
                                </linearGradient>
                              </defs>
                            </svg>
                            <div className="absolute inset-0 flex items-center justify-center">
                              <span className="text-sm font-extrabold bg-gradient-to-br from-green-600 to-emerald-500 bg-clip-text text-transparent">
                                {memoryUsage !== null ? `${memoryUsage.toFixed(1)}%` : '—'}
                              </span>
                            </div>
                          </div>
                          <p className="text-xs font-semibold text-gray-600">RAM</p>
                        </div>

                      {/* Disk Circle - Vibrant Orange Gradient */}
                      <div className="text-center">
                        <div className="relative w-11 h-11 mx-auto mb-1.5">
                          <svg className="w-11 h-11 transform -rotate-90" viewBox="0 0 36 36">
                            <path
                              d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                              fill="none"
                              stroke="#fed7aa"
                              strokeWidth="3"
                            />
                            <path
                              d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                              fill="none"
                              stroke="url(#diskGrad)"
                              strokeWidth="3"
                              strokeDasharray={`${diskUsage ?? 0}, 100`}
                              strokeLinecap="round"
                              className="transition-all duration-1000 ease-out"
                            />
                            <defs>
                              <linearGradient id="diskGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                                <stop offset="0%" stopColor="#ff9800" />
                                <stop offset="100%" stopColor="#f59e0b" />
                              </linearGradient>
                            </defs>
                          </svg>
                          <div className="absolute inset-0 flex items-center justify-center">
                            <span className="text-sm font-extrabold bg-gradient-to-br from-orange-600 to-amber-500 bg-clip-text text-transparent">
                              {diskUsage !== null ? `${diskUsage.toFixed(1)}%` : '—'}
                            </span>
                          </div>
                        </div>
                        <p className="text-xs font-semibold text-gray-600">Disk</p>
                      </div>
                    </div>

                      {/* Paytm-style Status Badge */}
                      <div className="flex items-center justify-between pt-3 border-t border-gray-100">
                        <div className="flex items-center space-x-2">
                          <div className={`w-2 h-2 rounded-full ${
                            isOnline ? 'bg-green-500 shadow-lg shadow-green-500/50 animate-pulse' : 'bg-gray-400'
                          }`}></div>
                          <span className={`text-xs font-semibold ${
                            isOnline ? 'text-green-600' : 'text-gray-500'
                          }`}>
                            {host.agent_status || 'offline'}
                          </span>
                        </div>
                        <span className="text-xs text-gray-400 font-medium">
                          {host.last_seen ? formatTimestamp(host.last_seen).split(' ')[1] : 'Never'}
                        </span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Recent Alerts */}
        {alerts.length > 0 && (
          <Card className="p-4">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-bold text-gray-900">Recent Alerts</h2>
              <Button variant="ghost" size="sm" onClick={() => navigate('/alerts')}>
                View All <ChevronRight className="w-4 h-4 ml-1" />
              </Button>
            </div>
            
            <div className="space-y-4">
              {alerts.slice(0, 5).map((alert) => (
                <div key={alert.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                  <div className="flex items-center">
                    <div className={`w-3 h-3 rounded-full mr-3 ${
                      alert.severity === 'critical' ? 'bg-red-500' :
                      alert.severity === 'warning' ? 'bg-yellow-500' : 'bg-blue-500'
                    }`}></div>
                    <div>
                      <p className="font-medium text-gray-900">{alert.message || 'Alert triggered'}</p>
                      <p className="text-sm text-gray-500">{alert.host_name || `Host #${alert.host_id}`}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <Badge 
                      variant={alert.severity === 'critical' ? 'error' : 'warning'}
                      size="sm"
                    >
                      {alert.severity}
                    </Badge>
                    <p className="text-xs text-gray-500 mt-1">
                      {new Date(alert.created_at).toLocaleString()}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        )}
      </div>
      </MainLayout>
  );
};

export default Dashboard;