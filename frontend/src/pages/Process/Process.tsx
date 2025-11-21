import React, { useState, useEffect } from 'react';
import { Activity, Cpu, MemoryStick, Server, RefreshCw, ArrowUpDown } from 'lucide-react';
import MainLayout from '../../components/layout/MainLayout';
import Card from '../../components/common/Card';
import Button from '../../components/common/Button';
import Badge from '../../components/common/Badge';
import apiClient from '../../services/api/client';
import useWebSocket from '../../hooks/useWebsocket';

interface ProcessMetric {
  id: number;
  host_id: number;
  pid: number;
  name: string;
  username: string;
  cpu_percent: number;
  memory_percent: number;
  memory_rss: number;
  status: string;
  num_threads: number;
  create_time: number;
  timestamp: string;
}

interface Host {
  id: number;
  hostname: string;
  ip: string;
  status: string;
}

interface ServerAlert {
  type: 'cpu' | 'memory' | 'disk' | 'unresponsive';
  severity: 'critical' | 'warning' | 'info';
  message: string;
  process?: string;
  pid?: number;
  value?: number;
  threshold?: number;
  timestamp: Date;
}

interface HostMetrics {
  cpu_usage: number;
  memory_usage: number;
  memory_total: number;
  memory_used: number;
  memory_free: number;
  disk_usage: number;
  disk_total: number;
  disk_used: number;
}

type SortOrder = 'asc' | 'desc';

