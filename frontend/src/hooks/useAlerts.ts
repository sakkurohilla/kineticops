import { useCallback, useEffect, useState } from 'react';
import alertsService, { Alert, AlertRule } from '../services/api/alertsService';
import { handleApiError } from '../utils/errorHandler';

export type AlertRuleRequest = Omit<AlertRule, 'id' | 'created_at'>;

export default function useAlerts() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchAlerts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await alertsService.getAlerts();
      setAlerts(data);
    } catch (err: any) {
      const apiError = handleApiError(err);
      console.error('[useAlerts] fetchAlerts error', apiError);
      setError(apiError.message);
      setAlerts([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchRules = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await alertsService.getRules();
      setRules(data);
    } catch (err: any) {
      const apiError = handleApiError(err);
      console.error('[useAlerts] fetchRules error', apiError);
      setError(apiError.message);
      setRules([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const createRule = useCallback(async (body: AlertRuleRequest) => {
    try {
      const created = await alertsService.createRule(body);
      setRules((prev) => [created, ...prev]);
      return created;
    } catch (err: any) {
      const apiError = handleApiError(err);
      console.error('[useAlerts] createRule error', apiError);
      throw apiError;
    }
  }, []);

  const silenceAlert = useCallback(async (alertId: number) => {
    try {
      await alertsService.silenceAlert(alertId);
      setAlerts((prev) => prev.map((a) => (a.id === alertId ? { ...a, status: 'silenced' } : a)));
    } catch (err: any) {
      const apiError = handleApiError(err);
      console.error('[useAlerts] silenceAlert error', apiError);
      // Still update local state as fallback
      setAlerts((prev) => prev.map((a) => (a.id === alertId ? { ...a, status: 'silenced' } : a)));
      throw apiError;
    }
  }, []);

  useEffect(() => {
    fetchAlerts();
    fetchRules();
    // refresh periodically
    const t = setInterval(fetchAlerts, 30000);
    return () => clearInterval(t);
  }, [fetchAlerts, fetchRules]);

  return {
    alerts,
    rules,
    loading,
    error,
    fetchAlerts,
    fetchRules,
    createRule,
    silenceAlert,
  };
}
