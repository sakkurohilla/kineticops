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

// Rest of your types remain the same...

// Host Types
export interface Host {
  id: string;
  name: string;
  ip: string;
  status: 'online' | 'offline' | 'maintenance';
  createdAt: string;
  updatedAt: string;
}

// Metrics Types
export interface Metric {
  id: string;
  hostId: string;
  type: string;
  value: number;
  timestamp: string;
  labels?: Record<string, string>;
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
  hostId: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  timestamp: string;
  metadata?: Record<string, any>;
}

// Alerts Types
export interface Alert {
  id: string;
  hostId: string;
  type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  message: string;
  status: 'active' | 'acknowledged' | 'resolved';
  createdAt: string;
}

// API Response Types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}
