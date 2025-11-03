import apiClient from './client';

export interface WorkflowSessionRequest {
  host_id: number;
  username: string;
  password?: string;
  ssh_key?: string;
}

export interface WorkflowSessionResponse {
  session_token: string;
  expires_at: string;
  host_id: number;
  status: string;
}

export interface ServiceControlRequest {
  action: 'start' | 'stop' | 'restart' | 'enable' | 'disable';
}

export interface ServiceControlResponse {
  success: boolean;
  output: string;
  error?: string;
}

const workflowApi = {
  createWorkflowSession: async (data: WorkflowSessionRequest): Promise<WorkflowSessionResponse> => {
    const response: any = await apiClient.post('/workflow/session', data);
    return response;
  },

  getWorkflowData: async (hostId: number, sessionToken: string): Promise<any> => {
    const response: any = await apiClient.get(`/hosts/${hostId}/workflow`, {
      headers: { 'X-Session-Token': sessionToken }
    });
    return response;
  },

  discoverServices: async (hostId: number, sessionToken: string): Promise<any> => {
    const response: any = await apiClient.post(`/workflow/${hostId}/discover`, {}, {
      headers: { 'X-Session-Token': sessionToken }
    });
    return response;
  },

  controlService: async (serviceId: number, action: ServiceControlRequest, sessionToken: string): Promise<ServiceControlResponse> => {
    const response: any = await apiClient.post(`/services/${serviceId}/control`, action, {
      headers: { 'X-Session-Token': sessionToken }
    });
    return response;
  },

  getServiceStatus: async (serviceId: number, sessionToken: string): Promise<any> => {
    const response: any = await apiClient.get(`/services/${serviceId}/status`, {
      headers: { 'X-Session-Token': sessionToken }
    });
    return response;
  },

  closeSession: async (sessionToken: string): Promise<void> => {
    await apiClient.delete('/workflow/session', {
      headers: { 'X-Session-Token': sessionToken }
    });
  }
};

export default workflowApi;