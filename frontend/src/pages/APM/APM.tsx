import React, { useState, useEffect } from 'react';
import { Activity, AlertTriangle, TrendingUp, Clock, Users, Zap } from 'lucide-react';
import MainLayout from '../../components/layout/MainLayout';
import Card from '../../components/common/Card';
import Button from '../../components/common/Button';
import Badge from '../../components/common/Badge';
import { LineChart, Line, AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import apiClient from '../../services/api/client';

interface Application {
  id: number;
  name: string;
  type: string;
  language: string;
  framework: string;
  status: string;
  last_seen: string;
}

interface TransactionStats {
  avg_response_time: number;
  throughput: number;
  error_rate: number;
  apdex: number;
  total_requests: number;
  error_count: number;
}

interface ErrorEvent {
  id: number;
  error_class: string;
  error_message: string;
  count: number;
  last_seen: string;
  resolved: boolean;
}

const APM: React.FC = () => {
  const [applications, setApplications] = useState<Application[]>([]);
  const [selectedApp, setSelectedApp] = useState<Application | null>(null);
  const [transactionStats, setTransactionStats] = useState<TransactionStats | null>(null);
  const [errors, setErrors] = useState<ErrorEvent[]>([]);
  const [performanceData, setPerformanceData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [timeRange, setTimeRange] = useState('24h');

  useEffect(() => {
    fetchApplications();
  }, []);

  useEffect(() => {
    if (selectedApp) {
      fetchApplicationData();
    }
  }, [selectedApp, timeRange]);

  const fetchApplications = async () => {
    try {
      const response = await apiClient.get('/apm/applications');
      const data = response.data || [];
      setApplications(data);
      if (data.length > 0) {
        setSelectedApp(data[0]);
      }
    } catch (error) {
      console.error('Failed to fetch applications:', error);
      setApplications([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchApplicationData = async () => {
    if (!selectedApp) return;

    try {
      const end = new Date();
      const start = new Date();
      
      switch (timeRange) {
        case '1h':
          start.setHours(end.getHours() - 1);
          break;
        case '6h':
          start.setHours(end.getHours() - 6);
          break;
        case '24h':
          start.setDate(end.getDate() - 1);
          break;
        case '7d':
          start.setDate(end.getDate() - 7);
          break;
        default:
          start.setDate(end.getDate() - 1);
      }

      // Fetch transaction stats
      const statsResponse = await apiClient.get(`/apm/applications/${selectedApp.id}/transactions/stats?start=${start.toISOString()}&end=${end.toISOString()}`);
      setTransactionStats(statsResponse.data || null);

      // Fetch errors
      const errorsResponse = await apiClient.get(`/apm/applications/${selectedApp.id}/errors?start=${start.toISOString()}&end=${end.toISOString()}&limit=10`);
      setErrors(errorsResponse.data || []);

      // Fetch performance metrics
      const metricsResponse = await apiClient.get(`/apm/applications/${selectedApp.id}/metrics?start=${start.toISOString()}&end=${end.toISOString()}`);
      setPerformanceData(metricsResponse.data || []);

    } catch (error) {
      console.error('Failed to fetch application data:', error);
    }
  };

  const createApplication = async () => {
    const name = prompt('Application name:');
    if (!name) return;

    const type = prompt('Application type (web/api/worker):') || 'web';
    const language = prompt('Programming language:') || 'javascript';

    try {
      const newApp = {
        name,
        type,
        language,
        framework: '',
        host_id: 1 // Default host
      };

      const response = await apiClient.post('/apm/applications', newApp);
      if (response.data) {
        setApplications([...applications, response.data]);
        setSelectedApp(response.data);
      }
    } catch (error) {
      console.error('Failed to create application:', error);
      alert('Failed to create application');
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms.toFixed(0)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatThroughput = (rpm: number) => {
    if (rpm < 1) return `${(rpm * 60).toFixed(1)}/min`;
    return `${rpm.toFixed(1)}/min`;
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
                <Zap className="w-8 h-8 text-blue-600" />
                Application Performance Monitoring
              </h1>
              <p className="text-gray-600">Monitor application performance, errors, and user experience</p>
            </div>
            <div className="flex items-center gap-4">
              {/* Time Range Selector */}
              <select
                value={timeRange}
                onChange={(e) => setTimeRange(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500"
              >
                <option value="1h">Last Hour</option>
                <option value="6h">Last 6 Hours</option>
                <option value="24h">Last 24 Hours</option>
                <option value="7d">Last 7 Days</option>
              </select>
              <Button variant="primary" onClick={createApplication}>
                Add Application
              </Button>
            </div>
          </div>

          {/* Application Selector */}
          {applications.length > 0 && (
            <div className="flex gap-2 flex-wrap">
              {applications.map((app) => (
                <button
                  key={app.id}
                  onClick={() => setSelectedApp(app)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    selectedApp?.id === app.id
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${
                      app.status === 'healthy' ? 'bg-green-500' : 
                      app.status === 'warning' ? 'bg-yellow-500' : 'bg-red-500'
                    }`}></div>
                    {app.name}
                    <Badge variant="info" size="sm">{app.language}</Badge>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {applications.length === 0 ? (
          // Empty State
          <Card>
            <div className="text-center py-12">
              <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Zap className="w-10 h-10 text-blue-600" />
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-2">No Applications Yet</h3>
              <p className="text-gray-600 mb-6 max-w-md mx-auto">
                Start monitoring your applications by adding your first application. Track performance, errors, and user experience.
              </p>
              <Button variant="primary" size="lg" onClick={createApplication}>
                Add Your First Application
              </Button>
            </div>
          </Card>
        ) : selectedApp ? (
          <>
            {/* Key Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              <Card>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Response Time</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {transactionStats ? formatDuration(transactionStats.avg_response_time) : '0ms'}
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
                    <p className="text-sm font-medium text-gray-600">Throughput</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {transactionStats ? formatThroughput(transactionStats.throughput) : '0/min'}
                    </p>
                  </div>
                  <div className="p-3 bg-green-100 rounded-full">
                    <TrendingUp className="w-6 h-6 text-green-600" />
                  </div>
                </div>
              </Card>

              <Card>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Error Rate</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {transactionStats ? `${transactionStats.error_rate.toFixed(2)}%` : '0%'}
                    </p>
                  </div>
                  <div className="p-3 bg-red-100 rounded-full">
                    <AlertTriangle className="w-6 h-6 text-red-600" />
                  </div>
                </div>
              </Card>

              <Card>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Apdex Score</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {transactionStats ? transactionStats.apdex.toFixed(2) : '0.00'}
                    </p>
                  </div>
                  <div className="p-3 bg-purple-100 rounded-full">
                    <Users className="w-6 h-6 text-purple-600" />
                  </div>
                </div>
              </Card>
            </div>

            {/* Charts */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
              {/* Response Time Chart */}
              <Card>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-gray-900">Response Time</h3>
                  <Badge variant="info" size="sm">ms</Badge>
                </div>
                <div className="h-64">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={performanceData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="timestamp" />
                      <YAxis />
                      <Tooltip />
                      <Line type="monotone" dataKey="response_time" stroke="#3b82f6" strokeWidth={2} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </Card>

              {/* Throughput Chart */}
              <Card>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-gray-900">Throughput</h3>
                  <Badge variant="success" size="sm">req/min</Badge>
                </div>
                <div className="h-64">
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={performanceData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="timestamp" />
                      <YAxis />
                      <Tooltip />
                      <Area type="monotone" dataKey="throughput" stroke="#10b981" fill="#10b981" fillOpacity={0.3} />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              </Card>

              {/* Error Rate Chart */}
              <Card>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-gray-900">Error Rate</h3>
                  <Badge variant="error" size="sm">%</Badge>
                </div>
                <div className="h-64">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={performanceData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="timestamp" />
                      <YAxis />
                      <Tooltip />
                      <Bar dataKey="error_rate" fill="#ef4444" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </Card>

              {/* Apdex Chart */}
              <Card>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-gray-900">Apdex Score</h3>
                  <Badge variant="info" size="sm">0-1</Badge>
                </div>
                <div className="h-64">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={performanceData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="timestamp" />
                      <YAxis domain={[0, 1]} />
                      <Tooltip />
                      <Line type="monotone" dataKey="apdex" stroke="#8b5cf6" strokeWidth={2} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </Card>
            </div>

            {/* Recent Errors */}
            <Card>
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-bold text-gray-900 flex items-center gap-2">
                  <AlertTriangle className="w-5 h-5 text-red-600" />
                  Recent Errors
                </h3>
                <Button variant="outline" size="sm">
                  View All Errors
                </Button>
              </div>

              {errors.length === 0 ? (
                <div className="text-center py-8">
                  <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                    <Activity className="w-8 h-8 text-green-600" />
                  </div>
                  <h4 className="text-lg font-bold text-gray-900 mb-2">No Errors Found</h4>
                  <p className="text-gray-600">Your application is running smoothly with no recent errors.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {errors.map((error) => (
                    <div key={error.id} className="flex items-start gap-4 p-4 bg-red-50 rounded-lg border border-red-200">
                      <div className="p-2 bg-red-100 rounded-full">
                        <AlertTriangle className="w-4 h-4 text-red-600" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <h4 className="font-medium text-gray-900">{error.error_class}</h4>
                          <Badge variant="error" size="sm">{error.count} occurrences</Badge>
                          {error.resolved && <Badge variant="success" size="sm">Resolved</Badge>}
                        </div>
                        <p className="text-sm text-gray-600 mb-2">{error.error_message}</p>
                        <p className="text-xs text-gray-500">Last seen: {new Date(error.last_seen).toLocaleString()}</p>
                      </div>
                      <Button variant="outline" size="sm">
                        View Details
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </>
        ) : null}
      </div>
    </MainLayout>
  );
};

export default APM;