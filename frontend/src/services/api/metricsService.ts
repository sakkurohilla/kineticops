import apiClient from './client';

export interface MetricData {
  id: number;
  host_id: number;
  name: string;
  value: number;
  timestamp: string;
}

class MetricsService {
  async getMetricsRange(timeRange: string, hostId?: number): Promise<MetricData[]> {
    const params = new URLSearchParams();
    params.append('range', timeRange);
    if (hostId) params.append('host_id', hostId.toString());
    
    return await apiClient.get(`/metrics/range?${params.toString()}`);
  }

  async getMetricsCustomRange(startTime: string, endTime: string, hostId?: number): Promise<MetricData[]> {
    const params = new URLSearchParams();
    params.append('start', startTime);
    params.append('end', endTime);
    if (hostId) params.append('host_id', hostId.toString());
    
    return await apiClient.get(`/metrics?${params.toString()}`);
  }

  async getLatestMetric(hostId: number, name?: string): Promise<MetricData> {
    const params = new URLSearchParams();
    params.append('host_id', hostId.toString());
    if (name) params.append('name', name);
    
    return await apiClient.get(`/metrics/latest?${params.toString()}`);
  }

  async collectMetric(data: {
    host_id: number;
    name: string;
    value: number;
    labels?: Record<string, string>;
  }): Promise<void> {
    return await apiClient.post('/metrics/collect', data);
  }
}

export default new MetricsService();