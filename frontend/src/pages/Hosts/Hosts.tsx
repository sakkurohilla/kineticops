import React, { useState, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import { Plus, Server, Search, RefreshCw, Users } from 'lucide-react';
import Button from '../../components/common/Button';
import SimpleAddHostForm from '../../components/hosts/SimpleAddHostForm';
import { useHosts } from '../../hooks/useHosts';
// analytics moved to Dashboard
import { useNavigate } from 'react-router-dom';
import hostService from '../../services/api/hostService';
import { formatTimestamp } from '../../utils/dateUtils';

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
            <Button variant="primary" onClick={() => setShowAddForm(true)}>
              <Plus className="w-4 h-4" />
              Add Host
            </Button>
          </div>
        </div>

        {/* Stats - Clean Professional Style */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-blue-500/10 to-cyan-500/10 backdrop-blur-lg border border-white/20 shadow-lg p-4">
            <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none" />
            <div className="relative z-10 flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-gray-600 mb-1">Total Hosts</p>
                <p className="text-2xl font-bold text-gray-900">{hosts.length}</p>
              </div>
              <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-cyan-500 rounded-lg flex items-center justify-center shadow-md">
                <Server className="w-5 h-5 text-white" />
              </div>
            </div>
          </div>
          <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-emerald-500/10 to-teal-500/10 backdrop-blur-lg border border-white/20 shadow-lg p-4">
            <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none" />
            <div className="relative z-10 flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-gray-600 mb-1">Online</p>
                <p className="text-2xl font-bold text-gray-900">{hosts.filter(h => h.agent_status === 'online').length}</p>
              </div>
              <div className="w-3 h-3 bg-emerald-500 rounded-full animate-pulse shadow-lg" />
            </div>
          </div>
          <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-red-500/10 to-pink-500/10 backdrop-blur-lg border border-white/20 shadow-lg p-4">
            <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none" />
            <div className="relative z-10 flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-gray-600 mb-1">Offline</p>
                <p className="text-2xl font-bold text-gray-900">{hosts.filter(h => h.agent_status !== 'online').length}</p>
              </div>
              <div className="w-3 h-3 bg-red-500 rounded-full shadow-lg" />
            </div>
          </div>
          <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-purple-500/10 to-indigo-500/10 backdrop-blur-lg border border-white/20 shadow-lg p-4">
            <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none" />
            <div className="relative z-10 flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-gray-600 mb-1">Groups</p>
                <p className="text-2xl font-bold text-gray-900">{groups.length - 1}</p>
              </div>
              <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-indigo-500 rounded-lg flex items-center justify-center shadow-md">
                <Users className="w-5 h-5 text-white" />
              </div>
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

        {/* Hosts Grid - Clean Professional Cards */}
        {!loading && !error && filteredHosts.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredHosts.map((host) => {
              const metrics = hostMetrics[host.id];
              const cpuUsage = metrics?.cpu_usage || 0;
              const memoryUsage = metrics?.memory_usage || 0;
              const diskUsage = metrics?.disk_usage || 0;
              const isOnline = host.agent_status === 'online';
              
              return (
                <div 
                  key={host.id} 
                  className="relative overflow-hidden rounded-xl bg-white backdrop-blur-lg border border-gray-200 shadow-md hover:shadow-xl transition-all duration-300 cursor-pointer"
                  onClick={() => navigate(`/hosts/${host.id}`)}
                >
                  <div className="p-4">
                    {/* Header */}
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-2">
                        <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${
                          isOnline ? 'bg-emerald-500' : 'bg-gray-400'
                        }`}>
                          <Server className="w-4 h-4 text-white" />
                        </div>
                        <div>
                          <h3 className="font-semibold text-gray-900 text-sm">{host.hostname || host.ip}</h3>
                          <p className="text-xs text-gray-500">{host.ip}</p>
                        </div>
                      </div>
                      <div className={`w-2 h-2 rounded-full ${
                        isOnline ? 'bg-emerald-500' : 'bg-gray-400'
                      }`} />
                    </div>

                    {/* Group Badge */}
                    {host.group && (
                      <div className="mb-3">
                        <span className="inline-block px-2 py-0.5 bg-blue-100 text-blue-700 text-xs font-medium rounded">
                          {host.group}
                        </span>
                      </div>
                    )}

                    {/* Metrics - Circular Progress */}
                    <div className="flex items-center justify-around mb-3">
                      {/* CPU */}
                      <div className="text-center">
                        <div className="relative w-12 h-12 mx-auto">
                          <svg className="transform -rotate-90 w-12 h-12">
                            <circle cx="24" cy="24" r="20" stroke="#e5e7eb" strokeWidth="4" fill="none" />
                            <circle 
                              cx="24" cy="24" r="20" 
                              stroke="#3b82f6" 
                              strokeWidth="4" 
                              fill="none"
                              strokeDasharray={`${2 * Math.PI * 20}`}
                              strokeDashoffset={`${2 * Math.PI * 20 * (1 - cpuUsage / 100)}`}
                              strokeLinecap="round"
                              className="transition-all duration-500"
                            />
                          </svg>
                          <div className="absolute inset-0 flex items-center justify-center">
                            <span className="text-xs font-bold text-gray-900">{cpuUsage.toFixed(0)}%</span>
                          </div>
                        </div>
                        <p className="text-xs text-gray-600 mt-1">CPU</p>
                      </div>

                      {/* RAM */}
                      <div className="text-center">
                        <div className="relative w-12 h-12 mx-auto">
                          <svg className="transform -rotate-90 w-12 h-12">
                            <circle cx="24" cy="24" r="20" stroke="#e5e7eb" strokeWidth="4" fill="none" />
                            <circle 
                              cx="24" cy="24" r="20" 
                              stroke="#10b981" 
                              strokeWidth="4" 
                              fill="none"
                              strokeDasharray={`${2 * Math.PI * 20}`}
                              strokeDashoffset={`${2 * Math.PI * 20 * (1 - memoryUsage / 100)}`}
                              strokeLinecap="round"
                              className="transition-all duration-500"
                            />
                          </svg>
                          <div className="absolute inset-0 flex items-center justify-center">
                            <span className="text-xs font-bold text-gray-900">{memoryUsage.toFixed(0)}%</span>
                          </div>
                        </div>
                        <p className="text-xs text-gray-600 mt-1">RAM</p>
                      </div>

                      {/* Disk */}
                      <div className="text-center">
                        <div className="relative w-12 h-12 mx-auto">
                          <svg className="transform -rotate-90 w-12 h-12">
                            <circle cx="24" cy="24" r="20" stroke="#e5e7eb" strokeWidth="4" fill="none" />
                            <circle 
                              cx="24" cy="24" r="20" 
                              stroke="#8b5cf6" 
                              strokeWidth="4" 
                              fill="none"
                              strokeDasharray={`${2 * Math.PI * 20}`}
                              strokeDashoffset={`${2 * Math.PI * 20 * (1 - diskUsage / 100)}`}
                              strokeLinecap="round"
                              className="transition-all duration-500"
                            />
                          </svg>
                          <div className="absolute inset-0 flex items-center justify-center">
                            <span className="text-xs font-bold text-gray-900">{diskUsage.toFixed(0)}%</span>
                          </div>
                        </div>
                        <p className="text-xs text-gray-600 mt-1">Disk</p>
                      </div>
                    </div>

                    {/* Status Footer */}
                    <div className="flex items-center justify-between pt-3 border-t border-gray-100">
                      <span className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium ${
                        isOnline ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-700'
                      }`}>
                        <div className={`w-1.5 h-1.5 rounded-full ${
                          isOnline ? 'bg-emerald-500' : 'bg-gray-500'
                        }`} />
                        {isOnline ? 'online' : 'offline'}
                      </span>
                      {metrics?.uptime && (
                        <span className="text-xs text-gray-500">
                          {formatTimestamp(metrics.uptime)}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
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