import React, { useState, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import { Plus, Server, Search, RefreshCw, Users, Trash2, Download } from 'lucide-react';
import Button from '../../components/common/Button';
import Badge from '../../components/common/Badge';
import SimpleAddHostForm from '../../components/hosts/SimpleAddHostForm';
import { useHosts } from '../../hooks/useHosts';
// analytics moved to Dashboard
import { useNavigate } from 'react-router-dom';
import hostService from '../../services/api/hostService';
import { formatTimestamp } from '../../utils/dateUtils';
import { downloadCSV, downloadJSON, formatHostsForExport } from '../../utils/exportUtils';

const Hosts: React.FC = () => {
  const { hosts, loading, error, refetch } = useHosts();
  const [showAddForm, setShowAddForm] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedGroup, setSelectedGroup] = useState<string>('all');
  const [refreshing, setRefreshing] = useState(false);
  const [hostMetrics, setHostMetrics] = useState<Record<number, any>>({});
  const navigate = useNavigate();
  const [selectedHostId, setSelectedHostId] = useState<number | null>(null);

  const handleRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setTimeout(() => setRefreshing(false), 500);
  };

  const handleDelete = async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!window.confirm('Are you sure you want to delete this host?')) {
      return;
    }

    try {
      await hostService.deleteHost(id);
      refetch();
    } catch (err: any) {
      alert('Failed to delete host: ' + err.message);
    }
  };

  const handleExportCSV = () => {
    const exportData = formatHostsForExport(filteredHosts);
    const timestamp = new Date().toISOString().split('T')[0];
    downloadCSV(exportData, `hosts-export-${timestamp}`);
  };

  const handleExportJSON = () => {
    const timestamp = new Date().toISOString().split('T')[0];
    downloadJSON(filteredHosts, `hosts-export-${timestamp}`);
  };

  // Fetch metrics for hosts
  useEffect(() => {
    const fetchMetrics = async () => {
      const metricsPromises = hosts.map(async (host) => {
        try {
          const metrics = await hostService.getLatestMetrics(host.id);
          return { hostId: host.id, metrics };
        } catch (err) {
          return { hostId: host.id, metrics: null };
        }
      });

      const results = await Promise.all(metricsPromises);
      const metricsMap: Record<number, any> = {};
      results.forEach(({ hostId, metrics }) => {
        if (metrics) {
          metricsMap[hostId] = metrics;
        }
      });
      setHostMetrics(metricsMap);
    };

    if (hosts.length > 0) {
      fetchMetrics();
      // default select first host if none selected
      if (!selectedHostId) setSelectedHostId(hosts[0].id);
    }
  }, [hosts]);

  // When selected host changes, ensure we fetch fresh latest metrics for that host
  useEffect(() => {
    const refreshSelected = async () => {
      if (!selectedHostId) return;
      try {
        const latest = await hostService.getLatestMetrics(selectedHostId);
        setHostMetrics((prev) => ({ ...(prev || {}), [selectedHostId]: latest }));
      } catch (e) {
        // ignore
      }
    };
    refreshSelected();
  }, [selectedHostId]);

  // analytics moved to Dashboard; Hosts page keeps list only

  // Get unique groups
  const groups = ['all', ...new Set(hosts.map(h => h.group || 'default'))];

  // Filter hosts
  const filteredHosts = hosts.filter(host => {
    const matchesSearch = host.hostname?.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         host.ip?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesGroup = selectedGroup === 'all' || (host.group || 'default') === selectedGroup;
    return matchesSearch && matchesGroup;
  });

  return (
    <MainLayout>
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Hosts Management</h1>
            <p className="text-gray-600">Monitor and manage your infrastructure</p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={handleRefresh} disabled={refreshing}>
              <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            <div className="relative group">
              <Button variant="outline">
                <Download className="w-4 h-4" />
                Export
              </Button>
              <div className="absolute right-0 mt-1 w-36 bg-white rounded-lg shadow-lg border border-gray-200 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-10">
                <button
                  onClick={handleExportCSV}
                  className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-t-lg"
                >
                  Export as CSV
                </button>
                <button
                  onClick={handleExportJSON}
                  className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-b-lg"
                >
                  Export as JSON
                </button>
              </div>
            </div>
            <Button variant="primary" onClick={() => setShowAddForm(true)}>
              <Plus className="w-4 h-4" />
              Add Host
            </Button>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
          <div className="bg-gradient-to-br from-blue-500 to-blue-700 rounded-lg p-3 text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-blue-100">Total Hosts</p>
                <p className="text-xl font-bold">{hosts.length}</p>
              </div>
              <Server className="w-5 h-5" />
            </div>
          </div>
          <div className="bg-gradient-to-br from-green-500 to-emerald-700 rounded-lg p-3 text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-green-100">Online</p>
                <p className="text-xl font-bold">{hosts.filter(h => h.agent_status === 'online').length}</p>
              </div>
              <div className="w-2 h-2 bg-green-300 rounded-full animate-pulse"></div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-red-500 to-red-700 rounded-lg p-3 text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-red-100">Offline</p>
                <p className="text-xl font-bold">{hosts.filter(h => h.agent_status !== 'online').length}</p>
              </div>
              <div className="w-2 h-2 bg-red-300 rounded-full"></div>
            </div>
          </div>
          <div className="bg-gradient-to-br from-purple-500 to-indigo-700 rounded-lg p-3 text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-purple-100">Groups</p>
                <p className="text-xl font-bold">{groups.length - 1}</p>
              </div>
              <Users className="w-5 h-5" />
            </div>
          </div>
        </div>

        {/* Filters */}
        <div className="flex flex-col md:flex-row gap-4">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="Search hosts..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
            />
          </div>
          <div className="flex gap-2 flex-wrap">
            {groups.map(group => (
              <button
                key={group}
                onClick={() => setSelectedGroup(group)}
                className={`px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                  selectedGroup === group
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {group === 'all' ? 'All Groups' : group}
                {group !== 'all' && (
                  <span className="ml-1 text-xs opacity-75">
                    ({hosts.filter(h => (h.group || 'default') === group).length})
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Loading State */}
        {loading && (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        )}

        {/* Error State */}
        {error && !loading && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-center">
            <p className="text-red-800 text-sm">{error}</p>
          </div>
        )}

        {/* Empty State */}
        {!loading && !error && hosts.length === 0 && (
          <div className="bg-white rounded-lg border border-gray-200 p-12 text-center">
            <Server className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-bold text-gray-900 mb-2">No Hosts Yet</h3>
            <p className="text-gray-600 mb-4">Add your first host to start monitoring</p>
            <Button variant="primary" onClick={() => setShowAddForm(true)}>
              <Plus className="w-4 h-4" />
              Add Host
            </Button>
          </div>
        )}

        {/* Hosts Grid - Dashboard Style with Analytics Sidebar */}
        {!loading && !error && filteredHosts.length > 0 && (
          <div className="flex gap-3">
            <div className="flex-1">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-3 gap-3">
                {filteredHosts.map((host) => {
              const metrics = hostMetrics[host.id];
              const cpuUsage = metrics?.cpu_usage || 0;
              const memoryUsage = metrics?.memory_usage || 0;
              const diskUsage = metrics?.disk_usage || 0;
              const isOnline = host.agent_status === 'online';
              
              return (
                <div 
                  key={host.id} 
                  className="bg-gradient-to-br from-white to-gray-50 rounded-lg p-3 shadow-md hover:shadow-lg transition-all duration-300 hover:scale-105 cursor-pointer border border-gray-100 group"
                  onClick={() => setSelectedHostId(host.id)}
                  onDoubleClick={() => navigate(`/hosts/${host.id}`)}
                >
                  {/* Host Header */}
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center space-x-2">
                      <div className={`w-6 h-6 rounded-lg flex items-center justify-center ${
                        isOnline ? 'bg-gradient-to-br from-green-400 to-emerald-600' : 'bg-gradient-to-br from-gray-400 to-gray-600'
                      }`}>
                        <Server className="w-3 h-3 text-white" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <h3 className="font-bold text-gray-900 text-sm truncate">{host.hostname || host.ip}</h3>
                        <p className="text-xs text-gray-500 truncate">{host.ip}</p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <div className={`w-2 h-2 rounded-full ${
                        isOnline ? 'bg-green-500 animate-pulse' : 'bg-red-500'
                      }`}></div>
                      <button
                        onClick={(e) => handleDelete(host.id, e)}
                        className="opacity-0 group-hover:opacity-100 p-1 hover:bg-red-100 rounded transition-all"
                      >
                        <Trash2 className="w-3 h-3 text-red-500" />
                      </button>
                    </div>
                  </div>

                  {/* Group Badge */}
                  <div className="mb-3">
                    <Badge variant="info" size="sm">
                      {host.group || 'default'}
                    </Badge>
                  </div>

                  {/* Metrics Grid */}
                  <div className="grid grid-cols-3 gap-2 mb-3">
                    {/* CPU Circle */}
                    <div className="text-center">
                      <div className="relative w-8 h-8 mx-auto mb-1">
                        <svg className="w-8 h-8 transform -rotate-90" viewBox="0 0 36 36">
                          <path
                            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                            fill="none"
                            stroke="#e5e7eb"
                            strokeWidth="2"
                          />
                          <path
                            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                            fill="none"
                            stroke={cpuUsage > 80 ? '#ef4444' : cpuUsage > 60 ? '#f59e0b' : '#10b981'}
                            strokeWidth="2"
                            strokeDasharray={`${cpuUsage}, 100`}
                            strokeLinecap="round"
                            className="transition-all duration-1000"
                          />
                        </svg>
                        <div className="absolute inset-0 flex items-center justify-center">
                          <span className="text-xs font-bold text-gray-700">{cpuUsage.toFixed(1)}%</span>
                        </div>
                      </div>
                      <p className="text-xs font-medium text-gray-600">CPU</p>
                    </div>

                    {/* Memory Circle */}
                    <div className="text-center">
                      <div className="relative w-8 h-8 mx-auto mb-1">
                        <svg className="w-8 h-8 transform -rotate-90" viewBox="0 0 36 36">
                          <path
                            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                            fill="none"
                            stroke="#e5e7eb"
                            strokeWidth="2"
                          />
                          <path
                            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                            fill="none"
                            stroke={memoryUsage > 80 ? '#ef4444' : memoryUsage > 60 ? '#f59e0b' : '#3b82f6'}
                            strokeWidth="2"
                            strokeDasharray={`${memoryUsage}, 100`}
                            strokeLinecap="round"
                            className="transition-all duration-1000"
                          />
                        </svg>
                        <div className="absolute inset-0 flex items-center justify-center">
                          <span className="text-xs font-bold text-gray-700">{memoryUsage.toFixed(1)}%</span>
                        </div>
                      </div>
                      <p className="text-xs font-medium text-gray-600">RAM</p>
                    </div>

                    {/* Disk Circle */}
                    <div className="text-center">
                      <div className="relative w-8 h-8 mx-auto mb-1">
                        <svg className="w-8 h-8 transform -rotate-90" viewBox="0 0 36 36">
                          <path
                            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                            fill="none"
                            stroke="#e5e7eb"
                            strokeWidth="2"
                          />
                          <path
                            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                            fill="none"
                            stroke={diskUsage > 80 ? '#ef4444' : diskUsage > 60 ? '#f59e0b' : '#8b5cf6'}
                            strokeWidth="2"
                            strokeDasharray={`${diskUsage}, 100`}
                            strokeLinecap="round"
                            className="transition-all duration-1000"
                          />
                        </svg>
                        <div className="absolute inset-0 flex items-center justify-center">
                          <span className="text-xs font-bold text-gray-700">{diskUsage.toFixed(1)}%</span>
                        </div>
                      </div>
                      <p className="text-xs font-medium text-gray-600">Disk</p>
                    </div>
                  </div>

                  {/* Status and Last Seen */}
                  <div className="flex items-center justify-between pt-2 border-t border-gray-200">
                    <Badge 
                      variant={isOnline ? 'success' : 'error'}
                      size="sm"
                    >
                      {host.agent_status || 'offline'}
                    </Badge>
                    <span className="text-xs text-gray-500">
                      {host.last_seen ? formatTimestamp(host.last_seen).split(' ')[1] : 'Never'}
                    </span>
                  </div>
                </div>
              );
            })}
              </div>
            </div>

            {/* Right-side analytics removed (moved to Dashboard) */}
          </div>
        )}
        {/* No Search Results */}
        {!loading && !error && hosts.length > 0 && filteredHosts.length === 0 && (
          <div className="bg-white rounded-lg border border-gray-200 p-8 text-center">
            <p className="text-gray-600">No hosts found matching your criteria</p>
          </div>
        )}
      </div>

      {/* Add Host Modal */}
      {showAddForm && (
        <SimpleAddHostForm
          onClose={() => setShowAddForm(false)}
          onSuccess={() => {
            refetch();
          }}
        />
      )}
    </MainLayout>
  );
};

export default Hosts;