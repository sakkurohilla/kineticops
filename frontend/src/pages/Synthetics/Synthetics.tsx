import React, { useState, useEffect } from 'react';
import { Globe, Play, Pause, Plus, TrendingUp, AlertTriangle, Clock, CheckCircle } from 'lucide-react';
import MainLayout from '../../components/layout/MainLayout';
import Card from '../../components/common/Card';
import Button from '../../components/common/Button';
import Badge from '../../components/common/Badge';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import apiClient from '../../services/api/client';

interface SyntheticMonitor {
  id: number;
  name: string;
  type: string;
  status: string;
  frequency: number;
  config: string;
  created_at: string;
}

interface SyntheticResult {
  id: number;
  monitor_id: number;
  success: boolean;
  duration: number;
  status_code: number;
  error: string;
  timestamp: string;
}

interface MonitorStats {
  total_runs: number;
  successful_runs: number;
  failed_runs: number;
  avg_duration: number;
  uptime: number;
}

const Synthetics: React.FC = () => {
  const [monitors, setMonitors] = useState<SyntheticMonitor[]>([]);
  const [selectedMonitor, setSelectedMonitor] = useState<SyntheticMonitor | null>(null);
  const [results, setResults] = useState<SyntheticResult[]>([]);
  const [stats, setStats] = useState<MonitorStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    fetchMonitors();
  }, []);

  useEffect(() => {
    if (selectedMonitor) {
      fetchMonitorData();
    }
  }, [selectedMonitor]);

  const fetchMonitors = async () => {
    try {
      const response = await apiClient.get('/synthetics/monitors');
      const data = response.data || [];
      setMonitors(data);
      if (data.length > 0) {
        setSelectedMonitor(data[0]);
      }
    } catch (error) {
      console.error('Failed to fetch monitors:', error);
      setMonitors([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchMonitorData = async () => {
    if (!selectedMonitor) return;

    try {
      const end = new Date();
      const start = new Date();
      start.setDate(end.getDate() - 7); // Last 7 days

      // Fetch results
      const resultsResponse = await apiClient.get(`/synthetics/monitors/${selectedMonitor.id}/results?start=${start.toISOString()}&end=${end.toISOString()}&limit=100`);
      setResults(resultsResponse.data || []);

      // Fetch stats
      const statsResponse = await apiClient.get(`/synthetics/monitors/${selectedMonitor.id}/stats?start=${start.toISOString()}&end=${end.toISOString()}`);
      setStats(statsResponse.data || null);

    } catch (error) {
      console.error('Failed to fetch monitor data:', error);
    }
  };

  const createMonitor = async (monitorData: any) => {
    try {
      const response = await apiClient.post('/synthetics/monitors', monitorData);
      if (response.data) {
        setMonitors([...monitors, response.data]);
        setSelectedMonitor(response.data);
        setShowCreateModal(false);
      }
    } catch (error) {
      console.error('Failed to create monitor:', error);
      alert('Failed to create monitor');
    }
  };

  const toggleMonitor = async (monitor: SyntheticMonitor) => {
    try {
      const newStatus = monitor.status === 'enabled' ? 'disabled' : 'enabled';
      await apiClient.put(`/synthetics/monitors/${monitor.id}`, { status: newStatus });
      
      setMonitors(monitors.map(m => 
        m.id === monitor.id ? { ...m, status: newStatus } : m
      ));
      
      if (selectedMonitor?.id === monitor.id) {
        setSelectedMonitor({ ...selectedMonitor, status: newStatus });
      }
    } catch (error) {
      console.error('Failed to toggle monitor:', error);
    }
  };

  const getStatusColor = (success: boolean) => {
    return success ? 'success' : 'error';
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms.toFixed(0)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const CreateMonitorModal = () => {
    const [formData, setFormData] = useState({
      name: '',
      type: 'ping',
      url: '',
      frequency: 300,
      timeout: 30
    });

    const handleSubmit = (e: React.FormEvent) => {
      e.preventDefault();
      
      const config = {
        url: formData.url,
        timeout: formData.timeout
      };

      createMonitor({
        name: formData.name,
        type: formData.type,
        frequency: formData.frequency,
        config: JSON.stringify(config),
        status: 'enabled'
      });
    };

    if (!showCreateModal) return null;

    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg p-6 w-full max-w-md">
          <h3 className="text-lg font-bold mb-4">Create Synthetic Monitor</h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Monitor Name</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Monitor Type</label>
              <select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
              >
                <option value="ping">Ping</option>
                <option value="simple_browser">Simple Browser</option>
                <option value="api_test">API Test</option>
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">URL</label>
              <input
                type="url"
                value={formData.url}
                onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                placeholder="https://example.com"
                required
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Check Frequency (seconds)</label>
              <select
                value={formData.frequency}
                onChange={(e) => setFormData({ ...formData, frequency: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
              >
                <option value={60}>1 minute</option>
                <option value={300}>5 minutes</option>
                <option value={600}>10 minutes</option>
                <option value={1800}>30 minutes</option>
                <option value={3600}>1 hour</option>
              </select>
            </div>
            
            <div className="flex gap-3 pt-4">
              <Button type="submit" variant="primary" fullWidth>
                Create Monitor
              </Button>
              <Button 
                type="button" 
                variant="outline" 
                fullWidth
                onClick={() => setShowCreateModal(false)}
              >
                Cancel
              </Button>
            </div>
          </form>
        </div>
      </div>
    );
  };

  if (loading) {
    return (
      <MainLayout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2 flex items-center gap-2">
                <Globe className="w-8 h-8 text-blue-600" />
                Synthetic Monitoring
              </h1>
              <p className="text-gray-600">Monitor website uptime, performance, and user experience</p>
            </div>
            <Button variant="primary" onClick={() => setShowCreateModal(true)}>
              <Plus className="w-4 h-4" />
              Create Monitor
            </Button>
          </div>

          {/* Monitor Selector */}
          {monitors.length > 0 && (
            <div className="flex gap-2 flex-wrap">
              {monitors.map((monitor) => (
                <button
                  key={monitor.id}
                  onClick={() => setSelectedMonitor(monitor)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    selectedMonitor?.id === monitor.id
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${
                      monitor.status === 'enabled' ? 'bg-green-500' : 'bg-gray-400'
                    }`}></div>
                    {monitor.name}
                    <Badge variant="info" size="sm">{monitor.type}</Badge>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {monitors.length === 0 ? (
          // Empty State
          <Card>
            <div className="text-center py-12">
              <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Globe className="w-10 h-10 text-blue-600" />
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-2">No Monitors Yet</h3>
              <p className="text-gray-600 mb-6 max-w-md mx-auto">
                Start monitoring your websites and APIs by creating your first synthetic monitor. Track uptime, performance, and user experience.
              </p>
              <Button variant="primary" size="lg" onClick={() => setShowCreateModal(true)}>
                Create Your First Monitor
              </Button>
            </div>
          </Card>
        ) : selectedMonitor ? (
          <>
            {/* Monitor Controls */}
            <div className="mb-6">
              <Card>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div>
                      <h2 className="text-xl font-bold text-gray-900">{selectedMonitor.name}</h2>
                      <p className="text-sm text-gray-600">
                        {selectedMonitor.type} • Every {selectedMonitor.frequency / 60} minutes
                      </p>
                    </div>
                    <Badge 
                      variant={selectedMonitor.status === 'enabled' ? 'success' : 'warning'}
                      size="sm"
                    >
                      {selectedMonitor.status}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-3">
                    <Button
                      variant={selectedMonitor.status === 'enabled' ? 'error' : 'success'}
                      size="sm"
                      onClick={() => toggleMonitor(selectedMonitor)}
                    >
                      {selectedMonitor.status === 'enabled' ? (
                        <>
                          <Pause className="w-4 h-4" />
                          Pause
                        </>
                      ) : (
                        <>
                          <Play className="w-4 h-4" />
                          Resume
                        </>
                      )}
                    </Button>
                  </div>
                </div>
              </Card>
            </div>

            {/* Key Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              <Card>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Uptime</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {stats ? `${stats.uptime.toFixed(2)}%` : '0%'}
                    </p>
                  </div>
                  <div className="p-3 bg-green-100 rounded-full">
                    <CheckCircle className="w-6 h-6 text-green-600" />
                  </div>
                </div>
              </Card>

              <Card>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Avg Response Time</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {stats ? formatDuration(stats.avg_duration) : '0ms'}
                    </p>
                  </div>
                  <div className="p-3 bg-blue-100 rounded-full">
                    <Clock className="w-6 h-6 text-blue-600" />
                  </div>
                </div>
              </Card>

              <Card>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Total Checks</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {stats ? stats.total_runs.toLocaleString() : '0'}
                    </p>
                  </div>
                  <div className="p-3 bg-purple-100 rounded-full">
                    <TrendingUp className="w-6 h-6 text-purple-600" />
                  </div>
                </div>
              </Card>

              <Card>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Failed Checks</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {stats ? stats.failed_runs.toLocaleString() : '0'}
                    </p>
                  </div>
                  <div className="p-3 bg-red-100 rounded-full">
                    <AlertTriangle className="w-6 h-6 text-red-600" />
                  </div>
                </div>
              </Card>
            </div>

            {/* Response Time Chart */}
            <div className="mb-8">
              <Card>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-gray-900">Response Time (Last 7 Days)</h3>
                  <Badge variant="info" size="sm">ms</Badge>
                </div>
                <div className="h-64">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={results}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis 
                        dataKey="timestamp" 
                        tickFormatter={(value) => new Date(value).toLocaleDateString()}
                      />
                      <YAxis />
                      <Tooltip 
                        labelFormatter={(value) => new Date(value).toLocaleString()}
                        formatter={(value: number) => [formatDuration(value), 'Response Time']}
                      />
                      <Line 
                        type="monotone" 
                        dataKey="duration" 
                        stroke="#3b82f6" 
                        strokeWidth={2}
                        dot={(props) => {
                          const result = results.find(r => r.timestamp === props.payload.timestamp);
                          return (
                            <circle
                              cx={props.cx}
                              cy={props.cy}
                              r={3}
                              fill={result?.success ? '#10b981' : '#ef4444'}
                            />
                          );
                        }}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </Card>
            </div>

            {/* Recent Results */}
            <Card>
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-bold text-gray-900">Recent Check Results</h3>
                <Button variant="outline" size="sm">
                  View All Results
                </Button>
              </div>

              {results.length === 0 ? (
                <div className="text-center py-8">
                  <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                    <Clock className="w-8 h-8 text-gray-400" />
                  </div>
                  <h4 className="text-lg font-bold text-gray-900 mb-2">No Results Yet</h4>
                  <p className="text-gray-600">Check results will appear here once the monitor starts running.</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {results.slice(0, 10).map((result) => (
                    <div key={result.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                      <div className="flex items-center gap-3">
                        <div className={`w-3 h-3 rounded-full ${
                          result.success ? 'bg-green-500' : 'bg-red-500'
                        }`}></div>
                        <div>
                          <p className="text-sm font-medium text-gray-900">
                            {new Date(result.timestamp).toLocaleString()}
                          </p>
                          {result.error && (
                            <p className="text-xs text-red-600">{result.error}</p>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-4 text-sm text-gray-600">
                        <span>{formatDuration(result.duration)}</span>
                        <Badge 
                          variant={getStatusColor(result.success)} 
                          size="sm"
                        >
                          {result.status_code || (result.success ? 'Success' : 'Failed')}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </>
        ) : null}

        <CreateMonitorModal />
      </div>
    </MainLayout>
  );
};

export default Synthetics;