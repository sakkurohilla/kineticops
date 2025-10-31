import apiClient from './client';

export interface Alert {
  id: number;
  host_id: number;
  host_name?: string;
  message: string;
  severity: string;
  status: string;
  triggered_at: string;
  created_at: string;
}

export interface AlertRule {
  id?: number;
  metric_name: string;
  operator: string;
  threshold: number;
  window: number;
  frequency: number;
  notification_webhook?: string;
  escalation_policy?: string;
  created_at?: string;
}

class AlertsService {
  async getAlerts(limit?: number): Promise<Alert[]> {
    const params = limit ? `?limit=${limit}` : '';
    return await apiClient.get(`/alerts${params}`);
  }

  async getRules(): Promise<AlertRule[]> {
    return await apiClient.get('/alerts/rules');
  }

  async createRule(rule: Omit<AlertRule, 'id' | 'created_at'>): Promise<AlertRule> {
    return await apiClient.post('/alerts/rules', rule);
  }

  async updateAlert(id: number, data: { status?: string }): Promise<void> {
    return await apiClient.patch(`/alerts/${id}`, data);
  }

  async silenceAlert(id: number): Promise<void> {
    return this.updateAlert(id, { status: 'silenced' });
  }
}

export default new AlertsService();