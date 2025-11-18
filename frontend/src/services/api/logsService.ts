import apiClient from './client';

export interface LogEntry {
  id: string;
  host_id: number;
  timestamp: string;
  level: string;
  message: string;
  meta?: Record<string, string>;
  correlation_id?: string;
}

export interface LogFilters {
  level?: string;
  host_id?: number;
  start?: string;
  end?: string;
  search?: string;
  limit?: number;
  skip?: number;
}

class LogsService {
  async searchLogs(filters: LogFilters = {}): Promise<{ total: number; limit: number; skip: number; logs: LogEntry[] }> {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        params.append(key, value.toString());
      }
    });

    return await apiClient.get(`/logs?${params.toString()}`);
  }

  async getSources(): Promise<{ sources: string[]; levels: string[] }> {
    return await apiClient.get('/logs/sources');
  }

  async collectLog(data: {
    host_id: number;
    level: string;
    message: string;
    meta?: Record<string, string>;
    correlation_id?: string;
  }): Promise<void> {
    return await apiClient.post('/logs', data);
  }
}

export default new LogsService();