const Process: React.FC = () => {
  const [hosts, setHosts] = useState<Host[]>([]);
  const [selectedHost, setSelectedHost] = useState<Host | null>(null);
  const [cpuProcesses, setCpuProcesses] = useState<ProcessMetric[]>([]);
  const [memoryProcesses, setMemoryProcesses] = useState<ProcessMetric[]>([]);
  const [serverAlerts, setServerAlerts] = useState<ServerAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [sortByCPU, setSortByCPU] = useState<SortOrder>('desc');
  const [sortByMem, setSortByMem] = useState<SortOrder>('desc');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [viewMode, setViewMode] = useState<'cpu' | 'memory' | 'both'>('memory');

  // WebSocket handler for real-time process updates
  const handleWebSocketMessage = (data: any) => {
    if (!selectedHost) return;
    
    if (data.type === 'processes' && data.host_id === selectedHost.id) {
      const cpuProcs = data.processes?.top_cpu || [];
      const memProcs = data.processes?.top_memory || [];
      
      if (cpuProcs.length > 0) {
        setCpuProcesses(cpuProcs);
      }
      if (memProcs.length > 0) {
        setMemoryProcesses(memProcs);
      }
      
      // Generate alerts from WebSocket data
      generateServerAlerts(cpuProcs, memProcs, null);
    } else if (data.type === 'metric' && data.host_id === selectedHost.id) {
      // WebSocket provides system metrics - use them for alerts
      const systemMetrics: HostMetrics = {
        cpu_usage: data.cpu_usage || 0,
        memory_usage: data.memory_usage || 0,
        memory_total: data.memory_total || 0,
        memory_used: data.memory_used || 0,
        memory_free: (data.memory_total || 0) - (data.memory_used || 0),
        disk_usage: data.disk_usage || 0,
        disk_total: data.disk_total || 0,
        disk_used: data.disk_used || 0
      };
      // Generate alerts with system metrics
      generateServerAlerts(cpuProcesses, memoryProcesses, systemMetrics);
    }
  };

  useWebSocket(handleWebSocketMessage);

  useEffect(() => {
    fetchHosts();
  }, []);

  useEffect(() => {
    if (selectedHost) {
      fetchProcesses();
    }
  }, [selectedHost]);

  useEffect(() => {
    if (!autoRefresh || !selectedHost) return;
    
    const interval = setInterval(() => {
      fetchProcesses();
    }, 5000); // Refresh every 5 seconds

    return () => clearInterval(interval);
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

  const fetchProcesses = async () => {
    if (!selectedHost) return;

    try {
      // Fetch host metrics
      const metricsData: any = await apiClient.get(`/hosts/${selectedHost.id}/metrics/latest`);

      // Fetch top 10 CPU processes
      const cpuData: any = await apiClient.get(`/hosts/${selectedHost.id}/processes?sort=cpu&limit=10`);
      setCpuProcesses(cpuData?.processes || []);

      // Fetch top 10 memory processes
      const memData: any = await apiClient.get(`/hosts/${selectedHost.id}/processes?sort=memory&limit=10`);
      setMemoryProcesses(memData?.processes || []);

      // Generate server health alerts
      generateServerAlerts(cpuData?.processes || [], memData?.processes || [], metricsData);
    } catch (error) {
      console.error('Failed to fetch process metrics:', error);
    }
  };

  const generateServerAlerts = (cpuProcs: ProcessMetric[], memProcs: ProcessMetric[], metrics: HostMetrics | null) => {
    const alerts: ServerAlert[] = [];
    const now = new Date();

    if (metrics) {
      // Critical CPU Alert (>90%)
      if (metrics.cpu_usage > 90) {
        alerts.push({
          type: 'unresponsive',
          severity: 'critical',
          message: `Server Unresponsive - Critical CPU Load`,
          value: metrics.cpu_usage,
          threshold: 90,
          timestamp: now
        });
      } else if (metrics.cpu_usage > 80) {
        alerts.push({
          type: 'cpu',
          severity: 'critical',
          message: `Critical System CPU Usage`,
          value: metrics.cpu_usage,
          threshold: 80,
          timestamp: now
        });
      }

      // Critical Memory Alert (>90%)
      if (metrics.memory_usage > 90) {
        alerts.push({
          type: 'unresponsive',
          severity: 'critical',
          message: `Server Unresponsive - Out of Memory`,
          value: metrics.memory_usage,
          threshold: 90,
          timestamp: now
        });
      } else if (metrics.memory_usage > 80) {
        alerts.push({
          type: 'memory',
          severity: 'critical',
          message: `Critical System Memory Usage`,
          value: metrics.memory_usage,
          threshold: 80,
          timestamp: now
        });
      }

      // Critical Disk Alert (>95%)
      if (metrics.disk_usage > 95) {
        alerts.push({
          type: 'unresponsive',
          severity: 'critical',
          message: `Server Unresponsive - Disk Full`,
          value: metrics.disk_usage,
          threshold: 95,
          timestamp: now
        });
      } else if (metrics.disk_usage > 85) {
        alerts.push({
          type: 'disk',
          severity: 'warning',
          message: `Critical Disk Space Usage`,
          value: metrics.disk_usage,
          threshold: 85,
          timestamp: now
        });
      }
    }

    // Critical CPU process alerts (>80% = critical, >50% = warning)
    cpuProcs.filter(p => p.cpu_percent > 50).forEach(proc => {
      alerts.push({
        type: 'cpu',
        severity: proc.cpu_percent > 80 ? 'critical' : 'warning',
        message: proc.cpu_percent > 80 ? 'Critical Process CPU Usage' : 'High CPU Usage',
        process: proc.name,
        pid: proc.pid,
        value: proc.cpu_percent,
        threshold: proc.cpu_percent > 80 ? 80 : 50,
        timestamp: now
      });
    });

    // Critical Memory process alerts (>50% = critical, >30% = warning)
    memProcs.filter(p => p.memory_percent > 30).forEach(proc => {
      alerts.push({
        type: 'memory',
        severity: proc.memory_percent > 50 ? 'critical' : 'warning',
        message: proc.memory_percent > 50 ? 'Critical Process Memory Usage' : 'High Memory Usage',
        process: proc.name,
        pid: proc.pid,
        value: proc.memory_percent,
        threshold: proc.memory_percent > 50 ? 50 : 30,
        timestamp: now
      });
    });

    setServerAlerts(alerts);
  };

  const formatMemory = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  };

  const getStatusBadge = (status: string) => {
    const statusLower = status.toLowerCase();
    
    // Map status names to display
    const statusMap: { [key: string]: { variant: 'success' | 'warning' | 'error' | 'info', text: string } } = {
      'running': { variant: 'success', text: 'Running' },
      'sleep': { variant: 'info', text: 'Sleep' },
      'sleeping': { variant: 'info', text: 'Sleep' },
      'idle': { variant: 'info', text: 'Idle' },
      'stopped': { variant: 'error', text: 'Stopped' },
      'zombie': { variant: 'warning', text: 'Zombie' },
      'disk-sleep': { variant: 'info', text: 'D-Sleep' },
    };
    
    const badge = statusMap[statusLower] || { variant: 'info', text: status };
    return <Badge variant={badge.variant} size="sm">{badge.text}</Badge>;
  };

  const toggleSort = (table: 'cpu' | 'memory') => {
    if (table === 'cpu') {
      const newOrder = sortByCPU === 'desc' ? 'asc' : 'desc';
      setSortByCPU(newOrder);
      setCpuProcesses([...cpuProcesses].sort((a, b) => 
        newOrder === 'desc' ? b.cpu_percent - a.cpu_percent : a.cpu_percent - b.cpu_percent
      ));
    } else {
      const newOrder = sortByMem === 'desc' ? 'asc' : 'desc';
      setSortByMem(newOrder);
      setMemoryProcesses([...memoryProcesses].sort((a, b) => 
        newOrder === 'desc' ? b.memory_percent - a.memory_percent : a.memory_percent - b.memory_percent
      ));
    }
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
        <div className="mb-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2 flex items-center gap-3">
                <Activity className="w-8 h-8 text-blue-600" />
                Process Monitoring
              </h1>
              <p className="text-gray-600">Real-time process monitoring and system resource analysis</p>
            </div>
            <div className="flex items-center gap-3">
              <Button
                variant={autoRefresh ? 'primary' : 'outline'}
                onClick={() => setAutoRefresh(!autoRefresh)}
                size="sm"
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
                {autoRefresh ? 'Auto Refresh ON' : 'Auto Refresh OFF'}
              </Button>
            </div>
          </div>

          {/* Host Selector - Compact */}
          {hosts.length > 0 && (
            <div className="flex items-center gap-3 p-4 bg-gradient-to-r from-blue-50 to-purple-50 rounded-xl border border-blue-100 flex-wrap">
              <Server className="w-5 h-5 text-blue-600" />
              <span className="text-sm font-medium text-gray-700">Host:</span>
              <select
                className="px-3 py-1.5 bg-white border border-gray-300 rounded-lg text-sm font-medium text-gray-900 shadow-sm hover:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                value={selectedHost?.id || ''}
                onChange={(e) => {
                  const host = hosts.find(h => h.id === Number(e.target.value));
                  setSelectedHost(host || null);
                }}
              >
                {hosts.map(host => (
                  <option key={host.id} value={host.id}>
                    {host.hostname} ({host.ip})
                  </option>
                ))}
              </select>
              
              <div className="h-6 w-px bg-gray-300 ml-2"></div>
              
              <span className="text-sm font-medium text-gray-700 ml-2">View:</span>
              <select
                className="px-3 py-1.5 bg-white border border-gray-300 rounded-lg text-sm font-medium text-gray-900 shadow-sm hover:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                value={viewMode}
                onChange={(e) => setViewMode(e.target.value as 'cpu' | 'memory' | 'both')}
              >
                <option value="both">Both (Side by Side)</option>
                <option value="cpu">CPU Usage Only</option>
                <option value="memory">Memory Usage Only</option>
              </select>
            </div>
          )}
        </div>

        {hosts.length === 0 ? (
          // Empty State
          <Card>
            <div className="text-center py-12">
              <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Server className="w-10 h-10 text-blue-600" />
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-2">No Hosts Connected</h3>
              <p className="text-gray-600 mb-4 max-w-md mx-auto">
                Connect a host with the KineticOps agent to start monitoring processes and system resources.
              </p>
              <Button variant="outline" onClick={fetchHosts} size="sm">
                <RefreshCw className="w-4 h-4 mr-2" />
                Retry Loading Hosts
              </Button>
            </div>
          </Card>
        ) : selectedHost ? (
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            {/* Main Content - 3 columns */}
            <div className="lg:col-span-3 space-y-6">
            {/* Process Tables */}
            <div className={`grid grid-cols-1 gap-6 mb-6 ${viewMode === 'both' ? 'lg:grid-cols-2' : ''}`}>
              {/* Top Processes by CPU */}
              {(viewMode === 'cpu' || viewMode === 'both') && (
              <Card>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-gray-900 flex items-center gap-2">
                    <div className="p-2 bg-blue-50 rounded-lg">
                      <Cpu className="w-5 h-5 text-blue-600" />
                    </div>
                    <span>Top Processes by CPU</span>
                  </h3>
                  <Badge variant="info" size="sm">Top 10</Badge>
                </div>

                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50">
                      <tr className="border-b-2 border-gray-200">
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">PID</th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">Process</th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">User</th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider cursor-pointer hover:text-blue-600"
                            onClick={() => toggleSort('cpu')}>
                          <div className="flex items-center gap-1">
                            CPU%
                            <ArrowUpDown className="w-3 h-3" />
                          </div>
                        </th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">Memory</th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {cpuProcesses.length === 0 ? (
                        <tr>
                          <td colSpan={6} className="text-center py-8 text-gray-500">
                            No process data available
                          </td>
                        </tr>
                      ) : (
                        cpuProcesses.map((proc, idx) => (
                          <tr key={`${proc.pid}-${idx}`} className="border-b border-gray-100 hover:bg-blue-50/50 transition-colors">
                            <td className="py-1 px-2 font-mono text-xs text-gray-700 font-semibold">{proc.pid}</td>
                            <td className="py-1 px-2">
                              <div className="font-medium text-gray-900 text-xs truncate max-w-[150px]" title={proc.name}>
                                {proc.name}
                              </div>
                            </td>
                            <td className="py-1 px-2 text-gray-600 text-xs truncate max-w-[80px]">{proc.username}</td>
                            <td className="py-1 px-2">
                              <div className="flex items-center gap-2">
                                <div className="flex-1 bg-gray-200 rounded-full h-2 shadow-inner min-w-[60px]">
                                  <div
                                    className={`h-2 rounded-full transition-all ${
                                      proc.cpu_percent > 80 ? 'bg-gradient-to-r from-red-500 to-red-600' :
                                      proc.cpu_percent > 50 ? 'bg-gradient-to-r from-yellow-500 to-yellow-600' : 
                                      'bg-gradient-to-r from-blue-500 to-blue-600'
                                    }`}
                                    style={{ width: `${Math.min(proc.cpu_percent, 100)}%` }}
                                  ></div>
                                </div>
                                <span className="font-semibold text-gray-900 text-xs whitespace-nowrap w-[45px] text-right">
                                  {proc.cpu_percent.toFixed(1)}%
                                </span>
                              </div>
                            </td>
                            <td className="py-1 px-2 text-xs text-gray-600 font-medium whitespace-nowrap">
                              {formatMemory(proc.memory_rss)}
                            </td>
                            <td className="py-1 px-2">
                              {getStatusBadge(proc.status)}
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </Card>
              )}

              {/* Top Processes by Memory */}
              {(viewMode === 'memory' || viewMode === 'both') && (
              <Card>
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-gray-900 flex items-center gap-2">
                    <div className="p-2 bg-purple-50 rounded-lg">
                      <MemoryStick className="w-5 h-5 text-purple-600" />
                    </div>
                    <span>Top Processes by Memory</span>
                  </h3>
                  <Badge variant="info" size="sm">Top 10</Badge>
                </div>

                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50">
                      <tr className="border-b-2 border-gray-200">
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">PID</th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">Process</th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">User</th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">CPU%</th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider cursor-pointer hover:text-purple-600"
                            onClick={() => toggleSort('memory')}>
                          <div className="flex items-center gap-1">
                            Memory%
                            <ArrowUpDown className="w-3 h-3" />
                          </div>
                        </th>
                        <th className="text-left py-1 px-2 font-semibold text-gray-700 text-xs uppercase tracking-wider">Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {memoryProcesses.length === 0 ? (
                        <tr>
                          <td colSpan={6} className="text-center py-8 text-gray-500">
                            No process data available
                          </td>
                        </tr>
                      ) : (
                        memoryProcesses.map((proc, idx) => (
                          <tr key={`${proc.pid}-${idx}`} className="border-b border-gray-100 hover:bg-purple-50/50 transition-colors">
                            <td className="py-1 px-2 font-mono text-xs text-gray-700 font-semibold">{proc.pid}</td>
                            <td className="py-1 px-2">
                              <div className="font-medium text-gray-900 text-xs truncate max-w-[150px]" title={proc.name}>
                                {proc.name}
                              </div>
                            </td>
                            <td className="py-1 px-2 text-gray-600 text-xs truncate max-w-[80px]">{proc.username}</td>
                            <td className="py-1 px-2 text-xs text-gray-600 font-medium whitespace-nowrap">
                              {proc.cpu_percent.toFixed(1)}%
                            </td>
                            <td className="py-1 px-2">
                              <div className="flex items-center gap-2">
                                <div className="flex-1 bg-gray-200 rounded-full h-2 shadow-inner min-w-[60px]">
                                  <div
                                    className={`h-2 rounded-full transition-all ${
                                      proc.memory_percent > 80 ? 'bg-gradient-to-r from-red-500 to-red-600' :
                                      proc.memory_percent > 50 ? 'bg-gradient-to-r from-yellow-500 to-yellow-600' : 
                                      'bg-gradient-to-r from-purple-500 to-purple-600'
                                    }`}
                                    style={{ width: `${Math.min(proc.memory_percent, 100)}%` }}
                                  ></div>
                                </div>
                                <span className="font-semibold text-gray-900 text-xs whitespace-nowrap w-[45px] text-right">
                                  {proc.memory_percent.toFixed(1)}%
                                </span>
                              </div>
                              <div className="text-xs text-gray-500 font-medium mt-0.5 whitespace-nowrap">{formatMemory(proc.memory_rss)}</div>
                            </td>
                            <td className="py-1 px-2">
                              {getStatusBadge(proc.status)}
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </Card>
              )}
            </div>

            {/* Process Summary */}
            <Card>
                <div className="flex items-center justify-between flex-wrap gap-4 lg:gap-6">
                <div className="flex items-center gap-4 md:gap-6 lg:gap-8 flex-wrap">
                  <div className="flex items-center gap-3">
                    <div className="p-2 md:p-3 bg-blue-50 rounded-xl">
                      <Activity className="w-5 h-5 md:w-6 md:h-6 text-blue-600" />
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 uppercase tracking-wide font-medium">Total</p>
                      <p className="text-2xl md:text-3xl font-bold text-gray-900 mt-0.5 md:mt-1">
                        {new Set([...cpuProcesses, ...memoryProcesses].map(p => p.pid)).size}
                      </p>
                    </div>
                  </div>
                  
                  <div className="h-12 md:h-16 w-px bg-gray-200"></div>
                  
                  <div className="flex items-center gap-3">
                    <div className="p-3 bg-green-50 rounded-xl">
                      <div className="w-6 h-6 flex items-center justify-center">
                        <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                      </div>
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 uppercase tracking-wide font-medium">Running</p>
                      <p className="text-3xl font-bold text-green-600 mt-1">
                        {Array.from(new Set([...cpuProcesses, ...memoryProcesses].map(p => p.pid))).filter(pid => {
                          const proc = [...cpuProcesses, ...memoryProcesses].find(p => p.pid === pid);
                          const status = proc?.status.toLowerCase() || '';
                          return proc && (status === 'running' || status === 'r');
                        }).length}
                      </p>
                    </div>
                  </div>
                  
                  <div className="h-16 w-px bg-gray-200"></div>
                  
                  <div className="flex items-center gap-3">
                    <div className="p-3 bg-blue-50 rounded-xl">
                      <div className="w-6 h-6 flex items-center justify-center">
                        <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
                      </div>
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 uppercase tracking-wide font-medium">Sleeping</p>
                      <p className="text-3xl font-bold text-blue-600 mt-1">
                        {Array.from(new Set([...cpuProcesses, ...memoryProcesses].map(p => p.pid))).filter(pid => {
                          const proc = [...cpuProcesses, ...memoryProcesses].find(p => p.pid === pid);
                          const status = proc?.status.toLowerCase() || '';
                          return proc && (status === 'sleep' || status === 'sleeping' || status === 'idle' || status === 's' || status === 'd' || status === 'i');
                        }).length}
                      </p>
                    </div>
                  </div>
                  
                  <div className="h-16 w-px bg-gray-200"></div>
                  
                  <div className="flex items-center gap-3">
                    <div className="p-3 bg-purple-50 rounded-xl">
                      <div className="w-6 h-6 flex items-center justify-center text-purple-600 font-bold text-sm">
                        T
                      </div>
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 uppercase tracking-wide font-medium">Total Threads</p>
                      <p className="text-3xl font-bold text-purple-600 mt-1">
                        {Array.from(new Set([...cpuProcesses, ...memoryProcesses].map(p => p.pid))).reduce((sum, pid) => {
                          const proc = [...cpuProcesses, ...memoryProcesses].find(p => p.pid === pid);
                          return sum + (proc?.num_threads || 0);
                        }, 0)}
                      </p>
                    </div>
                  </div>
                </div>
                
                <div className="flex items-center gap-2 px-4 py-2 bg-gray-50 rounded-lg">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                  <div className="text-right">
                    <p className="text-xs text-gray-500 font-medium">Last updated</p>
                    <p className="text-sm font-semibold text-gray-700">{new Date().toLocaleTimeString()}</p>
                  </div>
                </div>
              </div>
            </Card>
            </div>

            {/* Trigger Alerts Sidebar - 1 column */}
            <div className="lg:col-span-1">
              <Card>
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-base font-bold text-gray-900 flex items-center gap-2">
                    <div className="p-1.5 bg-red-50 rounded-lg">
                      <svg className="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                      </svg>
                    </div>
                    <span className="text-sm">System Alerts</span>
                  </h3>
                  <Badge variant={serverAlerts.length > 0 ? 'error' : 'success'} size="sm">
                    {serverAlerts.length}
                  </Badge>
                </div>

                <div className="space-y-2">
                  {/* Server Alerts */}
                  {serverAlerts.map((alert, idx) => {
                    const bgColor = alert.severity === 'critical' 
                      ? (alert.type === 'unresponsive' ? 'bg-red-100 border-red-300' : 'bg-red-50 border-red-200')
                      : (alert.type === 'disk' ? 'bg-orange-50 border-orange-200' : 'bg-yellow-50 border-yellow-200');
                    const textColor = alert.severity === 'critical' ? 'text-red-900' : 'text-yellow-900';
                    const dotColor = alert.severity === 'critical' 
                      ? (alert.type === 'unresponsive' ? 'bg-red-600' : 'bg-red-500')
                      : 'bg-yellow-500';

                    return (
                      <div key={`alert-${idx}`} className={`p-2 border rounded-lg ${bgColor} ${alert.type === 'unresponsive' ? 'ring-1 ring-red-500' : ''}`}>
                        <div className="flex items-start gap-2">
                          <div className={`flex-shrink-0 w-1.5 h-1.5 ${dotColor} rounded-full mt-1 ${alert.type === 'unresponsive' ? 'animate-pulse' : ''}`}></div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between gap-1 mb-0.5">
                              <p className={`text-xs font-semibold ${textColor} uppercase tracking-wide leading-tight`}>
                                {alert.type === 'unresponsive' ? '🚨 UNRESPONSIVE' : alert.message}
                              </p>
                              <span className="text-xs text-gray-500 whitespace-nowrap">
                                {alert.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                              </span>
                            </div>
                            {alert.process && (
                              <p className="text-xs font-medium text-gray-900 truncate" title={alert.process}>
                                {alert.process}
                              </p>
                            )}
                            <div className="flex items-center gap-1.5 mt-1 flex-wrap">
                              {alert.pid && (
                                <>
                                  <span className="text-xs text-gray-600">PID: {alert.pid}</span>
                                  <span className="text-xs text-gray-400">•</span>
                                </>
                              )}
                              {alert.value !== undefined && (
                                <span className={`text-xs font-bold ${alert.severity === 'critical' ? 'text-red-600' : 'text-yellow-600'}`}>
                                  {alert.value.toFixed(1)}%
                                </span>
                              )}

                            </div>
                            {alert.type === 'unresponsive' && (
                              <div className="mt-2 p-2 bg-white/50 rounded text-xs text-gray-700">
                                <strong>Action Required:</strong> Server may be unresponsive. Check system resources immediately.
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    );
                  })}

                  {/* No Alerts */}
                  {serverAlerts.length === 0 && (
                    <div className="text-center py-8">
                      <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-3">
                        <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                      </div>
                      <p className="text-sm font-medium text-gray-900 mb-1">All Systems Healthy</p>
                      <p className="text-xs text-gray-500">No critical alerts detected</p>
                    </div>
                  )}
                </div>
              </Card>
            </div>
          </div>
        ) : null}
      </div>
    </MainLayout>
  );
};

export default Process;
