import { useState, useEffect, useCallback } from 'react';
import workflowApi, { WorkflowSessionRequest, WorkflowSessionResponse } from '../services/api/workflowApi';

export const useWorkflowSession = () => {
  const [session, setSession] = useState<WorkflowSessionResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');

  const createSession = useCallback(async (data: WorkflowSessionRequest) => {
    setLoading(true);
    setError('');
    try {
      const response = await workflowApi.createWorkflowSession(data);
      setSession(response);
      return response;
    } catch (err: any) {
      setError(err.message || 'Failed to create session');
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const closeSession = useCallback(async () => {
    if (session) {
      try {
        await workflowApi.closeSession(session.session_token);
        setSession(null);
      } catch (err) {
        console.error('Failed to close session:', err);
      }
    }
  }, [session]);

  const isExpired = useCallback(() => {
    if (!session) return true;
    return new Date(session.expires_at) <= new Date();
  }, [session]);

  return {
    session,
    loading,
    error,
    createSession,
    closeSession,
    isExpired
  };
};

export const useServices = (hostId: number, sessionToken?: string) => {
  const [services, setServices] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');

  const discoverServices = useCallback(async () => {
    if (!sessionToken) return;
    
    setLoading(true);
    setError('');
    try {
      const response = await workflowApi.discoverServices(hostId, sessionToken);
      setServices(response.services || []);
    } catch (err: any) {
      setError(err.message || 'Failed to discover services');
    } finally {
      setLoading(false);
    }
  }, [hostId, sessionToken]);

  useEffect(() => {
    if (sessionToken) {
      discoverServices();
    }
  }, [discoverServices, sessionToken]);

  return {
    services,
    loading,
    error,
    refetch: discoverServices
  };
};

export const useServiceControl = (sessionToken?: string) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');

  const controlService = useCallback(async (serviceId: number, action: 'start' | 'stop' | 'restart' | 'enable' | 'disable') => {
    if (!sessionToken) throw new Error('No session token');

    setLoading(true);
    setError('');
    try {
      const response = await workflowApi.controlService(serviceId, { action }, sessionToken);
      return response;
    } catch (err: any) {
      setError(err.message || 'Failed to control service');
      throw err;
    } finally {
      setLoading(false);
    }
  }, [sessionToken]);

  return {
    controlService,
    loading,
    error
  };
};