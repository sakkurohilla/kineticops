// Auth Types
export interface User {
  id: number;
  username: string;
  email: string;
  created_at?: string;
}

export interface AuthResponse {
  token: string;
  refresh_token: string;
  user?: User;
  msg?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

// Host Types - Matching backend structure
export interface Host {
  id: number;
  hostname?: string;
  name?: string;
  ip: string;
  os?: string;
  group?: string;
  tags?: string;
  status: string;
  agent_status?: string;
  tenant_id?: number;
  reg_token?: string;
  last_seen?: string;
  created_at?: string;
  updated_at?: string;
}

// Metrics Types
export interface Metric {
  id: number;
  host_id: number;
  metric_name: string;
  metric_value: number;
  unit?: string;
  labels?: string;
  timestamp: string;
  tenant_id?: number;
}

export interface MetricAggregate {
  bucket: string;
  min: number;
  max: number;
  avg: number;
  p95: number;
  anomaly: boolean;
}

// Logs Types
export interface Log {
  id: string;
  timestamp: string;
  level: string;
  host_id: string;
  source: string;
  message: string;
  metadata?: Record<string, any>;
}


// Alerts Types - Matching backend structure
export interface Alert {
  id: number;
  host_id: number;
  host_name?: string;
  rule_id?: number;
  alert_type?: string;
  type?: string;
  severity: string;
  message: string;
  status: string;
  triggered_at?: string;
  resolved_at?: string;
  created_at: string;
  tenant_id?: number;
}

// API Response Types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
  msg?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

// Dashboard specific types
export interface DashboardStats {
  totalHosts: number;
  onlineHosts: number;
  warnings: number;
  critical: number;
}

export interface ActivityItem {
  id: number;
  host: string;
  message: string;
  type: 'warning' | 'success' | 'info' | 'error';
  time: string;
}
