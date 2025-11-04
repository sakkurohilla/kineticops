import React, { useState } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import { Plus, Server, Search, RefreshCw } from 'lucide-react';
import Button from '../../components/common/Button';
import SimpleAddHostForm from '../../components/hosts/SimpleAddHostForm';
import HostCard from '../../components/hosts/HostCard';
import { useHosts } from '../../hooks/useHosts';
import hostService from '../../services/api/hostService';

const Hosts: React.FC = () => {
  const { hosts, loading, error, refetch } = useHosts();
  const [showAddForm, setShowAddForm] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [refreshing, setRefreshing] = useState(false);

  const handleRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setTimeout(() => setRefreshing(false), 500);
  };

  const handleDelete = async (id: number) => {
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

  const filteredHosts = hosts.filter(
    (host) =>
      host.hostname?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      host.ip?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      host.group?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <MainLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 mb-2">Hosts</h1>
            <p className="text-gray-600">
              Manage and monitor your infrastructure hosts
            </p>
          </div>

          <div className="flex gap-3">
            <Button
              variant="outline"
              onClick={handleRefresh}
              disabled={refreshing}
            >
              <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            <Button variant="primary" onClick={() => setShowAddForm(true)}>
              <Plus className="w-4 h-4" />
              Add Host
            </Button>
          </div>
        </div>

        {/* Search Bar */}
        <div className="mb-6">
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search hosts by name, IP, or group..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600 mb-1">Total Hosts</p>
                <p className="text-3xl font-bold text-gray-900">{hosts.length}</p>
              </div>
              <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
                <Server className="w-6 h-6 text-blue-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600 mb-1">Online</p>
                <p className="text-3xl font-bold text-green-600">
                  {hosts.filter((h) => h.agent_status === 'online').length}
                </p>
              </div>
              <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
                <div className="w-3 h-3 bg-green-600 rounded-full animate-pulse"></div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600 mb-1">Offline</p>
                <p className="text-3xl font-bold text-red-600">
                  {hosts.filter((h) => h.agent_status === 'offline' || !h.agent_status).length}
                </p>
              </div>
              <div className="w-12 h-12 bg-red-100 rounded-lg flex items-center justify-center">
                <div className="w-3 h-3 bg-red-600 rounded-full"></div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600 mb-1">Groups</p>
                <p className="text-3xl font-bold text-purple-600">
                  {new Set(hosts.map((h) => h.group || 'default')).size}
                </p>
              </div>
              <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center">
                <Server className="w-6 h-6 text-purple-600" />
              </div>
            </div>
          </div>
        </div>

        {/* Loading State */}
        {loading && (
          <div className="flex items-center justify-center py-16">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading hosts...</p>
            </div>
          </div>
        )}

        {/* Error State */}
        {error && !loading && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
            <p className="text-red-800">{error}</p>
            <Button variant="outline" size="sm" onClick={refetch} className="mt-4">
              Try Again
            </Button>
          </div>
        )}

        {/* Empty State */}
        {!loading && !error && hosts.length === 0 && (
          <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-12 text-center">
            <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <Server className="w-10 h-10 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold text-gray-900 mb-2">No Hosts Yet</h3>
            <p className="text-gray-600 mb-6 max-w-md mx-auto">
              Get started by adding your first host. You'll be able to monitor metrics, logs, and performance in real-time.
            </p>
            <Button variant="primary" size="lg" onClick={() => setShowAddForm(true)}>
              <Plus className="w-5 h-5" />
              Add Host
            </Button>
          </div>
        )}

        {/* Hosts Grid */}
        {!loading && !error && filteredHosts.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredHosts.map((host) => (
              <HostCard key={host.id} host={host} onDelete={handleDelete} />
            ))}
          </div>
        )}

        {/* No Search Results */}
        {!loading && !error && hosts.length > 0 && filteredHosts.length === 0 && (
          <div className="bg-white rounded-lg border border-gray-200 p-8 text-center">
            <p className="text-gray-600">No hosts found matching "{searchQuery}"</p>
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