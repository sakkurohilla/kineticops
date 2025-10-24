interface AppConfig {
  apiBaseUrl: string;
  wsUrl: string;
  env: string;
}

const config: AppConfig = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  wsUrl: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws',
  env: import.meta.env.VITE_ENV || 'development',
};

export default config;
