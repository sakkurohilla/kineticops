import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import MainLayout from '../../components/layout/MainLayout';
import { 
  ArrowLeft, 
  Server, 
  Edit, 
  Trash2, 
  RefreshCw,
  AlertCircle
} from 'lucide-react';
import Button from '../../components/common/Button';
import Badge from '../../components/common/Badge';
import Card from '../../components/common/Card';
import HostDashboard from '../../components/hosts/HostDashboard';
import hostService from '../../services/api/hostService';
import { Host } from '../../types';
import AddHostForm from './AddHostForm';

const HostDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [host, setHost] = useState<Host | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [showEditModal, setShowEditModal] = useState(false);

  useEffect(() => {
    if (id) {
      fetchHostDetails();
    }
  }, [id]);

  const fetchHostDetails = async () => {
    try {
      setLoading(true);
      setError('');
      const data = await hostService.getHost(parseInt(id!));
      setHost(data);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch host details');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm('Are you sure you want to delete this host? This action cannot be undone.')) {
      return;
    }

    try {
      await hostService.deleteHost(parseInt(id!));
      navigate('/hosts');
    } catch (err: any) {
      alert('Failed to delete host: ' + err.message);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'online':
        return 'success';
      case 'offline':
        return 'error';
      case 'warning':
        return 'warning';
      default:
        return 'info';
    }
  };

  if (loading) {
    return (
      <MainLayout>
        <div className="p-6 lg:p-8 flex items-center justify-center min-h-screen">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading host details...</p>
          </div>
        </div>
      </MainLayout>
    );
  }

  if (error || !host) {
    return (
      <MainLayout>
        <div className="p-6 lg:p-8">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6">
            <div className="flex items-center gap-3 mb-4">
              <AlertCircle className="w-6 h-6 text-red-600" />
              <h3 className="font-semibold text-red-900">Error Loading Host</h3>
            </div>
            <p className="text-sm text-red-700 mb-4">{error || 'Host not found'}</p>
            <Button variant="outline" onClick={() => navigate('/hosts')}>
              <ArrowLeft className="w-4 h-4" />
              Back to Hosts
            </Button>
          </div>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="mb-8">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate('/hosts')}
            className="mb-4"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to Hosts
          </Button>

          <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
            <div className="flex items-start gap-4">
              <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg">
                <Server className="w-8 h-8 text-white" />
              </div>
              <div>
                <h1 className="text-3xl font-bold text-gray-900 mb-2">
                  {host.hostname || 'Unnamed Host'}
                </h1>
                <div className="flex items-center gap-3">
                  <Badge variant={getStatusColor(host.agent_status || 'offline')}>
                    {host.agent_status || 'offline'}
                  </Badge>
                  <span className="text-gray-600">{host.ip}</span>
                </div>
              </div>
            </div>

            <div className="flex gap-3">
              <Button variant="outline" onClick={fetchHostDetails}>
                <RefreshCw className="w-4 h-4" />
                Refresh
              </Button>
                <Button variant="outline" onClick={() => setShowEditModal(true)}>
                  <Edit className="w-4 h-4" />
                  Edit
                </Button>
              <Button
                variant="outline"
                onClick={handleDelete}
                className="text-red-600 hover:bg-red-50 border-red-200"
              >
                <Trash2 className="w-4 h-4" />
                Delete
              </Button>
            </div>
          </div>
        </div>

        {/* Host Information Card */}
        <Card className="mb-8">
          <h2 className="text-xl font-bold text-gray-900 mb-4">Host Information</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Hostname</p>
              <p className="text-base text-gray-900">{host.hostname || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">IP Address</p>
              <p className="text-base text-gray-900">{host.ip}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Operating System</p>
              <p className="text-base text-gray-900">{host.os || 'Linux'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Group</p>
              <p className="text-base text-gray-900">{host.group || 'default'}</p>
            </div>

            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Last Seen</p>
              <p className="text-base text-gray-900">
                {host.last_seen ? new Date(host.last_seen).toLocaleString() : 'Never'}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Created At</p>
              <p className="text-base text-gray-900">
                {host.created_at ? new Date(host.created_at).toLocaleString() : 'N/A'}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Tenant ID</p>
              <p className="text-base text-gray-900">{host.tenant_id || 'N/A'}</p>
            </div>
          </div>

          {/* Tags */}
          {host.tags && (
            <div className="mt-6 pt-6 border-t border-gray-200">
              <p className="text-sm font-medium text-gray-600 mb-2">Tags</p>
              <div className="flex flex-wrap gap-2">
                {host.tags.split(',').map((tag, idx) => (
                  <span
                    key={idx}
                    className="px-3 py-1 bg-blue-50 text-blue-700 text-sm font-medium rounded-lg"
                  >
                    {tag.trim()}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Description */}
          {host.description && (
            <div className="mt-6 pt-6 border-t border-gray-200">
              <p className="text-sm font-medium text-gray-600 mb-2">Description</p>
              <p className="text-base text-gray-900">{host.description}</p>
            </div>
          )}
        </Card>

        {/* Metrics Dashboard */}
        <HostDashboard hostId={parseInt(id!)} />
        {showEditModal && host && (
          <AddHostForm
            mode="edit"
            hostId={host.id}
            initialData={{
              hostname: host.hostname || '',
              ip: host.ip || '',
              ssh_user: host.ssh_user || 'root',
              ssh_port: host.ssh_port || 22,
              os: host.os || 'linux',
              group: host.group || 'default',
              tags: host.tags || '',
              description: host.description || '',
              ssh_password: '',
            }}
            onClose={() => setShowEditModal(false)}
            onSuccess={() => {
              setShowEditModal(false);
              fetchHostDetails();
            }}
          />
        )}
      </div>
    </MainLayout>
  );
};

export default HostDetails;