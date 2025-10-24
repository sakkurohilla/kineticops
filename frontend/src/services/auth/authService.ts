import apiClient from '../api/client';
import { AuthResponse, User } from '../../types';

class AuthService {
  // Login with username
  async login(username: string, password: string): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/login', {
      username,
      password,
    });
    
    // Store tokens (backend returns refresh_token with underscore)
    this.setTokens(response.token, response.refresh_token);
    
    return response;
  }

  // Register
  async register(username: string, email: string, password: string): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/register', {
      username,
      email,
      password,
    });
    
    // Backend returns msg, not tokens on register
    // User needs to login after registration
    
    return response;
  }

  // Refresh token
  async refreshToken(): Promise<string> {
    const refreshToken = this.getRefreshToken();
    
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }
    
    const response = await apiClient.post<{ token: string }>('/auth/refresh', {
      refresh_token: refreshToken,
    });
    
    this.setToken(response.token);
    
    return response.token;
  }

  // Get current user (you'll need to add this endpoint to backend)
  async getCurrentUser(): Promise<User> {
    return await apiClient.get<User>('/auth/me');
  }

  // Logout
  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
  }

  // Forgot password
  async forgotPassword(email: string): Promise<{ msg: string }> {
    return await apiClient.post('/auth/forgot-password', { email });
  }

  // Token management
  private setTokens(token: string, refreshToken: string): void {
    localStorage.setItem('token', token);
    localStorage.setItem('refreshToken', refreshToken);
  }

  private setToken(token: string): void {
    localStorage.setItem('token', token);
  }

  getToken(): string | null {
    return localStorage.getItem('token');
  }

  private getRefreshToken(): string | null {
    return localStorage.getItem('refreshToken');
  }

  isAuthenticated(): boolean {
    return !!this.getToken();
  }
}

export default new AuthService();
