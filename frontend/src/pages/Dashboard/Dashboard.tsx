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
  RefreshCw,
  ChevronRight,
  Database
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import hostService from '../../services/api/hostService';
import apiClient from '../../services/api/client';
// import { useMetrics } from '../../hooks/useMetrics';
import useWebsocket from '../../hooks/useWebsocket';

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
      <div className="p-6 lg:p-8 space-y-8">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Infrastructure Dashboard</h1>
            <p className="text-gray-600 mt-1">Real-time monitoring and system overview</p>
          </div>
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2 px-3 py-2 bg-green-50 rounded-lg border border-green-200">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-sm font-medium text-green-700">Live</span>
            </div>
            <Button variant="outline" onClick={fetchDashboardData}>
              <RefreshCw className="w-4 h-4" />
              Refresh
            </Button>
          </div>
        </div>

        {/* Key Metrics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {/* Total Hosts */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/hosts')}>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Hosts</p>
                <p className="text-3xl font-bold text-gray-900 mt-2">{stats.totalHosts}</p>
                <div className="flex items-center mt-2 space-x-2">
                  <span className="text-sm text-green-600">{stats.onlineHosts} online</span>
                  <span className="text-sm text-red-600">{stats.offlineHosts} offline</span>
                </div>
              </div>
              <div className="p-3 bg-blue-100 rounded-lg">
                <Server className="w-8 h-8 text-blue-600" />
              </div>
            </div>
          </div>

          {/* System Health */}
          <Card className="p-6 hover:shadow-lg transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">System Health</p>
                <p className="text-3xl font-bold text-gray-900 mt-2">{stats.systemHealth}%</p>
                <div className="w-full bg-gray-200 rounded-full h-2 mt-3">
                  <div 
                    className={`h-2 rounded-full transition-all duration-500 ${
                      stats.systemHealth >= 90 ? 'bg-green-500' :
                      stats.systemHealth >= 70 ? 'bg-yellow-500' : 'bg-red-500'
                    }`}
                    style={{ width: `${stats.systemHealth}%` }}
                  ></div>
                </div>
              </div>
              <div className="p-3 bg-green-100 rounded-lg">
                <Shield className="w-8 h-8 text-green-600" />
              </div>
            </div>
          </Card>

          {/* Active Alerts */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-lg transition-shadow cursor-pointer" onClick={() => navigate('/alerts')}>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Active Alerts</p>
                <p className="text-3xl font-bold text-gray-900 mt-2">{stats.criticalAlerts + stats.warningAlerts}</p>
                <div className="flex items-center mt-2 space-x-2">
                  <span className="text-sm text-red-600">{stats.criticalAlerts} critical</span>
                  <span className="text-sm text-yellow-600">{stats.warningAlerts} warning</span>
                </div>
              </div>
              <div className="p-3 bg-red-100 rounded-lg">
                <AlertTriangle className="w-8 h-8 text-red-600" />
              </div>
            </div>
          </div>

          {/* Average CPU */}
          <Card className="p-6 hover:shadow-lg transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Avg CPU Usage</p>
                <p className={`text-3xl font-bold mt-2 ${getMetricColor(stats.avgCpuUsage)}`}>
                  {stats.avgCpuUsage.toFixed(1)}%
                </p>
                <div className="flex items-center mt-2">
                  {getTrendIcon(stats.avgCpuUsage, 50)}
                  <span className="text-sm text-gray-500 ml-1">vs last hour</span>
                </div>
              </div>
              <div className="p-3 bg-purple-100 rounded-lg">
                <Cpu className="w-8 h-8 text-purple-600" />
              </div>
            </div>
          </Card>
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Resource Usage Overview */}
          <div className="lg:col-span-2">
            <Card className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-gray-900">Resource Usage</h2>
                <Button variant="ghost" size="sm" onClick={() => navigate('/metrics')}>
                  View Details <ChevronRight className="w-4 h-4 ml-1" />
                </Button>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* CPU Usage */}
                <div className="text-center">
                  <div className="relative w-24 h-24 mx-auto mb-4">
                    <svg className="w-24 h-24 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#e5e7eb"
                        strokeWidth="2"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#3b82f6"
                        strokeWidth="2"
                        strokeDasharray={`${stats.avgCpuUsage}, 100`}
                      />
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-lg font-bold text-gray-900">{stats.avgCpuUsage.toFixed(0)}%</span>
                    </div>
                  </div>
                  <p className="text-sm font-medium text-gray-600">CPU Usage</p>
                </div>

                {/* Memory Usage */}
                <div className="text-center">
                  <div className="relative w-24 h-24 mx-auto mb-4">
                    <svg className="w-24 h-24 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#e5e7eb"
                        strokeWidth="2"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#10b981"
                        strokeWidth="2"
                        strokeDasharray={`${stats.avgMemoryUsage}, 100`}
                      />
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-lg font-bold text-gray-900">{stats.avgMemoryUsage.toFixed(0)}%</span>
                    </div>
                  </div>
                  <p className="text-sm font-medium text-gray-600">Memory Usage</p>
                </div>

                {/* Disk Usage */}
                <div className="text-center">
                  <div className="relative w-24 h-24 mx-auto mb-4">
                    <svg className="w-24 h-24 transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#e5e7eb"
                        strokeWidth="2"
                      />
                      <path
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                        fill="none"
                        stroke="#f59e0b"
                        strokeWidth="2"
                        strokeDasharray={`${stats.avgDiskUsage}, 100`}
                      />
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-lg font-bold text-gray-900">{stats.avgDiskUsage.toFixed(0)}%</span>
                    </div>
                  </div>
                  <p className="text-sm font-medium text-gray-600">Disk Usage</p>
                </div>
              </div>
            </Card>
          </div>

          {/* Quick Actions */}
          <div>
            <Card className="p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-6">Quick Actions</h2>
              <div className="space-y-3">
                <Button 
                  variant="primary" 
                  fullWidth 
                  className="justify-start"
                  onClick={() => navigate('/hosts')}
                >
                  <Server className="w-4 h-4 mr-3" />
                  Add New Host
                </Button>
                <Button 
                  variant="outline" 
                  fullWidth 
                  className="justify-start"
                  onClick={() => navigate('/metrics')}
                >
                  <Activity className="w-4 h-4 mr-3" />
                  View Metrics
                </Button>
                <Button 
                  variant="outline" 
                  fullWidth 
                  className="justify-start"
                  onClick={() => navigate('/alerts')}
                >
                  <AlertTriangle className="w-4 h-4 mr-3" />
                  Manage Alerts
                </Button>
                <Button 
                  variant="outline" 
                  fullWidth 
                  className="justify-start"
                  onClick={() => navigate('/logs')}
                >
                  <Database className="w-4 h-4 mr-3" />
                  View Logs
                </Button>
              </div>
            </Card>
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
                            {host.last_seen ? new Date(host.last_seen).toLocaleString() : 'Never'}
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