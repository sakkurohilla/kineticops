import React, { useState, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import { Plus, Bell, Trash2, RefreshCw } from 'lucide-react';
import Button from '../../components/common/Button';
import api from '../../services/api/client';

interface AlertRule {
  id: number;
  metric_name: string;
  operator: string;
  threshold: number;
  window: number;
  frequency: number;
  notification_webhook: string;
  escalation_policy: string;
  created_at: string;
}

const AlertRules: React.FC = () => {
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [formData, setFormData] = useState({
    metric_name: 'cpu_usage',
    operator: '>',
    threshold: 90,
    window: 5,
    frequency: 3,
    notification_webhook: '',
    escalation_policy: '',
  });

  useEffect(() => {
    fetchRules();
  }, []);

  const fetchRules = async () => {
    try {
      setLoading(true);
      const response = await api.get('/alerts/rules');
      setRules(response.data || []);
    } catch (err) {
      console.error('Failed to fetch alert rules:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.post('/alerts/rules', formData);
      setShowCreateForm(false);
      setFormData({
        metric_name: 'cpu_usage',
        operator: '>',
        threshold: 90,
        window: 5,
        frequency: 3,
        notification_webhook: '',
        escalation_policy: '',
      });
      fetchRules();
    } catch (err: any) {
      alert('Failed to create alert rule: ' + (err.message || 'Unknown error'));
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this alert rule?')) {
      return;
    }
    try {
      await api.delete(`/alerts/rules/${id}`);
      fetchRules();
    } catch (err: any) {
      alert('Failed to delete alert rule: ' + (err.message || 'Unknown error'));
    }
  };

  const getMetricLabel = (metricName: string) => {
    const labels: Record<string, string> = {
      cpu_usage: 'CPU Usage',
      memory_usage: 'Memory Usage',
      disk_usage: 'Disk Usage',
      network_in: 'Network In',
      network_out: 'Network Out',
      load_average: 'Load Average',
    };
    return labels[metricName] || metricName;
  };

  const getOperatorLabel = (operator: string) => {
    const labels: Record<string, string> = {
      '>': 'Greater than',
      '<': 'Less than',
      '>=': 'Greater than or equal',
      '<=': 'Less than or equal',
      '==': 'Equal to',
    };
    return labels[operator] || operator;
  };

  return (
    <MainLayout>
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Alert Rules</h1>
            <p className="text-gray-600">Configure monitoring alerts for your infrastructure</p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={fetchRules}>
              <RefreshCw className="w-4 h-4" />
              Refresh
            </Button>
            <Button variant="primary" onClick={() => setShowCreateForm(true)}>
              <Plus className="w-4 h-4" />
              Create Rule
            </Button>
          </div>
        </div>

        {/* Rules List */}
        {loading ? (
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">Loading alert rules...</p>
          </div>
        ) : rules.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
            <Bell className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">No Alert Rules</h3>
            <p className="text-gray-600 mb-4">Get started by creating your first alert rule</p>
            <Button variant="primary" onClick={() => setShowCreateForm(true)}>
              <Plus className="w-4 h-4" />
              Create Alert Rule
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {rules.map((rule) => (
              <div key={rule.id} className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-lg transition-shadow">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center space-x-2">
                    <div className="w-10 h-10 bg-gradient-to-br from-orange-500 to-red-600 rounded-lg flex items-center justify-center">
                      <Bell className="w-5 h-5 text-white" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-gray-900">{getMetricLabel(rule.metric_name)}</h3>
                      <p className="text-xs text-gray-500">Rule #{rule.id}</p>
                    </div>
                  </div>
                  <div className="flex gap-1">
                    <button
                      onClick={() => handleDelete(rule.id)}
                      className="p-1 text-gray-400 hover:text-red-600 transition-colors"
                      title="Delete rule"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-600">Condition:</span>
                    <span className="font-medium text-gray-900">
                      {getOperatorLabel(rule.operator)} {rule.threshold}%
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-600">Window:</span>
                    <span className="font-medium text-gray-900">{rule.window} minutes</span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-600">Frequency:</span>
                    <span className="font-medium text-gray-900">{rule.frequency} breaches</span>
                  </div>
                  {rule.notification_webhook && (
                    <div className="pt-2 border-t border-gray-100">
                      <span className="text-xs text-gray-500">Webhook configured</span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Create Form Modal */}
        {showCreateForm && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
              <div className="p-6">
                <h2 className="text-xl font-bold text-gray-900 mb-4">Create Alert Rule</h2>
                
                <form onSubmit={handleCreate} className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Metric
                    </label>
                    <select
                      value={formData.metric_name}
                      onChange={(e) => setFormData({ ...formData, metric_name: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      required
                    >
                      <option value="cpu_usage">CPU Usage</option>
                      <option value="memory_usage">Memory Usage</option>
                      <option value="disk_usage">Disk Usage</option>
                      <option value="network_in">Network In</option>
                      <option value="network_out">Network Out</option>
                      <option value="load_average">Load Average</option>
                    </select>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Operator
                      </label>
                      <select
                        value={formData.operator}
                        onChange={(e) => setFormData({ ...formData, operator: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        required
                      >
                        <option value=">">Greater than (&gt;)</option>
                        <option value="<">Less than (&lt;)</option>
                        <option value=">=">Greater than or equal (&gt;=)</option>
                        <option value="<=">Less than or equal (&lt;=)</option>
                        <option value="==">Equal to (==)</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Threshold (%)
                      </label>
                      <input
                        type="number"
                        value={formData.threshold}
                        onChange={(e) => setFormData({ ...formData, threshold: parseFloat(e.target.value) })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        min="0"
                        max="100"
                        step="0.1"
                        required
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Time Window (minutes)
                      </label>
                      <input
                        type="number"
                        value={formData.window}
                        onChange={(e) => setFormData({ ...formData, window: parseInt(e.target.value) })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        min="1"
                        required
                      />
                      <p className="text-xs text-gray-500 mt-1">Time period to evaluate the condition</p>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Breach Frequency
                      </label>
                      <input
                        type="number"
                        value={formData.frequency}
                        onChange={(e) => setFormData({ ...formData, frequency: parseInt(e.target.value) })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        min="1"
                        required
                      />
                      <p className="text-xs text-gray-500 mt-1">Number of breaches to trigger alert</p>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Notification Webhook (Optional)
                    </label>
                    <input
                      type="url"
                      value={formData.notification_webhook}
                      onChange={(e) => setFormData({ ...formData, notification_webhook: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="https://hooks.slack.com/services/..."
                    />
                    <p className="text-xs text-gray-500 mt-1">Webhook URL for notifications (Slack, Discord, etc.)</p>
                  </div>

                  <div className="flex gap-3 pt-4">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => setShowCreateForm(false)}
                      className="flex-1"
                    >
                      Cancel
                    </Button>
                    <Button type="submit" variant="primary" className="flex-1">
                      Create Rule
                    </Button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        )}
      </div>
    </MainLayout>
  );
};

export default AlertRules;
