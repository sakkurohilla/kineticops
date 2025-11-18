import { useState, useEffect } from 'react';
import { Activity, RefreshCw, Server, Zap, AlertCircle } from 'lucide-react';
import MainLayout from '../../components/layout/MainLayout';
import Card from '../../components/common/Card';
import Badge from '../../components/common/Badge';
import Button from '../../components/common/Button';
import apiClient from '../../services/api/client';
import useWebSocket from '../../hooks/useWebsocket';

interface ServiceMetric {
  name: string;
  display_name: string;
  status: string;
  sub_status: string;
  cpu_percent: number;
  memory_percent: number;
  memory_mb: number;
  pid: number;
  restart_count: number;
  enabled: boolean;
}

interface Host {
  id: number;
  hostname: string;
}

function Services() {
  const [hosts, setHosts] = useState<Host[]>([]);
  const [selectedHost, setSelectedHost] = useState<Host | null>(null);
  const [cpuServices, setCpuServices] = useState<ServiceMetric[]>([]);
  const [memoryServices, setMemoryServices] = useState<ServiceMetric[]>([]);
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);

  // WebSocket handler for real-time service updates
  const handleWebSocketMessage = (data: any) => {
    if (!selectedHost) return;
    
    if (data.type === 'services' && data.host_id === selectedHost.id) {
      const cpuSvcs = data.services?.top_cpu || [];
      const memSvcs = data.services?.top_memory || [];
      
      if (cpuSvcs.length > 0) {
        setCpuServices(cpuSvcs);
      }
      if (memSvcs.length > 0) {
        setMemoryServices(memSvcs);
      }
    }
  };

  useWebSocket(handleWebSocketMessage);

  useEffect(() => {
    fetchHosts();
  }, []);

  useEffect(() => {
    if (selectedHost && autoRefresh) {
      const interval = setInterval(() => {
        // Services data comes via WebSocket, no need to poll
      }, 10000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, selectedHost]);

  const fetchHosts = async () => {
    try {
      const response = await apiClient.get('/hosts');
      const data = Array.isArray(response) ? response : (response?.data || []);
      setHosts(data);
      if (data.length > 0) {
        setSelectedHost(data[0]);
      }
    } catch (error) {
      console.error('Failed to fetch hosts:', error);
      setHosts([]);
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusLower = status.toLowerCase();
    
    const statusMap: { [key: string]: { variant: 'success' | 'warning' | 'error' | 'info', text: string } } = {
      'active': { variant: 'success', text: 'Active' },
      'inactive': { variant: 'info', text: 'Inactive' },
      'failed': { variant: 'error', text: 'Failed' },
      'activating': { variant: 'warning', text: 'Starting' },
      'deactivating': { variant: 'warning', text: 'Stopping' },
    };
    
    const badge = statusMap[statusLower] || { variant: 'info', text: status };
    return <Badge variant={badge.variant} size="sm">{badge.text}</Badge>;
  };

  const getSubStatusBadge = (subStatus: string) => {
    const statusLower = subStatus.toLowerCase();
    
    const statusMap: { [key: string]: { variant: 'success' | 'warning' | 'error' | 'info', text: string } } = {
      'running': { variant: 'success', text: 'Running' },
      'exited': { variant: 'info', text: 'Exited' },
      'dead': { variant: 'error', text: 'Dead' },
      'failed': { variant: 'error', text: 'Failed' },
    };
    
    const badge = statusMap[statusLower] || { variant: 'info', text: subStatus };
    return <Badge variant={badge.variant} size="sm">{badge.text}</Badge>;
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

  const uniqueServices = Array.from(new Set([...cpuServices, ...memoryServices].map(s => s.name)));
  const activeCount = uniqueServices.filter(name => {
    const svc = [...cpuServices, ...memoryServices].find(s => s.name === name);
    return svc && svc.status.toLowerCase() === 'active';
  }).length;
  const failedCount = uniqueServices.filter(name => {
    const svc = [...cpuServices, ...memoryServices].find(s => s.name === name);
    return svc && svc.status.toLowerCase() === 'failed';
  }).length;

  return (
    <MainLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2 flex items-center gap-3">
                <Server className="w-8 h-8 text-blue-600" />
                Services Monitoring
              </h1>
              <p className="text-gray-600">Top resource-consuming systemd services</p>
            </div>
            <div className="flex items-center gap-3">
              <Button
                variant={autoRefresh ? 'primary' : 'outline'}
                onClick={() => setAutoRefresh(!autoRefresh)}
                size="sm"
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
                {autoRefresh ? 'Live' : 'Paused'}
              </Button>
            </div>
          </div>

          {/* Host Selector */}
          {hosts.length > 0 && (
            <Card>
              <div className="flex items-center gap-3">
                <Server className="w-5 h-5 text-gray-500" />
                <select
                  className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  value={selectedHost?.id || ''}
                  onChange={(e) => {
                    const host = hosts.find(h => h.id === Number(e.target.value));
                    setSelectedHost(host || null);
                  }}
                >
                  {hosts.map(host => (
                    <option key={host.id} value={host.id}>
                      {host.hostname}
                    </option>
                  ))}
                </select>
              </div>
            </Card>
          )}
        </div>

        {/* Service Summary */}
        <Card className="mb-6">
          <div className="flex items-center justify-between flex-wrap gap-4">
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-3">
                <div className="p-3 bg-blue-50 rounded-xl">
                  <Server className="w-6 h-6 text-blue-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase tracking-wide font-medium">Total Services</p>
                  <p className="text-3xl font-bold text-gray-900 mt-1">
                    {uniqueServices.length}
                  </p>
                </div>
              </div>
              
              <div className="h-16 w-px bg-gray-200"></div>
              
              <div className="flex items-center gap-3">
                <div className="p-3 bg-green-50 rounded-xl">
                  <div className="w-6 h-6 flex items-center justify-center">
                    <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  </div>
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase tracking-wide font-medium">Active</p>
                  <p className="text-3xl font-bold text-green-600 mt-1">{activeCount}</p>
                </div>
              </div>
              
              <div className="h-16 w-px bg-gray-200"></div>
              
              <div className="flex items-center gap-3">
                <div className="p-3 bg-red-50 rounded-xl">
                  <AlertCircle className="w-6 h-6 text-red-600" />
                </div>
                <div>
                  <p className="text-xs text-gray-500 uppercase tracking-wide font-medium">Failed</p>
                  <p className="text-3xl font-bold text-red-600 mt-1">{failedCount}</p>
                </div>
              </div>
            </div>
            
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-sm text-gray-600">Last updated: {new Date().toLocaleTimeString()}</span>
            </div>
          </div>
        </Card>

        {/* Services Tables */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
          {/* Top Services by CPU */}
          <Card>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                <Zap className="w-5 h-5 text-orange-500" />
                Top Services by CPU
              </h2>
              <Badge variant="info" size="sm">Top {cpuServices.length}</Badge>
            </div>
            
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50 border-b border-gray-200">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Service</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CPU%</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Memory</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Restarts</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {cpuServices.map((service, idx) => (
                    <tr key={`${service.name}-${idx}`} className="hover:bg-gray-50 transition-colors">
                      <td className="px-4 py-3">
                        <div>
                          <p className="text-sm font-medium text-gray-900">{service.name}</p>
                          <p className="text-xs text-gray-500 truncate max-w-xs">{service.display_name}</p>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex flex-col gap-1">
                          {getStatusBadge(service.status)}
                          {getSubStatusBadge(service.sub_status)}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <div className="w-16 bg-gray-200 rounded-full h-2">
                            <div
                              className="bg-orange-500 h-2 rounded-full transition-all duration-500"
                              style={{ width: `${Math.min(service.cpu_percent, 100)}%` }}
                            ></div>
                          </div>
                          <span className="text-sm font-medium text-gray-900">{service.cpu_percent.toFixed(1)}%</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-900">{service.memory_mb.toFixed(1)} MB</td>
                      <td className="px-4 py-3 text-sm text-gray-900">{service.restart_count}</td>
                    </tr>
                  ))}
                  {cpuServices.length === 0 && (
                    <tr>
                      <td colSpan={5} className="px-4 py-8 text-center text-gray-500">
                        No service data available
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </Card>

          {/* Top Services by Memory */}
          <Card>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                <Activity className="w-5 h-5 text-green-500" />
                Top Services by Memory
              </h2>
              <Badge variant="info" size="sm">Top {memoryServices.length}</Badge>
            </div>
            
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50 border-b border-gray-200">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Service</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Memory%</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Memory</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Restarts</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {memoryServices.map((service, idx) => (
                    <tr key={`${service.name}-${idx}`} className="hover:bg-gray-50 transition-colors">
                      <td className="px-4 py-3">
                        <div>
                          <p className="text-sm font-medium text-gray-900">{service.name}</p>
                          <p className="text-xs text-gray-500 truncate max-w-xs">{service.display_name}</p>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex flex-col gap-1">
                          {getStatusBadge(service.status)}
                          {getSubStatusBadge(service.sub_status)}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <div className="w-16 bg-gray-200 rounded-full h-2">
                            <div
                              className="bg-green-500 h-2 rounded-full transition-all duration-500"
                              style={{ width: `${Math.min(service.memory_percent, 100)}%` }}
                            ></div>
                          </div>
                          <span className="text-sm font-medium text-gray-900">{service.memory_percent.toFixed(1)}%</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-900">{service.memory_mb.toFixed(1)} MB</td>
                      <td className="px-4 py-3 text-sm text-gray-900">{service.restart_count}</td>
                    </tr>
                  ))}
                  {memoryServices.length === 0 && (
                    <tr>
                      <td colSpan={5} className="px-4 py-8 text-center text-gray-500">
                        No service data available
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </Card>
        </div>
      </div>
    </MainLayout>
  );
}

export default Services;
