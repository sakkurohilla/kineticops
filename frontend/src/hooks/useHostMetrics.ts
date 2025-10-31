import { useState, useEffect } from 'react';
import hostService from '../services/api/hostService';
import useWebsocket from './useWebsocket';

export interface MetricData {
  cpu_usage?: number;
  memory_usage?: number;
  memory_total?: number;
  memory_used?: number;
  disk_usage?: number;
  disk_total?: number;
  disk_used?: number;
  network_in?: number;
  network_out?: number;
  uptime?: number;
  load_average?: string;
  timestamp?: string;
  seq?: number;
}

const useHostMetrics = (hostId: number | undefined, autoRefresh = true) => {
  const [metrics, setMetrics] = useState<MetricData | null>(null);
  const [series, setSeries] = useState<MetricData[]>([]);
  const [lastSeq, setLastSeq] = useState<number>(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  const fetchLatest = async () => {
    try {
      setError('');
      const res = await hostService.getLatestMetrics(hostId as number);
      if (res) {
        setMetrics(res);
      }
    } catch (e: any) {
      setError(e.message || 'Failed to fetch latest metrics');
    } finally {
      setLoading(false);
    }
  };

  const fetchHistory = async () => {
    try {
      const list = await hostService.getHostMetrics(hostId as number);
      if (Array.isArray(list) && list.length) {
        setSeries(list as MetricData[]);
      }
    } catch (e) {
      // ignore
    }
  };

  useEffect(() => {
    if (!hostId) return;

    // hydrate from localStorage to avoid UI flash
    try {
      const saved = localStorage.getItem(`host_${hostId}_last_metric`);
      if (saved) setMetrics(JSON.parse(saved));
      if (saved) {
        try {
          const parsed = JSON.parse(saved);
          if (parsed && parsed.seq) setLastSeq(Number(parsed.seq));
        } catch (e) {}
      }
      const savedSeries = localStorage.getItem(`host_${hostId}_series`);
      if (savedSeries) setSeries(JSON.parse(savedSeries));
    } catch (e) {
      // ignore
    }

    fetchLatest();
    fetchHistory();

    let interval: any;
    if (autoRefresh) {
      interval = setInterval(fetchLatest, 30000);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [hostId, autoRefresh]);

  // websocket realtime updates
  useWebsocket((data: any) => {
    try {
      if (!data || !hostId) return;
      const hid = data.host_id ?? data.hostId;
      if (!hid || parseInt(String(hid), 10) !== hostId) return;

      // Prefer monotonic sequence id ordering (server supplies `seq`). If not present, fall back to timestamp.
      const incomingSeq = data.seq ? Number(data.seq) : 0;
      if (incomingSeq && lastSeq && incomingSeq <= lastSeq) {
        return; // older or duplicate
      }

      if (incomingSeq) {
        setLastSeq(incomingSeq);
      } else {
        // fallback: timestamp ordering
        const incomingTsStr = data.timestamp || new Date().toISOString();
        const incomingTs = new Date(incomingTsStr).getTime();
        const currentTs = metrics && metrics.timestamp ? new Date(metrics.timestamp).getTime() : 0;
        if (currentTs !== 0 && incomingTs <= currentTs) {
          return;
        }
      }

      setMetrics((prev) => {
        const next: any = prev ? { ...prev } : {};
        [
          'cpu_usage',
          'memory_usage',
          'memory_total',
          'memory_used',
          'disk_usage',
          'disk_total',
          'disk_used',
          'network_in',
          'network_out',
          'uptime',
          'load_average',
          'timestamp',
        ].forEach((k) => {
          if (data[k] !== undefined) next[k] = data[k];
        });
        // keep seq if present
        if (data.seq !== undefined) next.seq = Number(data.seq);
        try {
          localStorage.setItem(`host_${hostId}_last_metric`, JSON.stringify(next));
        } catch (e) {}
        return next;
      });

      setSeries((prev) => {
        const next = prev ? prev.slice() : [];
        const point: MetricData = {
          cpu_usage: data.cpu_usage ?? 0,
          memory_usage: data.memory_usage ?? 0,
          memory_total: data.memory_total ?? 0,
          memory_used: data.memory_used ?? 0,
          disk_usage: data.disk_usage ?? 0,
          disk_total: data.disk_total ?? 0,
          disk_used: data.disk_used ?? 0,
          network_in: data.network_in ?? 0,
          network_out: data.network_out ?? 0,
          uptime: data.uptime ?? 0,
          load_average: data.load_average ?? '',
          timestamp: data.timestamp ?? new Date().toISOString(),
          seq: data.seq !== undefined ? Number(data.seq) : undefined,
        };
        next.push(point);
        if (next.length > 120) next.shift();
        try {
          localStorage.setItem(`host_${hostId}_series`, JSON.stringify(next));
        } catch (e) {}
        return next;
      });
    } catch (e) {
      // ignore
    }
  });

  return { metrics, series, loading, error, refetch: fetchLatest };
};

export default useHostMetrics;
