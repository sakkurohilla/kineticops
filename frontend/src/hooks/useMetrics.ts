import { useState, useEffect, useCallback } from 'react';
import apiClient from '../services/api/client';
import { TimeRange } from '../components/metrics/TimeRangeSelector';

interface MetricData {
  timestamp: string;
  value: number;
}

interface MetricsData {
  cpu: MetricData[];
  memory: MetricData[];
  disk: MetricData[];
  network: MetricData[];
}

export const useMetrics = (timeRange: TimeRange, hostId?: number, autoRefresh = true) => {
  const [data, setData] = useState<MetricsData>({
    cpu: [],
    memory: [],
    disk: [],
    network: [],
  });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string>('');

  const fetchMetrics = useCallback(async () => {
    try {
      setError('');
      
      // Build query parameters
      const params = new URLSearchParams();
      params.append('range', timeRange);
      if (hostId) params.append('host_id', hostId.toString());

      // Fetch metrics from backend
      const response: any = await apiClient.get(`/metrics?${params.toString()}`);
      
      // Transform backend data to chart format
      const metricsData = response.data || response || [];
      
      // Group by metric type
      const grouped: MetricsData = {
        cpu: [],
        memory: [],
        disk: [],
        network: [],
      };

      metricsData.forEach((metric: any) => {
        const dataPoint = {
          timestamp: metric.timestamp || metric.created_at,
          value: parseFloat(metric.metric_value || metric.value || 0),
        };

        const metricName = (metric.metric_name || metric.type || '').toLowerCase();
        
        if (metricName.includes('cpu')) {
          grouped.cpu.push(dataPoint);
        } else if (metricName.includes('memory') || metricName.includes('mem')) {
          grouped.memory.push(dataPoint);
        } else if (metricName.includes('disk')) {
          grouped.disk.push(dataPoint);
        } else if (metricName.includes('network') || metricName.includes('net')) {
          grouped.network.push(dataPoint);
        }
      });

      // Sort by timestamp
      Object.keys(grouped).forEach((key) => {
        grouped[key as keyof MetricsData].sort(
          (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
        );
      });

      setData(grouped);
    } catch (err: any) {
      console.error('[useMetrics] Error fetching metrics:', err);
      setError(err.message || 'Failed to load metrics');
      // Set empty data on error
      setData({
        cpu: [],
        memory: [],
        disk: [],
        network: [],
      });
    } finally {
      setIsLoading(false);
    }
  }, [timeRange, hostId]);

  useEffect(() => {
    fetchMetrics();

    // Auto-refresh every 30 seconds if enabled
    if (autoRefresh) {
      const interval = setInterval(fetchMetrics, 30000);
      return () => clearInterval(interval);
    }
  }, [fetchMetrics, autoRefresh]);

  return { data, isLoading, error, refetch: fetchMetrics };
};
