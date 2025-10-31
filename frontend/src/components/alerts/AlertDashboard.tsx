import React, { useState, useEffect } from 'react';
import { 
  AlertTriangle, 
  CheckCircle, 
  Clock, 
  Zap, 
  Bell, 
  BellOff,
  Activity,
  Filter,
  RefreshCw
} from 'lucide-react';
import Card from '../common/Card';
import Button from '../common/Button';
import Badge from '../common/Badge';
import Input from '../common/Input';
import apiClient from '../../services/api/client';

interface AlertStats {
  total: number;
  open: number;
  acknowledged: number;
  silenced: number;
  resolved: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
}

interface Alert {
  id: number;
  title: string;
  description: string;
  severity: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'INFO';
  status: 'OPEN' | 'ACKNOWLEDGED' | 'SILENCED' | 'RESOLVED' | 'CLOSED';
  value: number;
  threshold: number;
  metric_name: string;
  host?: {
    hostname: string;
    ip: string;
  };
  triggered_at: string;
  acknowledged_at?: string;
  silenced_until?: string;
  runbook_url?: string;
  dashboard_url?: string;
}

interface AlertFilters {
  status: string[];
  severity: string[];
  search: string;
  timeRange: string;
}

const AlertDashboard: React.FC = () => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [stats, setStats] = useState<AlertStats | null>(null);
  const [filters, setFilters] = useState<AlertFilters>({
    status: [],
    severity: [],
    search: '',
    timeRange: '24h'
  });
  const [loading, setLoading] = useState(true);
  const [selectedAlerts, setSelectedAlerts] = useState<Set<number>>(new Set());
  const [showFilters, setShowFilters] = useState(false);

  useEffect(() => {
    fetchData();
  }, [filters]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [alertsResponse, statsResponse] = await Promise.all([
        fetchAlerts(),
        fetchStats()
      ]);
      setAlerts(alertsResponse);
      setStats(statsResponse);
    } catch (error) {
      console.error('Failed to fetch alert data:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchAlerts = async (): Promise<Alert[]> => {
    const params = new URLSearchParams();
    
    if (filters.status.length > 0) {
      params.append('status', filters.status.join(','));
    }
    if (filters.severity.length > 0) {
      params.append('severity', filters.severity.join(','));
    }
    if (filters.search) {
      params.append('search', filters.search);
    }
    
    // Convert timeRange to start_time
    const now = new Date();
    let startTime = new Date();
    switch (filters.timeRange) {
      case '1h':
        startTime.setHours(now.getHours() - 1);
        break;
      case '24h':
        startTime.setDate(now.getDate() - 1);
        break;
      case '7d':
        startTime.setDate(now.getDate() - 7);
        break;
      case '30d':
        startTime.setDate(now.getDate() - 30);
        break;
    }
    params.append('start_time', startTime.toISOString());
    
    const response = await apiClient.get(`/alerts?${params.toString()}`);
    return response.data || response || [];
  };

  const fetchStats = async (): Promise<AlertStats> => {
    const response = await apiClient.get('/alerts/stats');
    return response.data || response;
  };

  const handleAlertAction = async (alertIds: number[], action: string, payload: any = {}) => {
    try {
      const promises = alertIds.map(id => 
        apiClient.post(`/alerts/${id}/${action}`, payload)
      );
      await Promise.all(promises);
      
      // Refresh data
      fetchData();
      setSelectedAlerts(new Set());
      
      // Show success message
      console.log(`Successfully ${action}d ${alertIds.length} alert(s)`);
    } catch (error) {
      console.error(`Failed to ${action} alerts:`, error);
    }
  };

  const toggleAlertSelection = (alertId: number) => {
    const newSelection = new Set(selectedAlerts);
    if (newSelection.has(alertId)) {
      newSelection.delete(alertId);
    } else {
      newSelection.add(alertId);
    }
    setSelectedAlerts(newSelection);
  };

  const selectAllAlerts = () => {
    const allIds = alerts.map(alert => alert.id);
    setSelectedAlerts(new Set(allIds));
  };

  const clearSelection = () => {
    setSelectedAlerts(new Set());
  };

  const getSeverityColor = (severity: string): "error" | "warning" | "info" | "success" => {
    switch (severity) {
      case 'CRITICAL': return 'error';
      case 'HIGH': return 'warning';
      case 'MEDIUM': return 'info';
      case 'LOW': return 'success';
      default: return 'info';
    }
  };

  const getStatusColor = (status: string): "error" | "warning" | "info" | "success" => {
    switch (status) {
      case 'OPEN': return 'error';
      case 'ACKNOWLEDGED': return 'warning';
      case 'SILENCED': return 'info';
      case 'RESOLVED': return 'success';
      default: return 'info';
    }
  };

  const formatRelativeTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <span className="ml-2 text-gray-600">Loading alerts...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Alert Dashboard</h1>
          <p className="text-gray-600 mt-1">Monitor and manage your infrastructure alerts</p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
            className="flex items-center gap-2"
          >
            <Filter className="w-4 h-4" />
            Filters
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={fetchData}
            className="flex items-center gap-2"
          >
            <RefreshCw className="w-4 h-4" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Statistics Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Alerts</p>
                <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
              </div>
              <Activity className="w-8 h-8 text-blue-600" />
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Open</p>
                <p className="text-2xl font-bold text-red-600">{stats.open}</p>
              </div>
              <AlertTriangle className="w-8 h-8 text-red-600" />
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Acknowledged</p>
                <p className="text-2xl font-bold text-yellow-600">{stats.acknowledged}</p>
              </div>
              <Bell className="w-8 h-8 text-yellow-600" />
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Silenced</p>
                <p className="text-2xl font-bold text-gray-600">{stats.silenced}</p>
              </div>
              <BellOff className="w-8 h-8 text-gray-600" />
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Critical</p>
                <p className="text-2xl font-bold text-red-800">{stats.critical}</p>
              </div>
              <Zap className="w-8 h-8 text-red-800" />
            </div>
          </Card>
        </div>
      )}

      {/* Filters Panel */}
      {showFilters && (
        <Card className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Search
              </label>
              <Input
                placeholder="Search alerts..."
                value={filters.search}
                onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
                className="w-full"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Status
              </label>
              <select
                multiple
                value={filters.status}
                onChange={(e) => {
                  const values = Array.from(e.target.selectedOptions, option => option.value);
                  setFilters(prev => ({ ...prev, status: values }));
                }}
                className="w-full p-2 border border-gray-300 rounded-md"
              >
                <option value="OPEN">Open</option>
                <option value="ACKNOWLEDGED">Acknowledged</option>
                <option value="SILENCED">Silenced</option>
                <option value="RESOLVED">Resolved</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Severity
              </label>
              <select
                multiple
                value={filters.severity}
                onChange={(e) => {
                  const values = Array.from(e.target.selectedOptions, option => option.value);
                  setFilters(prev => ({ ...prev, severity: values }));
                }}
                className="w-full p-2 border border-gray-300 rounded-md"
              >
                <option value="CRITICAL">Critical</option>
                <option value="HIGH">High</option>
                <option value="MEDIUM">Medium</option>
                <option value="LOW">Low</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Time Range
              </label>
              <select
                value={filters.timeRange}
                onChange={(e) => setFilters(prev => ({ ...prev, timeRange: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md"
              >
                <option value="1h">Last Hour</option>
                <option value="24h">Last 24 Hours</option>
                <option value="7d">Last 7 Days</option>
                <option value="30d">Last 30 Days</option>
              </select>
            </div>
          </div>
        </Card>
      )}

      {/* Bulk Actions */}
      {selectedAlerts.size > 0 && (
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <span className="text-sm font-medium text-gray-700">
                {selectedAlerts.size} alert(s) selected
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={clearSelection}
              >
                Clear Selection
              </Button>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleAlertAction(Array.from(selectedAlerts), 'acknowledge')}
              >
                Acknowledge
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const duration = prompt('Enter silence duration (e.g., 1h, 30m):');
                  if (duration) {
                    handleAlertAction(Array.from(selectedAlerts), 'silence', { duration });
                  }
                }}
              >
                Silence
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleAlertAction(Array.from(selectedAlerts), 'resolve')}
              >
                Resolve
              </Button>
            </div>
          </div>
        </Card>
      )}

      {/* Alerts List */}
      <Card>
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-900">
              Active Alerts ({alerts.length})
            </h2>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={selectAllAlerts}
              >
                Select All
              </Button>
            </div>
          </div>
        </div>

        <div className="divide-y divide-gray-200">
          {alerts.length === 0 ? (
            <div className="p-12 text-center">
              <CheckCircle className="w-12 h-12 text-green-500 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No Active Alerts</h3>
              <p className="text-gray-600">Your infrastructure is running smoothly!</p>
            </div>
          ) : (
            alerts.map((alert) => (
              <div
                key={alert.id}
                className="p-6 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-start gap-4">
                  <input
                    type="checkbox"
                    checked={selectedAlerts.has(alert.id)}
                    onChange={() => toggleAlertSelection(alert.id)}
                    className="mt-1"
                  />
                  
                  <div className="flex-1">
                    <div className="flex items-start justify-between mb-2">
                      <div>
                        <h3 className="text-lg font-semibold text-gray-900 mb-1">
                          {alert.title}
                        </h3>
                        <p className="text-gray-600 text-sm mb-2">
                          {alert.description}
                        </p>
                        <div className="flex items-center gap-4 text-sm text-gray-500">
                          <span>Host: {alert.host?.hostname || 'Unknown'}</span>
                          <span>Metric: {alert.metric_name}</span>
                          <span>Value: {alert.value}</span>
                          <span>Threshold: {alert.threshold}</span>
                        </div>
                      </div>
                      
                      <div className="flex flex-col items-end gap-2">
                        <div className="flex items-center gap-2">
                          <Badge 
                            variant={getSeverityColor(alert.severity)}
                            size="sm"
                          >
                            {alert.severity}
                          </Badge>
                          <Badge 
                            variant={getStatusColor(alert.status)}
                            size="sm"
                          >
                            {alert.status}
                          </Badge>
                        </div>
                        
                        <div className="flex items-center gap-1 text-xs text-gray-500">
                          <Clock className="w-3 h-3" />
                          {formatRelativeTime(alert.triggered_at)}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center justify-between mt-4">
                      <div className="flex items-center gap-2">
                        {alert.runbook_url && (
                          <a
                            href={alert.runbook_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 hover:text-blue-800 text-sm"
                          >
                            📖 Runbook
                          </a>
                        )}
                        {alert.dashboard_url && (
                          <a
                            href={alert.dashboard_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 hover:text-blue-800 text-sm"
                          >
                            📊 Dashboard
                          </a>
                        )}
                      </div>
                      
                      <div className="flex items-center gap-2">
                        {alert.status === 'OPEN' && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleAlertAction([alert.id], 'acknowledge')}
                          >
                            Acknowledge
                          </Button>
                        )}
                        {(alert.status === 'OPEN' || alert.status === 'ACKNOWLEDGED') && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              const duration = prompt('Enter silence duration (e.g., 1h, 30m):');
                              if (duration) {
                                handleAlertAction([alert.id], 'silence', { duration });
                              }
                            }}
                          >
                            Silence
                          </Button>
                        )}
                        {alert.status !== 'RESOLVED' && (
                          <Button
                            variant="primary"
                            size="sm"
                            onClick={() => handleAlertAction([alert.id], 'resolve')}
                          >
                            Resolve
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </Card>
    </div>
  );
};

export default AlertDashboard;