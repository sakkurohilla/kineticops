import React, { useState, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import StatsCard from './StatsCard';
import Card from '../../components/common/Card';
import Button from '../../components/common/Button';
import Badge from '../../components/common/Badge';
import { 
  Server, 
  Activity, 
  AlertTriangle, 
  CheckCircle,
  TrendingUp,
  Clock,
  Database,
  Plus
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../services/api/client';

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
  name: string;
  status: string;
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
        const hostsResponse: any = await apiClient.get('/hosts');
        const hosts: Host[] = Array.isArray(hostsResponse) ? hostsResponse : (hostsResponse.data || hostsResponse || []);

        console.log('[Dashboard] Hosts loaded:', hosts.length);

        const totalHosts = hosts.length;
        const onlineHosts = hosts.filter((h: Host) => h.status === 'online').length;
        const warnings = hosts.filter((h: Host) => h.status === 'warning').length;
        const critical = hosts.filter((h: Host) => h.status === 'critical' || h.status === 'offline').length;

        setStats({
          totalHosts,
          onlineHosts,
          warnings,
          critical,
        });
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

      // Fetch recent alerts
      try {
        const alertsResponse: any = await apiClient.get('/alerts?limit=5');
        const alerts: Alert[] = Array.isArray(alertsResponse) ? alertsResponse : (alertsResponse.data || alertsResponse || []);
        
        const activity: ActivityItem[] = alerts.map((alert: Alert, index: number) => ({
          id: index,
          host: alert.host_name || `Host #${alert.host_id}`,
          message: alert.message || alert.type || 'Alert triggered',
          type: (alert.severity === 'critical' ? 'error' : 
                alert.severity === 'high' ? 'warning' : 
                alert.severity === 'medium' ? 'info' : 'success') as 'warning' | 'success' | 'info' | 'error',
          time: getRelativeTime(alert.created_at),
        }));
        
        setRecentActivity(activity);
      } catch (alertError) {
        console.log('[Dashboard] No alerts available');
        setRecentActivity([]);
      }

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
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Dashboard</h1>
          <p className="text-gray-600">Welcome back! Here's what's happening with your infrastructure.</p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* Stats Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <StatsCard
            title="Total Hosts"
            value={stats.totalHosts}
            icon={Server}
            color="primary"
            isLoading={isLoading}
          />
          <StatsCard
            title="Online"
            value={stats.onlineHosts}
            icon={CheckCircle}
            color="success"
            isLoading={isLoading}
          />
          <StatsCard
            title="Warnings"
            value={stats.warnings}
            icon={AlertTriangle}
            color="warning"
            isLoading={isLoading}
          />
          <StatsCard
            title="Critical"
            value={stats.critical}
            icon={Activity}
            color="error"
            isLoading={isLoading}
          />
        </div>

        {/* Main Content Grid */}
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

          {/* Quick Actions */}
          <Card>
            <h2 className="text-xl font-bold text-gray-900 mb-4">Quick Actions</h2>
            <div className="space-y-3">
              <Button 
                variant="primary" 
                fullWidth 
                className="justify-start"
                onClick={() => navigate('/hosts')}
              >
                <Plus className="w-4 h-4" />
                Add New Host
              </Button>
              <Button 
                variant="outline" 
                fullWidth 
                className="justify-start"
                onClick={() => navigate('/metrics')}
              >
                <Activity className="w-4 h-4" />
                View Metrics
              </Button>
              <Button 
                variant="outline" 
                fullWidth 
                className="justify-start"
                onClick={() => navigate('/logs')}
              >
                <Database className="w-4 h-4" />
                Check Logs
              </Button>
              <Button 
                variant="outline" 
                fullWidth 
                className="justify-start"
                onClick={() => navigate('/alerts')}
              >
                <AlertTriangle className="w-4 h-4" />
                Manage Alerts
              </Button>
            </div>

            {/* System Health Info */}
            {!isLoading && stats.totalHosts > 0 && (
              <div className="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
                <div className="flex items-start gap-3">
                  <TrendingUp className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-blue-900">System Health</h3>
                    <p className="text-xs text-blue-700 mt-1">
                      {healthPercentage}% of your infrastructure is running smoothly
                    </p>
                  </div>
                </div>
              </div>
            )}

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
            <div className="p-6">
              <div className="text-center text-gray-500">
                <Server className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                <p>Host metrics will be displayed here</p>
                <p className="text-sm text-gray-400 mt-1">Real-time resource usage coming soon</p>
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
