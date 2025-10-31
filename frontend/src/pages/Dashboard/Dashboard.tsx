import React, { useState, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';

import SystemOverview from '../../components/dashboard/SystemOverview';
import ThreatIndicators from '../../components/dashboard/ThreatIndicators';
import LiveMetrics from '../../components/dashboard/LiveMetrics';
import RealTimeStatus from '../../components/dashboard/RealTimeStatus';
import Card from '../../components/common/Card';
import Button from '../../components/common/Button';
import Badge from '../../components/common/Badge';
import { 
  Server, 
  Activity, 
  AlertTriangle, 

  Clock,
  Database,
  Plus,
  Shield
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../services/api/client';
import hostService from '../../services/api/hostService';
import useWebsocket from '../../hooks/useWebsocket';

interface DashboardStats {
  totalHosts: number;
  onlineHosts: number;
  warnings: number;
  critical: number;
}

interface ActivityItem {
  id: number;
  host: string;
  message: string;
  type: 'warning' | 'success' | 'info' | 'error';
  time: string;
}

interface Host {
  id: number;
  hostname?: string;
  ip: string;
  agent_status?: string;
  last_seen?: string | null;
}

interface Alert {
  id: number;
  host_id: number;
  host_name?: string;
  message?: string;
  type?: string;
  severity?: string;
  created_at: string;
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(true);
  const [stats, setStats] = useState<DashboardStats>({
    totalHosts: 0,
    onlineHosts: 0,
    warnings: 0,
    critical: 0,
  });
  const [recentActivity, setRecentActivity] = useState<ActivityItem[]>([]);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [hostMetricsMap, setHostMetricsMap] = useState<Record<number, any>>({});

  // Update host metrics in overview when realtime websocket payloads arrive
  useWebsocket((payload: any) => {
    try {
      if (!payload || !payload.host_id) return;
      const hid = Number(payload.host_id);
      setHostMetricsMap((prev) => {
        // Only update if host exists in current list
        if (!hosts.find((h) => Number(h.id) === hid)) return prev;
        const next = { ...prev };
        const existing = next[hid] || {};
        // merge known fields
        next[hid] = {
          ...existing,
          cpu_usage: payload.cpu_usage ?? payload.CPUUsage ?? existing.cpu_usage,
          memory_usage: payload.memory_usage ?? payload.MemoryUsage ?? existing.memory_usage,
          disk_usage: payload.disk_usage ?? payload.DiskUsage ?? existing.disk_usage,
          network: (payload.network_in || 0) + (payload.network_out || 0) || existing.network,
          timestamp: payload.timestamp || existing.timestamp,
        };
        return next;
      });
    } catch (e) {
      // swallow
    }
  });
  const [error, setError] = useState<string>('');

  // Fetch real data from backend
  useEffect(() => {
    fetchDashboardData();
  }, []);

    const fetchDashboardData = async () => {
    try {
      setIsLoading(true);
      setError('');

      // Fetch hosts data
      try {
        const hostsResponse: any = await hostService.getAllHosts();
        const hosts: Host[] = Array.isArray(hostsResponse) ? hostsResponse : (hostsResponse || []);

        console.log('[Dashboard] Hosts loaded:', hosts.length);

        const totalHosts = hosts.length;
        const onlineHosts = hosts.filter((h: Host) => h.agent_status === 'online').length;
        const warnings = hosts.filter((h: Host) => h.agent_status === 'warning').length;
        const critical = hosts.filter((h: Host) => h.agent_status === 'offline' || h.agent_status === 'critical' || !h.agent_status).length;

        setStats({
          totalHosts,
          onlineHosts,
          warnings,
          critical,
        });

        // store hosts for the overview section
        setHosts(hosts);

        // fetch latest metric for each host (non-blocking)
        Promise.all(hosts.slice(0, 10).map(async (h) => {
          try {
            const metric: any = await hostService.getLatestMetrics(h.id as number);
            return { hostId: h.id, metric };
          } catch (e) {
            return { hostId: h.id, metric: null };
          }
        })).then((rows) => {
          const m: Record<number, any> = {};
          rows.forEach((r) => { if (r && r.hostId) m[Number(r.hostId)] = r.metric; });
          setHostMetricsMap(m);
        }).catch(() => {});
      } catch (hostError: any) {
        console.log('[Dashboard] Hosts endpoint not available:', hostError.message);
        // Don't show error - empty state is expected
        setStats({
          totalHosts: 0,
          onlineHosts: 0,
          warnings: 0,
          critical: 0,
        });
      }

      // Fetch recent activity from multiple sources
      const activities: ActivityItem[] = [];
      
      // 1. Recent alerts
      try {
        const alertsResponse: any = await apiClient.get('/alerts?limit=3');
        const alerts: Alert[] = Array.isArray(alertsResponse) ? alertsResponse : (alertsResponse.data || alertsResponse || []);
        
        alerts.forEach((alert: Alert, index: number) => {
          activities.push({
            id: 1000 + index,
            host: alert.host_name || `Host #${alert.host_id}`,
            message: alert.message || alert.type || 'Alert triggered',
            type: (alert.severity === 'critical' ? 'error' : 
                  alert.severity === 'high' ? 'warning' : 
                  alert.severity === 'medium' ? 'info' : 'success') as 'warning' | 'success' | 'info' | 'error',
            time: getRelativeTime(alert.created_at),
          });
        });
      } catch (alertError) {
        console.log('[Dashboard] No alerts available');
      }
      
      // 2. Host status changes
      hosts.forEach((host, index) => {
        if (host.last_seen) {
          const lastSeenTime = new Date(host.last_seen);
          const timeDiff = Date.now() - lastSeenTime.getTime();
          
          if (timeDiff < 24 * 60 * 60 * 1000) { // Last 24 hours
            activities.push({
              id: 2000 + index,
              host: host.hostname || host.ip,
              message: host.agent_status === 'online' ? 'Host came online' : 'Host status updated',
              type: host.agent_status === 'online' ? 'success' : 'info',
              time: getRelativeTime(host.last_seen),
            });
          }
        }
      });
      
      // 3. Add system events
      if (hosts.length > 0) {
        activities.push({
          id: 3000,
          host: 'System',
          message: `Monitoring ${hosts.length} host${hosts.length > 1 ? 's' : ''}`,
          type: 'info',
          time: 'ongoing',
        });
      }
      
      // Sort by most recent and limit to 5
      activities.sort((a, b) => {
        if (a.time === 'ongoing') return 1;
        if (b.time === 'ongoing') return -1;
        return 0; // Keep original order for now
      });
      
      setRecentActivity(activities.slice(0, 5));

    } catch (err: any) {
      console.error('[Dashboard] General error:', err);
      setStats({
        totalHosts: 0,
        onlineHosts: 0,
        warnings: 0,
        critical: 0,
      });
    } finally {
      setIsLoading(false);
    }
  };


  // Helper function to convert timestamp to relative time
  const getRelativeTime = (timestamp: string): string => {
    const now = new Date();
    const date = new Date(timestamp);
    const diff = Math.floor((now.getTime() - date.getTime()) / 1000); // seconds

    if (diff < 60) return 'just now';
    if (diff < 3600) return `${Math.floor(diff / 60)} min ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)} hour${Math.floor(diff / 3600) > 1 ? 's' : ''} ago`;
    return `${Math.floor(diff / 86400)} day${Math.floor(diff / 86400) > 1 ? 's' : ''} ago`;
  };

  const healthPercentage = stats.totalHosts > 0 
    ? Math.round((stats.onlineHosts / stats.totalHosts) * 100) 
    : 0;

  return (
    <MainLayout>
      <div className="p-6 lg:p-8">
        {/* Page Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">Security Dashboard</h1>
              <p className="text-gray-600">Real-time infrastructure monitoring and threat detection</p>
            </div>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2 px-3 py-2 bg-green-100 rounded-lg">
                <Shield className="w-5 h-5 text-green-600" />
                <span className="text-sm font-medium text-green-700">Protected</span>
              </div>
              <div className="text-right">
                <div className="text-sm text-gray-500">Last updated</div>
                <div className="text-sm font-medium text-gray-900">{new Date().toLocaleTimeString()}</div>
              </div>
            </div>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* System Overview */}
        <div className="mb-8">
          <SystemOverview stats={stats} isLoading={isLoading} />
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          {/* Live Metrics - Only show if we have hosts */}
          {stats.totalHosts > 0 && (
            <div className="lg:col-span-2">
              <LiveMetrics />
            </div>
          )}

          {/* Threat Indicators - Only show real alerts */}
          <div className={stats.totalHosts > 0 ? '' : 'lg:col-span-3'}>
            <ThreatIndicators indicators={recentActivity.map(activity => ({
              id: activity.id.toString(),
              type: activity.type === 'error' ? 'security' : 'performance',
              severity: activity.type === 'error' ? 'critical' : activity.type === 'warning' ? 'high' : 'low',
              title: activity.message,
              description: `Host: ${activity.host}`,
              count: 1,
              timestamp: activity.time
            }))} isLoading={isLoading} />
          </div>
        </div>

        {/* Secondary Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          {/* Recent Activity */}
          <Card className="lg:col-span-2" padding="none">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-bold text-gray-900">Recent Activity</h2>
                <Button 
                  variant="ghost" 
                  size="sm"
                  onClick={() => navigate('/alerts')}
                >
                  View All
                </Button>
              </div>
            </div>
            
            {/* Show loading, empty state, or activity list */}
            {isLoading ? (
              <div className="p-8 text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                <p className="text-sm text-gray-500 mt-2">Loading activity...</p>
              </div>
            ) : recentActivity.length === 0 ? (
              <div className="p-8 text-center">
                <Activity className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                <p className="text-gray-500">No recent activity</p>
                <p className="text-sm text-gray-400 mt-1">Activity will appear here once you add hosts</p>
              </div>
            ) : (
              <div className="divide-y divide-gray-200">
                {recentActivity.map((activity) => (
                  <div key={activity.id} className="p-4 hover:bg-gray-50 transition-colors">
                    <div className="flex items-start gap-4">
                      <div className={`mt-1 ${
                        activity.type === 'error' ? 'text-red-500' :
                        activity.type === 'warning' ? 'text-orange-500' :
                        activity.type === 'success' ? 'text-green-500' :
                        'text-blue-600'
                      }`}>
                        <Activity className="w-5 h-5" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900">{activity.host}</p>
                        <p className="text-sm text-gray-600 mt-1">{activity.message}</p>
                      </div>
                      <div className="flex items-center gap-2 flex-shrink-0">
                        <Badge
                          variant={
                            activity.type === 'error' ? 'error' :
                            activity.type === 'warning' ? 'warning' :
                            activity.type === 'success' ? 'success' :
                            'info'
                          }
                          size="sm"
                        >
                          {activity.type}
                        </Badge>
                        <span className="text-xs text-gray-500 flex items-center gap-1 whitespace-nowrap">
                          <Clock className="w-3 h-3" />
                          {activity.time}
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>

          {/* Security Actions */}
          <Card>
            <h2 className="text-xl font-bold text-gray-900 mb-4">Security Actions</h2>
            <div className="space-y-3">
              <Button 
                variant="primary" 
                fullWidth 
                className="justify-start bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
                onClick={() => navigate('/hosts')}
              >
                <Plus className="w-4 h-4" />
                Add New Host
              </Button>
              <Button 
                variant="outline" 
                fullWidth 
                className="justify-start border-blue-200 hover:bg-blue-50"
                onClick={() => navigate('/metrics')}
              >
                <Activity className="w-4 h-4" />
                View Metrics
              </Button>
              <Button 
                variant="outline" 
                fullWidth 
                className="justify-start border-green-200 hover:bg-green-50"
                onClick={() => navigate('/logs')}
              >
                <Database className="w-4 h-4" />
                Security Logs
              </Button>
              <Button 
                variant="outline" 
                fullWidth 
                className="justify-start border-red-200 hover:bg-red-50"
                onClick={() => navigate('/alerts')}
              >
                <AlertTriangle className="w-4 h-4" />
                Threat Alerts
              </Button>
            </div>

            {/* Security Status */}
            {!isLoading && stats.totalHosts > 0 && (
              <div className="mt-6 p-4 bg-gradient-to-r from-green-50 to-blue-50 rounded-lg border border-green-200">
                <div className="flex items-start gap-3">
                  <Shield className="w-5 h-5 text-green-600 mt-0.5 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-green-900">Security Status</h3>
                    <p className="text-xs text-green-700 mt-1">
                      {healthPercentage}% of endpoints are secure and monitored
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Real-time Status */}
            <div className="mt-4">
              <RealTimeStatus isConnected={!isLoading} />
            </div>

            {/* Threat Level Indicator */}
            <div className="mt-4 p-4 bg-gray-50 rounded-lg border border-gray-200">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-700">Threat Level</span>
                <Badge variant="success" size="sm">Low</Badge>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div className="bg-green-500 h-2 rounded-full" style={{ width: '25%' }}></div>
              </div>
              <p className="text-xs text-gray-600 mt-2">No active threats detected</p>
            </div>

            {/* Empty State - No Hosts */}
            {!isLoading && stats.totalHosts === 0 && (
              <div className="mt-6 p-4 bg-yellow-50 rounded-lg border border-yellow-200">
                <div className="flex items-start gap-3">
                  <Server className="w-5 h-5 text-yellow-600 mt-0.5 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-yellow-900">Get Started</h3>
                    <p className="text-xs text-yellow-700 mt-1">
                      Add your first host to start monitoring your infrastructure
                    </p>
                  </div>
                </div>
              </div>
            )}
          </Card>
        </div>

        {/* Top Hosts by Resource Usage - Only show if hosts exist */}
        {!isLoading && stats.totalHosts > 0 && (
          <Card padding="none">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-xl font-bold text-gray-900">Hosts Overview</h2>
            </div>
            <div className="p-4">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {hosts.length === 0 ? (
                  <div className="text-center text-gray-500 col-span-full">
                    <Server className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                    <p>Host metrics will be displayed here</p>
                    <p className="text-sm text-gray-400 mt-1">Real-time resource usage coming soon</p>
                  </div>
                ) : (
                  hosts.slice(0, 6).map((h) => {
                    const metric = hostMetricsMap[h.id];
                    const cpuVal = metric ? Number(metric.cpu_usage ?? metric.CPUUsage ?? metric.cpu ?? 0) : 0;
                    const memVal = metric ? Number(metric.memory_usage ?? metric.MemoryUsage ?? metric.memory ?? 0) : 0;
                    const diskVal = metric ? Number(metric.disk_usage ?? metric.DiskUsage ?? metric.disk ?? 0) : 0;
                    return (
                      <div key={h.id} className="p-4 bg-white rounded-lg border shadow-sm hover:shadow-md transition-shadow cursor-pointer" onClick={() => navigate(`/hosts/${h.id}`)}>
                        <div className="flex items-center justify-between mb-2">
                          <div>
                            <p className="text-sm font-medium text-gray-900">{h.hostname || h.ip}</p>
                            <div className="flex items-center gap-2 mt-1">
                              <div className={`w-2 h-2 rounded-full ${
                                h.agent_status === 'online' ? 'bg-green-500' : 'bg-red-500'
                              }`}></div>
                              <p className="text-xs text-gray-500">{h.agent_status || 'offline'}</p>
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="text-lg font-bold text-gray-900">{cpuVal.toFixed(1)}%</p>
                            <p className="text-xs text-gray-500">CPU</p>
                          </div>
                        </div>
                        <div className="text-sm text-gray-600">
                          <div>Memory: {memVal.toFixed(1)}%</div>
                          <div>Disk: {diskVal.toFixed(1)}%</div>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </div>
          </Card>
        )}

        {/* Empty State - No Hosts at All */}
        {!isLoading && stats.totalHosts === 0 && (
          <Card>
            <div className="text-center py-12">
              <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Server className="w-10 h-10 text-blue-600" />
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-2">No Hosts Yet</h3>
              <p className="text-gray-600 mb-6 max-w-md mx-auto">
                Start monitoring your infrastructure by adding your first host. You'll be able to track metrics, logs, and alerts in real-time.
              </p>
              <Button 
                variant="primary" 
                size="lg"
                onClick={() => navigate('/hosts')}
              >
                <Plus className="w-5 h-5" />
                Add Your First Host
              </Button>
            </div>
          </Card>
        )}
      </div>
    </MainLayout>
  );
};

export default Dashboard;
