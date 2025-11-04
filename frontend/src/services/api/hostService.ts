import apiClient from './client';
import { Host } from '../../types';
import cache from '../../utils/cache';
import { handleApiError } from '../../utils/errorHandler';

export interface CreateHostRequest {
  hostname: string;
  ip: string;
  ssh_user: string;
  ssh_password: string;
  ssh_port: number;
  os?: string;
  group?: string;
  tags?: string;
  description?: string;
}

export interface TestSSHRequest {
  ip: string;
  port: number;
  username: string;
  password: string;
  private_key?: string;
}

export interface AgentSetupRequest {
  setup_method: 'automatic' | 'manual' | 'none';
  hostname: string;
  ip: string;
  username: string;
  password?: string;
  ssh_key?: string;
  port: number;
}

export interface AgentSetupResponse {
  host_id: number;
  agent_id: number;
  token: string;
  setup_method: string;
  install_script?: string;
  status: string;
  message: string;
}

const hostService = {
  // Test SSH connection before creating host
  testSSHConnection: async (data: TestSSHRequest): Promise<{ success: boolean; message?: string; error?: string }> => {
    try {
      const response: any = await apiClient.post('/hosts/test-ssh', data);
      return response;
    } catch (error: any) {
      throw error;
    }
  },

  // Create new host
  createHost: async (data: CreateHostRequest): Promise<Host> => {
    try {
      const response: any = await apiClient.post('/hosts', data);
      cache.delete('hosts-all');
      return response;
    } catch (error: any) {
      const apiError = handleApiError(error);
      throw apiError;
    }
  },

  // Get all hosts
  getAllHosts: async (): Promise<Host[]> => {
    try {
      const cacheKey = 'hosts-all';
      const cached = cache.get<Host[]>(cacheKey);
      if (cached) return cached;

      const response: any = await apiClient.get('/hosts');
      const hosts = Array.isArray(response) ? response : (response.data || []);
      cache.set(cacheKey, hosts, 120000);
      return hosts;
    } catch (error: any) {
      const apiError = handleApiError(error);
      console.error('Failed to fetch hosts:', apiError);
      throw apiError;
    }
  },

  // Get single host
  getHost: async (id: number): Promise<Host> => {
    try {
      const response: any = await apiClient.get(`/hosts/${id}`);
      return response;
    } catch (error: any) {
      throw error;
    }
  },

  // Update host
  updateHost: async (id: number, data: Partial<Host>): Promise<any> => {
    try {
      const response: any = await apiClient.put(`/hosts/${id}`, data);
      return response;
    } catch (error: any) {
      throw error;
    }
  },

  // Delete host
  deleteHost: async (id: number): Promise<any> => {
    try {
      console.log('Deleting host:', id);
      const response: any = await apiClient.delete(`/hosts/${id}`);
      console.log('Delete response:', response);
      
      // Clear all related cache
      cache.delete('hosts-all');
      cache.delete(`host-${id}`);
      cache.delete(`metrics-latest-${id}`);
      
      return response;
    } catch (error: any) {
      console.error('Delete host error:', error);
      const apiError = handleApiError(error);
      throw apiError;
    }
  },

  // Get host metrics
  getHostMetrics: async (id: number): Promise<any[]> => {
    try {
      const response: any = await apiClient.get(`/hosts/${id}/metrics`);
      return Array.isArray(response) ? response : (response.data || []);
    } catch (error: any) {
      console.error('Failed to fetch metrics:', error);
      return [];
    }
  },

  // Get latest metrics for host
  getLatestMetrics: async (id: number): Promise<any> => {
    try {
      const cacheKey = `metrics-latest-${id}`;
      const cached = cache.get(cacheKey);
      if (cached) return cached;

      const response: any = await apiClient.get(`/hosts/${id}/metrics/latest`);
      cache.set(cacheKey, response, 30000);
      return response;
    } catch (error: any) {
      console.error('Failed to fetch latest metrics:', error);
      return null;
    }
  },

  // Create host with agent setup
  createHostWithAgent: async (data: AgentSetupRequest): Promise<AgentSetupResponse> => {
    try {
      const response: any = await apiClient.post('/hosts/with-agent', data);
      cache.delete('hosts-all');
      return response;
    } catch (error: any) {
      const apiError = handleApiError(error);
      throw apiError;
    }
  },

  // Get agent status
  getAgentStatus: async (hostId: number): Promise<any> => {
    try {
      const response: any = await apiClient.get(`/hosts/${hostId}/agent/status`);
      return response;
    } catch (error: any) {
      console.error('Failed to fetch agent status:', error);
      return null;
    }
  },

  // Get host services
  getHostServices: async (hostId: number): Promise<any[]> => {
    try {
      const response: any = await apiClient.get(`/hosts/${hostId}/services`);
      return Array.isArray(response) ? response : (response.data || []);
    } catch (error: any) {
      console.error('Failed to fetch host services:', error);
      return [];
    }
  },

  // Alias for getAllHosts
  getHosts: async (): Promise<Host[]> => {
    return hostService.getAllHosts();
  },
};

export default hostService;