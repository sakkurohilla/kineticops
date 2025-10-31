import { useState, useEffect, useCallback } from 'react';
import logsService, { LogFilters } from '../services/api/logsService';
import { handleApiError } from '../utils/errorHandler';
import { Log } from '../types';



interface UseLogsReturn {
  logs: Log[];
  isLoading: boolean;
  error: string | null;
  filters: LogFilters & { startDate?: string; endDate?: string };
  setFilters: (filters: LogFilters & { startDate?: string; endDate?: string }) => void;
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
  const [filters, setFilters] = useState<LogFilters & { startDate?: string; endDate?: string }>({});
  const [isTailing, setIsTailing] = useState(false);
  const [hasMore, setHasMore] = useState(true);

  const fetchLogs = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      const logsData = await logsService.searchLogs({
        level: filters.level,
        search: filters.search,
        start: filters.startDate,
        end: filters.endDate,
        limit: 100
      }) as Log[];
      
      setLogs(logsData);
      setHasMore(logsData.length >= 100);
    } catch (err: any) {
      const apiError = handleApiError(err);
      setError(apiError.message);
      setLogs([]);
    } finally {
      setIsLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  // Real-time log updates via WebSocket
  useEffect(() => {
    if (!isTailing) return;

    const { subscribe } = require('../services/ws/manager');
    const unsubscribe = subscribe((data: any) => {
      if (data.type === 'log' && data.log) {
        setLogs(prev => [data.log, ...prev.slice(0, 999)]); // Keep last 1000 logs
      }
    });

    // Also refresh periodically as fallback
    const interval = setInterval(fetchLogs, 30000);

    return () => {
      unsubscribe();
      clearInterval(interval);
    };
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
