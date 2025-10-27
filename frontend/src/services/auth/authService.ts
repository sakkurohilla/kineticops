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

  // Get current user
  async getCurrentUser(): Promise<User> {
    return await apiClient.get<User>('/auth/me');
  }

  // Logout
  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
  }

  // Forgot password - Request reset token
  async forgotPassword(email: string): Promise<{ msg: string; token?: string }> {
    return await apiClient.post('/auth/forgot-password', { email });
  }

  // Verify reset token
  async verifyResetToken(token: string): Promise<{ valid: boolean; email: string; expires_at: string }> {
    return await apiClient.post('/auth/verify-reset-token', { token });
  }

  // Reset password with token
  async resetPassword(token: string, newPassword: string): Promise<{ msg: string }> {
    return await apiClient.post('/auth/reset-password', {
      token,
      new_password: newPassword,
    });
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