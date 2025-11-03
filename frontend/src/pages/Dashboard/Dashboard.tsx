import React, { useState, useEffect } from 'react';
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
  Eye,
  Settings,
  ChevronRight,
  Database
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import hostService from '../../services/api/hostService';
import apiClient from '../../services/api/client';
// import { useMetrics } from '../../hooks/useMetrics';
import useWebsocket from '../../hooks/useWebsocket';
import { formatTimestamp } from '../../utils/dateUtils';

interface DashboardStats {
  totalHosts: number;
  onlineHosts: number;
  offlineHosts: number;
  criticalAlerts: number;
  warningAlerts: number;
  avgCpuUsage: number;
  avgMemoryUsage: number;
  avgDiskUsage: number;
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
  const [isLoading, setIsLoading] = useState(true);
  const [stats, setStats] = useState<DashboardStats>({
    totalHosts: 0,
    onlineHosts: 0,
    offlineHosts: 0,
    criticalAlerts: 0,
    warningAlerts: 0,
    avgCpuUsage: 0,
    avgMemoryUsage: 0,
    avgDiskUsage: 0,
    systemHealth: 100,
  });
  const [hosts, setHosts] = useState<Host[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [hostMetrics, setHostMetrics] = useState<Record<number, any>>({});
  
  // const { data: metricsData } = useMetrics('1h', undefined, true);

  // Real-time updates via WebSocket
  useWebsocket((payload: any) => {
    if (payload?.host_id) {
      setHostMetrics(prev => ({
        ...prev,
        [payload.host_id]: {
          cpu_usage: payload.cpu_usage || 0,
          memory_usage: payload.memory_usage || 0,
          disk_usage: payload.disk_usage || 0,
          network_in: payload.network_in || 0,
          network_out: payload.network_out || 0,
          timestamp: payload.timestamp || new Date().toISOString(),
        }
      }));
    }
  });

  useEffect(() => {
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, []);

  const fetchDashboardData = async () => {
    try {
      setIsLoading(true);

      // Fetch hosts
      const hostsData = await hostService.getAllHosts();
      setHosts(hostsData);

      // Fetch alerts
      try {
        const alertsData = await apiClient.get('/alerts?limit=10');
        setAlerts(Array.isArray(alertsData) ? alertsData : []);
      } catch (err) {
        setAlerts([]);
      }

      // Calculate stats
      const totalHosts = hostsData.length;
      const onlineHosts = hostsData.filter(h => h.agent_status === 'online').length;
      const offlineHosts = totalHosts - onlineHosts;
      
      // Fetch latest metrics for each host
      const metricsPromises = hostsData.map(async (host) => {
        try {
          const metrics = await hostService.getLatestMetrics(host.id);
          return { hostId: host.id, metrics };
        } catch (err) {
          return { hostId: host.id, metrics: null };
        }
      });

      const metricsResults = await Promise.all(metricsPromises);
      const metricsMap: Record<number, any> = {};
      let totalCpu = 0, totalMemory = 0, totalDisk = 0, validMetrics = 0;

      metricsResults.forEach(({ hostId, metrics }) => {
        if (metrics) {
          metricsMap[hostId] = metrics;
          totalCpu += metrics.cpu_usage || 0;
          totalMemory += metrics.memory_usage || 0;
          totalDisk += metrics.disk_usage || 0;
          validMetrics++;
        }
      });

      setHostMetrics(metricsMap);

      const avgCpuUsage = validMetrics > 0 ? totalCpu / validMetrics : 0;
      const avgMemoryUsage = validMetrics > 0 ? totalMemory / validMetrics : 0;
      const avgDiskUsage = validMetrics > 0 ? totalDisk / validMetrics : 0;
      
      const systemHealth = totalHosts > 0 ? Math.round((onlineHosts / totalHosts) * 100) : 100;

      setStats({
        totalHosts,
        onlineHosts,
        offlineHosts,
        criticalAlerts: alerts.filter(a => a.severity === 'critical').length,
        warningAlerts: alerts.filter(a => a.severity === 'warning' || a.severity === 'high').length,
        avgCpuUsage,
        avgMemoryUsage,
        avgDiskUsage,
        systemHealth,
      });

    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      setIsLoading(false);
    }
  };



  const getMetricColor = (value: number) => {
    if (value >= 90) return 'text-red-600';
    if (value >= 70) return 'text-yellow-600';
    return 'text-green-600';
  };

  const getTrendIcon = (current: number, previous: number) => {
    if (current > previous) return <TrendingUp className="w-4 h-4 text-red-500" />;
    if (current < previous) return <TrendingDown className="w-4 h-4 text-green-500" />;
    return <Activity className="w-4 h-4 text-gray-500" />;
  };

  if (isLoading) {
    return (
      <MainLayout>
        <div className="p-6 lg:p-8">
          <div className="animate-pulse space-y-6">
            <div className="h-8 bg-gray-200 rounded w-1/3"></div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {[1, 2, 3, 4].map(i => (
                <div key={i} className="h-32 bg-gray-200 rounded-lg"></div>
              ))}
            </div>
          </div>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="p-8 space-y-8">

        {/* Paytm-Style Colorful Metrics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {/* Total Hosts - Blue Gradient */}
          <div 
            className="bg-gradient-to-br from-blue-500 to-blue-700 rounded-3xl p-6 text-white cursor-pointer transform hover:scale-105 transition-all duration-300 shadow-xl hover:shadow-2xl"
            onClick={() => navigate('/hosts')}
          >
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center space-x-2 mb-2">
                  <Server className="w-6 h-6" />
                  <p className="text-blue-100 font-medium">Total Hosts</p>
                </div>
                <p className="text-4xl font-bold mb-3">{stats.totalHosts}</p>
                <div className="flex items-center space-x-3">
                  <div className="flex items-center space-x-1">
                    <div className="w-2 h-2 bg-green-400 rounded-full"></div>
                    <span className="text-sm font-medium">{stats.onlineHosts} online</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <div className="w-2 h-2 bg-red-400 rounded-full"></div>
                    <span className="text-sm font-medium">{stats.offlineHosts} offline</span>
                  </div>
                </div>
              </div>
              <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center backdrop-blur-sm">
                <Server className="w-8 h-8" />
              </div>
            </div>
          </div>

          {/* System Health - Green Gradient */}
          <div className="bg-gradient-to-br from-green-500 to-emerald-700 rounded-3xl p-6 text-white transform hover:scale-105 transition-all duration-300 shadow-xl hover:shadow-2xl">
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="flex items-center space-x-2 mb-2">
                  <Shield className="w-6 h-6" />
                  <p className="text-green-100 font-medium">System Health</p>
                </div>
                <p className="text-4xl font-bold mb-3">{stats.systemHealth}%</p>
                <div className="w-full bg-white/20 rounded-full h-3 backdrop-blur-sm">
                  <div 
                    className="h-3 rounded-full bg-gradient-to-r from-white to-green-200 transition-all duration-500 shadow-sm"
                    style={{ width: `${stats.systemHealth}%` }}
                  ></div>
                </div>
              </div>
              <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center backdrop-blur-sm">
                <Shield className="w-8 h-8" />
              </div>
            </div>
          </div>

          {/* Active Alerts - Orange Gradient */}
          <div 
            className="bg-gradient-to-br from-orange-500 to-red-600 rounded-3xl p-6 text-white cursor-pointer transform hover:scale-105 transition-all duration-300 shadow-xl hover:shadow-2xl"
            onClick={() => navigate('/alerts')}
          >
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center space-x-2 mb-2">
                  <AlertTriangle className="w-6 h-6" />
                  <p className="text-orange-100 font-medium">Active Alerts</p>
                </div>
                <p className="text-4xl font-bold mb-3">{stats.criticalAlerts + stats.warningAlerts}</p>
                <div className="flex items-center space-x-3">
                  <div className="flex items-center space-x-1">
                    <div className="w-2 h-2 bg-red-300 rounded-full"></div>
                    <span className="text-sm font-medium">{stats.criticalAlerts} critical</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <div className="w-2 h-2 bg-yellow-300 rounded-full"></div>
                    <span className="text-sm font-medium">{stats.warningAlerts} warning</span>
                  </div>
                </div>
              </div>
              <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center backdrop-blur-sm">
                <AlertTriangle className="w-8 h-8" />
              </div>
            </div>
          </div>

          {/* Average CPU - Purple Gradient */}
          <div className="bg-gradient-to-br from-purple-500 to-indigo-700 rounded-3xl p-6 text-white transform hover:scale-105 transition-all duration-300 shadow-xl hover:shadow-2xl">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center space-x-2 mb-2">
                  <Cpu className="w-6 h-6" />
                  <p className="text-purple-100 font-medium">Avg CPU Usage</p>
                </div>
                <p className="text-4xl font-bold mb-3">{stats.avgCpuUsage.toFixed(1)}%</p>
                <div className="flex items-center space-x-2">
                  {getTrendIcon(stats.avgCpuUsage, 50)}
                  <span className="text-sm font-medium text-purple-100">vs last hour</span>
                </div>
              </div>
              <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center backdrop-blur-sm">
                <Cpu className="w-8 h-8" />
              </div>
            </div>
          </div>
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Resource Usage Overview */}
          <div className="lg:col-span-2">
            <div className="bg-white/80 backdrop-blur-sm rounded-3xl p-8 shadow-xl border border-white/20">
              <div className="flex items-center justify-between mb-8">
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 bg-gradient-to-r from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center">
                    <Activity className="w-6 h-6 text-white" />
                  </div>
                  <h2 className="text-2xl font-bold bg-gradient-to-r from-gray-800 to-gray-600 bg-clip-text text-transparent">Resource Usage</h2>
                </div>
                <button 
                  onClick={() => navigate('/metrics')}
                  className="px-4 py-2 bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-full hover:shadow-lg transition-all duration-300 hover:scale-105"
                >
                  View Details <ChevronRight className="w-4 h-4 ml-1 inline" />
                </button>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                {/* CPU Usage - Animated Circle */}
                <div className="text-center p-6 bg-gradient-to-br from-blue-50 to-blue-100 rounded-2xl border border-blue-200">
                  <div className="relative w-28 h-28 mx-auto mb-4">
                    <svg className="w-28 h-28 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#dbeafe"
                        strokeWidth="3"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="url(#blueGradient)"
                        strokeWidth="3"
                        strokeDasharray={`${stats.avgCpuUsage}, 100`}
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
                      <div className="text-center">
                        <span className="text-2xl font-bold text-blue-700">{stats.avgCpuUsage.toFixed(0)}%</span>
                        <div className="w-1 h-1 bg-blue-500 rounded-full mx-auto mt-1 animate-pulse"></div>
                      </div>
                    </div>
                  </div>
                  <p className="text-sm font-semibold text-blue-700 flex items-center justify-center space-x-2">
                    <Cpu className="w-4 h-4" />
                    <span>CPU Usage</span>
                  </p>
                </div>

                {/* Memory Usage - Animated Circle */}
                <div className="text-center p-6 bg-gradient-to-br from-green-50 to-emerald-100 rounded-2xl border border-green-200">
                  <div className="relative w-28 h-28 mx-auto mb-4">
                    <svg className="w-28 h-28 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#dcfce7"
                        strokeWidth="3"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="url(#greenGradient)"
                        strokeWidth="3"
                        strokeDasharray={`${stats.avgMemoryUsage}, 100`}
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
                      <div className="text-center">
                        <span className="text-2xl font-bold text-green-700">{stats.avgMemoryUsage.toFixed(0)}%</span>
                        <div className="w-1 h-1 bg-green-500 rounded-full mx-auto mt-1 animate-pulse"></div>
                      </div>
                    </div>
                  </div>
                  <p className="text-sm font-semibold text-green-700 flex items-center justify-center space-x-2">
                    <Activity className="w-4 h-4" />
                    <span>Memory Usage</span>
                  </p>
                </div>

                {/* Disk Usage - Animated Circle */}
                <div className="text-center p-6 bg-gradient-to-br from-amber-50 to-orange-100 rounded-2xl border border-amber-200">
                  <div className="relative w-28 h-28 mx-auto mb-4">
                    <svg className="w-28 h-28 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#fef3c7"
                        strokeWidth="3"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="url(#orangeGradient)"
                        strokeWidth="3"
                        strokeDasharray={`${stats.avgDiskUsage}, 100`}
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
                      <div className="text-center">
                        <span className="text-2xl font-bold text-amber-700">{stats.avgDiskUsage.toFixed(0)}%</span>
                        <div className="w-1 h-1 bg-amber-500 rounded-full mx-auto mt-1 animate-pulse"></div>
                      </div>
                    </div>
                  </div>
                  <p className="text-sm font-semibold text-amber-700 flex items-center justify-center space-x-2">
                    <Database className="w-4 h-4" />
                    <span>Disk Usage</span>
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Actions - Paytm Style */}
          <div>
            <div className="bg-white/80 backdrop-blur-sm rounded-3xl p-8 shadow-xl border border-white/20">
              <div className="flex items-center space-x-3 mb-8">
                <div className="w-10 h-10 bg-gradient-to-r from-pink-500 to-rose-600 rounded-2xl flex items-center justify-center">
                  <TrendingUp className="w-6 h-6 text-white" />
                </div>
                <h2 className="text-2xl font-bold bg-gradient-to-r from-gray-800 to-gray-600 bg-clip-text text-transparent">Quick Actions</h2>
              </div>
              <div className="space-y-4">
                <button 
                  onClick={() => navigate('/hosts')}
                  className="w-full p-4 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-2xl hover:shadow-lg transition-all duration-300 hover:scale-105 flex items-center space-x-3"
                >
                  <div className="w-10 h-10 bg-white/20 rounded-xl flex items-center justify-center">
                    <Server className="w-5 h-5" />
                  </div>
                  <span className="font-semibold">Add New Host</span>
                </button>
                <button 
                  onClick={() => navigate('/metrics')}
                  className="w-full p-4 bg-gradient-to-r from-green-500 to-emerald-600 text-white rounded-2xl hover:shadow-lg transition-all duration-300 hover:scale-105 flex items-center space-x-3"
                >
                  <div className="w-10 h-10 bg-white/20 rounded-xl flex items-center justify-center">
                    <Activity className="w-5 h-5" />
                  </div>
                  <span className="font-semibold">View Metrics</span>
                </button>
                <button 
                  onClick={() => navigate('/alerts')}
                  className="w-full p-4 bg-gradient-to-r from-orange-500 to-red-600 text-white rounded-2xl hover:shadow-lg transition-all duration-300 hover:scale-105 flex items-center space-x-3"
                >
                  <div className="w-10 h-10 bg-white/20 rounded-xl flex items-center justify-center">
                    <AlertTriangle className="w-5 h-5" />
                  </div>
                  <span className="font-semibold">Manage Alerts</span>
                </button>
                <button 
                  onClick={() => navigate('/logs')}
                  className="w-full p-4 bg-gradient-to-r from-purple-500 to-indigo-600 text-white rounded-2xl hover:shadow-lg transition-all duration-300 hover:scale-105 flex items-center space-x-3"
                >
                  <div className="w-10 h-10 bg-white/20 rounded-xl flex items-center justify-center">
                    <Database className="w-5 h-5" />
                  </div>
                  <span className="font-semibold">View Logs</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Hosts Overview */}
        <Card className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-bold text-gray-900">Hosts Overview</h2>
            <Button variant="ghost" size="sm" onClick={() => navigate('/hosts')}>
              View All <ChevronRight className="w-4 h-4 ml-1" />
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
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200">
                    <th className="text-left py-3 px-4 font-medium text-gray-600">Host</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-600">Status</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-600">CPU</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-600">Memory</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-600">Disk</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-600">Last Seen</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-600">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {hosts.slice(0, 5).map((host) => {
                    const metrics = hostMetrics[host.id];
                    return (
                      <tr key={host.id} className="border-b border-gray-100 hover:bg-gray-50">
                        <td className="py-4 px-4">
                          <div className="flex items-center">
                            <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mr-3">
                              <Server className="w-5 h-5 text-blue-600" />
                            </div>
                            <div>
                              <p className="font-medium text-gray-900">{host.hostname || host.ip}</p>
                              <p className="text-sm text-gray-500">{host.ip}</p>
                            </div>
                          </div>
                        </td>
                        <td className="py-4 px-4">
                          <Badge 
                            variant={host.agent_status === 'online' ? 'success' : 'error'}
                            size="sm"
                          >
                            {host.agent_status || 'offline'}
                          </Badge>
                        </td>
                        <td className="py-4 px-4">
                          <span className={`font-medium ${getMetricColor(metrics?.cpu_usage || 0)}`}>
                            {(metrics?.cpu_usage || 0).toFixed(1)}%
                          </span>
                        </td>
                        <td className="py-4 px-4">
                          <span className={`font-medium ${getMetricColor(metrics?.memory_usage || 0)}`}>
                            {(metrics?.memory_usage || 0).toFixed(1)}%
                          </span>
                        </td>
                        <td className="py-4 px-4">
                          <span className={`font-medium ${getMetricColor(metrics?.disk_usage || 0)}`}>
                            {(metrics?.disk_usage || 0).toFixed(1)}%
                          </span>
                        </td>
                        <td className="py-4 px-4">
                          <span className="text-sm text-gray-500">
                            {host.last_seen ? formatTimestamp(host.last_seen) : 'Never'}
                          </span>
                        </td>
                        <td className="py-4 px-4">
                          <div className="flex items-center space-x-2">
                            <Button 
                              variant="ghost" 
                              size="sm"
                              onClick={() => navigate(`/hosts/${host.id}`)}
                            >
                              <Eye className="w-4 h-4" />
                            </Button>
                            <Button 
                              variant="ghost" 
                              size="sm"
                              onClick={() => navigate(`/hosts/${host.id}`)}
                            >
                              <Settings className="w-4 h-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </Card>

        {/* Recent Alerts */}
        {alerts.length > 0 && (
          <Card className="p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-gray-900">Recent Alerts</h2>
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