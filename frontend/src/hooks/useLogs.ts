import { useState, useEffect, useCallback } from 'react';
import apiClient from '../services/api/client';
import { Log } from '../types';

interface LogFilters {
  level?: string;
  source?: string;
  startDate?: string;
  endDate?: string;
  search?: string;
}

interface UseLogsReturn {
  logs: Log[];
  isLoading: boolean;
  error: string | null;
  filters: LogFilters;
  setFilters: (filters: LogFilters) => void;
  isTailing: boolean;
  toggleTailing: () => void;
  refetch: () => void;
  hasMore: boolean;
  loadMore: () => void;
}

export const useLogs = (): UseLogsReturn => {
  const [logs, setLogs] = useState<Log[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<LogFilters>({});
  const [isTailing, setIsTailing] = useState(false);
  const [hasMore, setHasMore] = useState(true);

  const fetchLogs = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      const params = new URLSearchParams();
      if (filters.level) params.append('level', filters.level);
      if (filters.source) params.append('source', filters.source);
      if (filters.startDate) params.append('start_date', filters.startDate);
      if (filters.endDate) params.append('end_date', filters.endDate);
      if (filters.search) params.append('search', filters.search);

      const response = await apiClient.get(`/logs?${params.toString()}`);
      setLogs(response.data.logs || []);
      setHasMore(response.data.has_more || false);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to fetch logs');
      setLogs([]);
    } finally {
      setIsLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  // Auto-refresh when tailing is enabled
  useEffect(() => {
    if (!isTailing) return;

    const interval = setInterval(() => {
      fetchLogs();
    }, 5000); // Refresh every 5 seconds

    return () => clearInterval(interval);
  }, [isTailing, fetchLogs]);

  const toggleTailing = useCallback(() => {
    setIsTailing(prev => !prev);
  }, []);

  const refetch = useCallback(() => {
    fetchLogs();
  }, [fetchLogs]);

  const loadMore = useCallback(() => {
    // Implement pagination if needed
    console.log('Load more logs');
  }, []);

  return {
    logs,
    isLoading,
    error,
    filters,
    setFilters,
    isTailing,
    toggleTailing,
    refetch,
    hasMore,
    loadMore,
  };
};
