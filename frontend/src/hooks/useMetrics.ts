import { useState, useEffect, useCallback } from 'react';
import metricsService from '../services/api/metricsService';
import { TimeRange } from '../components/metrics/TimeRangeSelector';
import useWebsocket from './useWebsocket';
import cache from '../utils/cache';
import { handleApiError } from '../utils/errorHandler';

interface MetricData {
  timestamp: string;
  value: number;
}

interface MetricsData {
  cpu: MetricData[];
  memory: MetricData[];
  disk: MetricData[];
  network: MetricData[];
  network_in: MetricData[];
  network_out: MetricData[];
  load_1min: MetricData[];
  load_5min: MetricData[];
  load_15min: MetricData[];
  uptime: MetricData[];
}

export const useMetrics = (timeRange: TimeRange, hostId?: number, autoRefresh = true, customStart?: string, customEnd?: string) => {
  const [data, setData] = useState<MetricsData>({
    cpu: [],
    memory: [],
    disk: [],
    network: [],
    network_in: [],
    network_out: [],
    load_1min: [],
    load_5min: [],
    load_15min: [],
    uptime: [],
  });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string>('');

  const fetchMetrics = useCallback(async () => {
    try {
      setError('');
      
      // Check cache first
      const cacheKey = `metrics-${timeRange}-${hostId || 'all'}-${customStart || ''}-${customEnd || ''}`;
      const cached = cache.get<MetricsData>(cacheKey);
      if (cached && !autoRefresh && timeRange !== 'custom') {
        setData(cached);
        return;
      }

      // Fetch metrics from backend (now returns aggregated data)
      console.log(`[useMetrics] Fetching ${timeRange} metrics for host ${hostId || 'all'}`);
      const aggregatedData = timeRange === 'custom' && customStart && customEnd
        ? await metricsService.getMetricsCustomRange(customStart, customEnd, hostId)
        : await metricsService.getMetricsRange(timeRange, hostId);
      
      console.log(`[useMetrics] Received aggregated data:`, aggregatedData);
      
      // Handle new aggregated format: {cpu_usage: [...], memory_usage: [...], etc}
      const grouped: MetricsData = {
        cpu: [],
        memory: [],
        disk: [],
        network: [],
        network_in: [],
        network_out: [],
        load_1min: [],
        load_5min: [],
        load_15min: [],
        uptime: [],
      };

      // Process aggregated data format
      if (aggregatedData && typeof aggregatedData === 'object') {
        const aggData = aggregatedData as any;
        
        // Map aggregated metrics to our format
        if (aggData.cpu_usage && Array.isArray(aggData.cpu_usage)) {
          grouped.cpu = aggData.cpu_usage.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value || 0
          }));
        }
        
        if (aggData.memory_usage && Array.isArray(aggData.memory_usage)) {
          grouped.memory = aggData.memory_usage.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value || 0
          }));
        }
        
        if (aggData.disk_usage && Array.isArray(aggData.disk_usage)) {
          grouped.disk = aggData.disk_usage.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value || 0
          }));
        }
        
        if (aggData.network_bytes && Array.isArray(aggData.network_bytes)) {
          grouped.network = aggData.network_bytes.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value / (1024 * 1024) || 0  // Convert bytes to MB
          }));
        }
        
        if (aggData.network_in_bytes && Array.isArray(aggData.network_in_bytes)) {
          grouped.network_in = aggData.network_in_bytes.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value / (1024 * 1024) || 0
          }));
        }
        
        if (aggData.network_out_bytes && Array.isArray(aggData.network_out_bytes)) {
          grouped.network_out = aggData.network_out_bytes.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value / (1024 * 1024) || 0
          }));
        }
        
        if (aggData.load_1min && Array.isArray(aggData.load_1min)) {
          grouped.load_1min = aggData.load_1min.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value || 0
          }));
        }
        
        if (aggData.load_5min && Array.isArray(aggData.load_5min)) {
          grouped.load_5min = aggData.load_5min.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value || 0
          }));
        }
        
        if (aggData.load_15min && Array.isArray(aggData.load_15min)) {
          grouped.load_15min = aggData.load_15min.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value || 0
          }));
        }
        
        if (aggData.uptime_seconds && Array.isArray(aggData.uptime_seconds)) {
          grouped.uptime = aggData.uptime_seconds.map((item: any) => ({
            timestamp: item.timestamp,
            value: item.value / 3600 || 0
          }));
        }
      }

      // Sort by timestamp and log data for debugging
      Object.keys(grouped).forEach((key) => {
        grouped[key as keyof MetricsData].sort(
          (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
        );
        console.log(`[useMetrics] ${key}: ${grouped[key as keyof MetricsData].length} points`);
      });

      // Merge per-metric and cache result
      const merged: MetricsData = {
        cpu: grouped.cpu.length > 0 ? grouped.cpu : data.cpu,
        memory: grouped.memory.length > 0 ? grouped.memory : data.memory,
        disk: grouped.disk.length > 0 ? grouped.disk : data.disk,
        network: grouped.network.length > 0 ? grouped.network : data.network,
        network_in: grouped.network_in.length > 0 ? grouped.network_in : data.network_in,
        network_out: grouped.network_out.length > 0 ? grouped.network_out : data.network_out,
        load_1min: grouped.load_1min.length > 0 ? grouped.load_1min : data.load_1min,
        load_5min: grouped.load_5min.length > 0 ? grouped.load_5min : data.load_5min,
        load_15min: grouped.load_15min.length > 0 ? grouped.load_15min : data.load_15min,
        uptime: grouped.uptime.length > 0 ? grouped.uptime : data.uptime,
      };

      setData(merged);
      cache.set(cacheKey, merged, 60000); // Cache for 1 minute
    } catch (err: any) {
      const apiError = handleApiError(err);
      console.error('[useMetrics] Error fetching metrics:', apiError);
      setError(apiError.message);
      setData({
        cpu: [],
        memory: [],
        disk: [],
        network: [],
        network_in: [],
        network_out: [],
        load_1min: [],
        load_5min: [],
        load_15min: [],
        uptime: [],
      });
    } finally {
      setIsLoading(false);
    }
  }, [timeRange, hostId, customStart, customEnd]);

  useEffect(() => {
    fetchMetrics();

    // Auto-refresh every 30 seconds if enabled
    if (autoRefresh) {
      const interval = setInterval(fetchMetrics, 30000);
      return () => clearInterval(interval);
    }
  }, [fetchMetrics, autoRefresh]);

  // Real-time websocket updates: merge incoming metric events into local state
  useWebsocket((payload: any) => {
    try {
      // payload expected to contain fields like cpu_usage, memory_usage, disk_usage, network_in, network_out, host_id, timestamp
      const ts = payload.timestamp || new Date().toISOString();

      // If hostId is specified, ignore events from other hosts
      if (hostId && payload.host_id && Number(payload.host_id) !== Number(hostId)) return;

      setData((prev) => {
        const next = { ...prev };

        // CPU
        if (typeof payload.cpu_usage === 'number') {
          next.cpu = [...next.cpu, { timestamp: ts, value: Number(payload.cpu_usage) }];
        }

        // Memory
        if (typeof payload.memory_usage === 'number') {
          next.memory = [...next.memory, { timestamp: ts, value: Number(payload.memory_usage) }];
        }

        // Disk
        if (typeof payload.disk_usage === 'number') {
          next.disk = [...next.disk, { timestamp: ts, value: Number(payload.disk_usage) }];
        }

        // Network (use network_in + network_out as a combined throughput metric)
        if (typeof payload.network_in === 'number' || typeof payload.network_out === 'number') {
          const netVal = (Number(payload.network_in || 0) + Number(payload.network_out || 0));
          next.network = [...next.network, { timestamp: ts, value: netVal }];
          
          if (typeof payload.network_in === 'number') {
            next.network_in = [...next.network_in, { timestamp: ts, value: payload.network_in }];
          }
          if (typeof payload.network_out === 'number') {
            next.network_out = [...next.network_out, { timestamp: ts, value: payload.network_out }];
          }
        }
        
        // Load average metrics
        if (payload.load_average && typeof payload.load_average === 'string') {
          const loads = payload.load_average.split(' ').map(Number);
          if (loads.length >= 3) {
            next.load_1min = [...next.load_1min, { timestamp: ts, value: loads[0] }];
            next.load_5min = [...next.load_5min, { timestamp: ts, value: loads[1] }];
            next.load_15min = [...next.load_15min, { timestamp: ts, value: loads[2] }];
          }
        }
        
        // Uptime
        if (typeof payload.uptime === 'number') {
          next.uptime = [...next.uptime, { timestamp: ts, value: payload.uptime / 3600 }];
        }

        // Keep arrays trimmed to last 1000 points to avoid memory growth
        (['cpu','memory','disk','network','network_in','network_out','load_1min','load_5min','load_15min','uptime'] as const).forEach((k) => {
          if ((next as any)[k].length > 1000) {
            (next as any)[k] = (next as any)[k].slice(-1000);
          }
        });

        return next;
      });
    } catch (e) {
      console.warn('[useMetrics] failed to handle websocket payload', e);
    }
  });

  return { data, isLoading, error, refetch: fetchMetrics };
};
