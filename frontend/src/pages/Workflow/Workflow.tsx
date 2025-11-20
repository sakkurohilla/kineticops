import React, { useState, useEffect } from 'react';
import { Search, Bell, Settings, Shield, ChevronRight, Eye, Download, Zap } from 'lucide-react';
import MainLayout from '../../components/layout/MainLayout';
import CredentialsModal from '../../components/workflow/CredentialsModal';
import ServiceBubble from '../../components/workflow/ServiceBubble';
import ServiceControlPanel from '../../components/workflow/ServiceControlPanel';
import { Host } from '../../types';
import hostService from '../../services/api/hostService';
import { useWorkflowSession, useServiceControl } from '../../hooks/useWorkflow';
import { handleApiError } from '../../utils/errorHandler';
import useWebSocket from '../../hooks/useWebsocket';

const Workflow: React.FC = () => {
  const [hosts, setHosts] = useState<Host[]>([]);
  const [selectedHost, setSelectedHost] = useState<Host | null>(null);
  const [hoveredHost, setHoveredHost] = useState<Host | null>(null);
  const [showCredentialsModal, setShowCredentialsModal] = useState(false);
  const [selectedService, setSelectedService] = useState<any>(null);
  const [showServiceControl, setShowServiceControl] = useState(false);
  const [activeTab, setActiveTab] = useState('overview');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedFilter, setSelectedFilter] = useState('all');
  const [autoHealing, setAutoHealing] = useState(true);
  const [loading, setLoading] = useState(true);


  const { session, createSession, loading: sessionLoading } = useWorkflowSession();

  const { controlService, loading: controlLoading } = useServiceControl(session?.session_token);

  const [services, setServices] = useState<any[]>([]);
  const [applications] = useState<any[]>([]);
  const [performanceMetrics] = useState<any[]>([]);
  const [logs] = useState<any[]>([]);
  const [selectedLogService, setSelectedLogService] = useState<string>('all');

  useEffect(() => {
    fetchHosts();
  }, []);

  // WebSocket handler for real-time service updates
  const handleWebSocketMessage = (data: any) => {
    if (!data || !selectedHost) return;
    
    // Handle services updates for the selected host
    if (data.type === 'services' && data.host_id === selectedHost.id) {
      const servicesData = data.services;
      if (Array.isArray(servicesData)) {
        console.log('[Workflow] Received services update via WebSocket:', servicesData.length, 'services');
        setServices(servicesData);
      }
    }
  };

  useWebSocket(handleWebSocketMessage);

  const fetchHosts = async () => {
    try {
      setLoading(true);
      const hostsData = await hostService.getAllHosts();
      setHosts(hostsData);
    } catch (err: any) {
      const apiError = handleApiError(err);
      console.error('Failed to fetch hosts:', apiError.message);
    } finally {
      setLoading(false);
    }
  };

  const handleHostClick = (host: Host) => {
    setSelectedHost(host);
    setShowCredentialsModal(true);
  };

  const handleCredentialsSubmit = async (credentials: { username: string; password?: string; ssh_key?: string; pem_file?: string }) => {
    try {
      const newSession = await createSession({
        host_id: selectedHost!.id,
        username: credentials.username,
        password: credentials.password,
        ssh_key: credentials.ssh_key || credentials.pem_file // Use pem_file as ssh_key
      });
      setShowCredentialsModal(false);
      
      // Fetch services using the new session token directly
      if (newSession?.session_token && selectedHost) {
        try {
          const response = await fetch(`/api/v1/workflow/${selectedHost.id}/discover`, {
            method: 'POST',
            headers: {
              'X-Session-Token': newSession.session_token,
              'Content-Type': 'application/json'
            }
          });
          
          if (response.ok) {
            const data = await response.json();
            // Ensure services is always an array
            const servicesData = data.services;
            if (Array.isArray(servicesData)) {
              setServices(servicesData);
            } else if (servicesData && typeof servicesData === 'object') {
              // If services is an object with services array inside
              setServices(Array.isArray(servicesData.services) ? servicesData.services : []);
            } else {
              setServices([]);
            }
          } else {
            console.error('Failed to fetch services:', response.statusText);
            setServices([]);
          }
        } catch (err) {
          console.error('Failed to fetch services:', err);
          setServices([]);
        }
      }
    } catch (err: any) {
      console.error('Session creation failed:', err);
      throw err; // Re-throw to show error in modal
    }
  };

  const handleServiceClick = (service: any) => {
    setSelectedService(service);
    setShowServiceControl(true);
  };

  const handleServiceAction = async (action: 'start' | 'stop' | 'restart' | 'enable' | 'disable') => {
    try {
      const result = await controlService(selectedService.id, action);
      return result;
    } catch (err) {
      throw err;
    }
  };

  const getStatusIcon = (status: string) => {
    const icons = { 
      'online': '🟢', 
      'offline': '🔴', 
      'warning': '⚠️',
      'running': '🟢', 
      'stopped': '🔴', 
      'healthy': '✅'
    };
    return icons[status as keyof typeof icons] || '❓';
  };

  const filteredHosts = hosts.filter(host => selectedFilter === 'all' || host.agent_status === selectedFilter);
  const filteredSearchHosts = filteredHosts.filter(host =>
    host.hostname?.toLowerCase().includes(searchQuery.toLowerCase()) || host.ip.includes(searchQuery)
  );

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
      <div className="min-h-screen bg-gray-50">
        {/* Header */}
        <div className="bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 border-b border-slate-700 px-6 py-4">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-3xl font-bold text-blue-400 flex items-center gap-2">
                <Zap className="w-8 h-8" />
                Workflow Control
              </h2>
              <p className="text-gray-400 text-sm mt-1">Enterprise host & service management</p>
            </div>
            <div className="flex items-center gap-3">
              <div className="relative">
                <Search size={16} className="absolute left-3 top-2.5 text-gray-500" />
                <input
                  type="text"
                  placeholder="Search..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9 pr-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-sm w-48 focus:ring-2 focus:ring-blue-500 text-white"
                />
              </div>
              <button className="relative p-2 hover:bg-slate-700 rounded-lg text-white">
                <Bell size={20} />
                <span className="absolute top-1 right-1 bg-red-500 text-white text-xs w-5 h-5 flex items-center justify-center rounded-full font-bold">3</span>
              </button>
              <button className="p-2 hover:bg-slate-700 rounded-lg text-white">
                <Settings size={20} />
              </button>
              <label className="flex items-center gap-2 px-3 py-1 bg-slate-700/50 rounded-lg cursor-pointer hover:bg-slate-700 text-white">
                <Shield size={16} className={autoHealing ? 'text-green-400' : 'text-gray-400'} />
                <span className="text-xs font-bold">{autoHealing ? 'Heal ON' : 'Heal OFF'}</span>
                <input type="checkbox" checked={autoHealing} onChange={(e) => setAutoHealing(e.target.checked)} className="hidden" />
              </label>
            </div>
          </div>

          {/* Filters */}
          <div className="flex gap-2">
            {['all', 'online', 'offline'].map((f) => (
              <button
                key={f}
                onClick={() => setSelectedFilter(f)}
                className={`px-4 py-1.5 rounded-lg text-sm font-semibold ${
                  selectedFilter === f ? 'bg-blue-600 text-white' : 'bg-slate-700/50 hover:bg-slate-700 text-gray-300'
                }`}
              >
                {f === 'all' ? '📋 All' : f === 'online' ? '✅ Online' : '⚠️ Offline'}
              </button>
            ))}
          </div>
        </div>

        {/* Content */}
        <div className="p-6">
          {!selectedHost || !session ? (
            // Host List
            <div>
              <h3 className="text-2xl font-bold mb-6 text-gray-900">Monitored Hosts ({filteredSearchHosts.length})</h3>

              {filteredSearchHosts.length === 0 ? (
                <div className="text-center py-12">
                  <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                    <span className="text-2xl">🖥️</span>
                  </div>
                  <h3 className="text-xl font-bold text-gray-900 mb-2">No Hosts Available</h3>
                  <p className="text-gray-600 mb-6">Add hosts to your infrastructure to start managing workflows.</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                  {filteredSearchHosts.map((host) => (
                    <div key={host.id} onMouseEnter={() => setHoveredHost(host)} onMouseLeave={() => setHoveredHost(null)} className="relative">
                      <div
                        onClick={() => handleHostClick(host)}
                        className={`w-40 h-40 rounded-full mx-auto flex items-center justify-center cursor-pointer transition transform hover:scale-110 hover:-translate-y-3 shadow-xl ${
                          host.agent_status === 'online' ? 'bg-gradient-to-br from-emerald-400 to-emerald-600' : 'bg-gradient-to-br from-yellow-400 to-yellow-600'
                        }`}
                      >
                        <div className="text-center text-white">
                          <div className="text-5xl mb-1">🖥️</div>
                          <p className="font-bold text-sm">{host.hostname}</p>
                          <p className="text-xs opacity-90">{host.ip}</p>
                        </div>
                        <div className={`absolute inset-0 rounded-full border-4 ${
                          host.agent_status === 'online' ? 'border-green-300 animate-pulse' : 'border-yellow-300'
                        }`}></div>
                        <div className="absolute bottom-2 right-2 bg-black/70 px-2 py-1 rounded text-xs font-bold text-green-400">
                          {host.agent_status === 'online' ? '100%' : '0%'}
                        </div>
                      </div>

                      {hoveredHost?.id === host.id && (
                        <div className="absolute top-full mt-4 left-1/2 -translate-x-1/2 z-50 w-72 bg-white rounded-lg shadow-xl border border-gray-200 p-4">
                          <div className="flex items-center justify-between mb-3 pb-3 border-b border-gray-200">
                            <h4 className="font-bold text-gray-900">{host.hostname}</h4>
                            <span className="text-xl">{getStatusIcon(host.agent_status || 'offline')}</span>
                          </div>
                          <div className="grid grid-cols-2 gap-2 text-sm mb-3">
                            <div className="bg-gray-100 p-2 rounded">
                              <p className="text-xs text-gray-600">IP</p>
                              <p className="text-blue-600 font-mono text-xs">{host.ip}</p>
                            </div>
                            <div className="bg-gray-100 p-2 rounded">
                              <p className="text-xs text-gray-600">OS</p>
                              <p className="font-bold text-xs">{host.os || 'Linux'}</p>
                            </div>
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : (
            // Host Management
            <div>
              <button
                onClick={() => { setSelectedHost(null); }}
                className="mb-4 text-blue-600 hover:text-blue-800 font-bold flex items-center gap-1"
              >
                <ChevronRight size={18} className="rotate-180" /> Back to Hosts
              </button>

              <div className="bg-blue-50 rounded-lg p-6 border border-blue-200 mb-6">
                <div className="flex justify-between items-start mb-3">
                  <div>
                    <h2 className="text-3xl font-bold mb-2 text-gray-900">{selectedHost.hostname}</h2>
                    <p className="text-gray-600 text-sm">{selectedHost.ip} • {selectedHost.os || 'Linux'} • Session Active</p>
                  </div>
                  <div className="flex gap-2">
                    <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-semibold flex items-center gap-2 text-white">
                      <Eye size={16} /> Health Check
                    </button>
                    <button className="px-4 py-2 bg-gray-600 hover:bg-gray-700 rounded text-sm font-semibold flex items-center gap-2 text-white">
                      <Download size={16} /> Export
                    </button>
                  </div>
                </div>
              </div>

              {/* Performance Metrics */}
              {performanceMetrics.length > 0 && (
                <div className="grid grid-cols-4 gap-4 mb-6">
                  {performanceMetrics.map((m, i) => (
                    <div key={i} className="bg-white rounded-lg p-4 border border-gray-200 shadow-sm">
                      <p className="text-xs text-gray-600 mb-2">{m.name}</p>
                      <p className="text-xl font-bold text-gray-900">{m.value}</p>
                    </div>
                  ))}
                </div>
              )}

              {/* Tabs */}
              <div className="flex gap-2 mb-6 border-b border-gray-200 pb-4">
                {['overview', 'services', 'logs', 'analytics'].map((t) => (
                  <button
                    key={t}
                    onClick={() => setActiveTab(t)}
                    className={`px-4 py-2 rounded font-semibold transition ${
                      activeTab === t ? 'bg-blue-600 text-white' : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    {t === 'overview' && '📊'} {t === 'services' && '⚙️'} {t === 'logs' && '📝'} {t === 'analytics' && '📈'}
                    {' '}{t.charAt(0).toUpperCase() + t.slice(1)}
                  </button>
                ))}
              </div>

              {activeTab === 'overview' && (
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  <div className="bg-white rounded-lg p-6 border border-gray-200 shadow-sm">
                    <h4 className="text-lg font-bold mb-4 text-gray-900">⚙️ Services</h4>
                    {!Array.isArray(services) || services.length === 0 ? (
                      <div className="text-center py-8 text-gray-500">
                        <p>No services discovered</p>
                        <p className="text-sm mt-1">Services will appear here once agent is connected</p>
                      </div>
                    ) : (
                      <div className="grid grid-cols-2 gap-3">
                        {services.map((s) => (
                          <ServiceBubble
                            key={s.id}
                            service={s}
                            onClick={() => handleServiceClick(s)}
                            size="sm"
                          />
                        ))}
                      </div>
                    )}
                  </div>

                  <div className="bg-white rounded-lg p-6 border border-gray-200 shadow-sm">
                    <h4 className="text-lg font-bold mb-4 text-gray-900">📦 Applications</h4>
                    {applications.length === 0 ? (
                      <div className="text-center py-8 text-gray-500">
                        <p>No applications detected</p>
                        <p className="text-sm mt-1">Applications will appear here once discovered</p>
                      </div>
                    ) : (
                      <div className="grid grid-cols-2 gap-3">
                        {applications.map((a) => (
                          <ServiceBubble
                            key={a.id}
                            service={a}
                            onClick={() => handleServiceClick(a)}
                            size="sm"
                          />
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )}

              {activeTab === 'logs' && (
                <div>
                  <div className="mb-4 flex items-center gap-4">
                    <label className="text-sm font-medium text-gray-700">Service:</label>
                    <select
                      value={selectedLogService}
                      onChange={(e) => setSelectedLogService(e.target.value)}
                      className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="all">All Services</option>
                      {services.map((s) => (
                        <option key={s.id} value={s.name}>{s.name}</option>
                      ))}
                    </select>
                    <button className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
                      🔄 Refresh
                    </button>
                    <button className="px-4 py-2 bg-green-600 text-white rounded-lg text-sm hover:bg-green-700">
                      📡 Live Tail
                    </button>
                  </div>
                  <div className="bg-gray-900 rounded-lg p-4 border border-gray-700 font-mono text-xs max-h-96 overflow-y-auto">
                    {logs.length === 0 ? (
                      <div className="text-center py-8 text-gray-500">
                        <p>No logs available</p>
                        <p className="text-xs mt-1">Logs will appear here once services are running</p>
                      </div>
                    ) : (
                      logs.map((log, i) => (
                        <div key={i} className={`mb-2 ${
                          log.level === 'ERROR' ? 'text-red-400' :
                          log.level === 'WARNING' ? 'text-yellow-400' :
                          log.level === 'SUCCESS' ? 'text-green-400' :
                          'text-gray-400'
                        }`}>
                          <span className="text-gray-600">[{log.time}]</span>
                          {' '}<span className="font-bold">[{log.level}]</span>
                          {' '}<span className="text-purple-400">{log.service}:</span>
                          {' '}<span>{log.message}</span>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              )}

              {activeTab === 'services' && (
                <div className="bg-white rounded-lg p-6 border border-gray-200 shadow-sm">
                  <h4 className="text-lg font-bold mb-4 text-gray-900">Service Details</h4>
                  {services.length === 0 ? (
                    <div className="text-center py-12">
                      <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                        <span className="text-2xl">⚙️</span>
                      </div>
                      <h3 className="text-lg font-bold text-gray-900 mb-2">No Services Found</h3>
                      <p className="text-gray-600 mb-4">Services will appear here once the agent discovers them on the host.</p>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {services.map((service) => (
                        <div key={service.id} className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                          <div className="flex items-center justify-between mb-2">
                            <h5 className="font-bold text-gray-900">{service.name}</h5>
                            <span className={`px-2 py-1 rounded text-xs font-bold ${
                              service.status === 'running' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                            }`}>
                              {service.status}
                            </span>
                          </div>
                          <div className="text-sm text-gray-600 space-y-1">
                            <p>Port: {service.port || 'N/A'}</p>
                            <p>PID: {service.process_id || 'N/A'}</p>
                            <p>CPU: {service.cpu_usage || 0}%</p>
                            <p>Memory: {service.memory_usage || 0}MB</p>
                          </div>
                          <button
                            onClick={() => handleServiceClick(service)}
                            className="mt-3 w-full px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                          >
                            Manage
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {activeTab === 'analytics' && (
                <div className="bg-white rounded-lg p-6 text-center">
                  <p className="text-gray-500">Analytics features coming soon</p>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Modals */}
        <CredentialsModal
          isOpen={showCredentialsModal}
          onClose={() => setShowCredentialsModal(false)}
          onSubmit={handleCredentialsSubmit}
          hostName={selectedHost?.hostname || ''}
          hostIP={selectedHost?.ip || ''}
          loading={sessionLoading}
        />

        {showServiceControl && selectedService && (
          <ServiceControlPanel
            service={selectedService}
            onAction={handleServiceAction}
            onClose={() => setShowServiceControl(false)}
            loading={controlLoading}
          />
        )}
      </div>
    </MainLayout>
  );
};

export default Workflow;