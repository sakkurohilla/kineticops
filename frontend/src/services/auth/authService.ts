import apiClient, { BASE_URL } from '../api/client';
import { AuthResponse, User } from '../../types';

class AuthService {
  // Login with username
  async login(username: string, password: string): Promise<AuthResponse> {
    console.log('[AuthService] Attempting login for:', username);
    console.log('[AuthService] API Base URL:', BASE_URL);
    
    try {
      const response = await apiClient.post('/auth/login', {
        username,
        password,
      }) as AuthResponse;
      
      console.log('[AuthService] Login successful');
      
      // Store tokens (backend returns refresh_token with underscore)
      this.setTokens(response.token, response.refresh_token);
      
      return response;
    } catch (error: any) {
      console.error('[AuthService] Login failed:', error);
      throw error;
    }
  }

  // Register
  async register(username: string, email: string, password: string): Promise<AuthResponse> {
    console.log('[AuthService] Attempting registration for:', username);
    
    const response = await apiClient.post('/auth/register', {
      username,
      email,
      password,
    }) as AuthResponse;
    
    console.log('[AuthService] Registration successful');
    
    return response;
  }

  // Refresh token
  async refreshToken(): Promise<string> {
    const refreshToken = this.getRefreshToken();
    
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }
    
    console.log('[AuthService] Refreshing token...');
    
    const response = await apiClient.post('/auth/refresh', {
      refresh_token: refreshToken,
    }) as { token: string };
    
    this.setToken(response.token);
    console.log('[AuthService] Token refreshed successfully');
    
    return response.token;
  }

  // Get current user
  async getCurrentUser(): Promise<User> {
    console.log('[AuthService] Fetching current user...');
    const response = await apiClient.get('/auth/me') as User;
    return response;
  }

  // Logout
  logout(): void {
    console.log('[AuthService] Logging out...');
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
  }

  // Forgot password - Request reset token
  async forgotPassword(email: string): Promise<{ msg: string; token?: string }> {
    console.log('[AuthService] Requesting password reset for:', email);
    const response = await apiClient.post('/auth/forgot-password', { email }) as { msg: string; token?: string };
    return response;
  }

  // Verify reset token
  async verifyResetToken(token: string): Promise<{ valid: boolean; email: string; expires_at: string }> {
    console.log('[AuthService] Verifying reset token...');
    const response = await apiClient.post('/auth/verify-reset-token', { token }) as { valid: boolean; email: string; expires_at: string };
    return response;
  }

  // Reset password with token
  async resetPassword(token: string, newPassword: string): Promise<{ msg: string }> {
    console.log('[AuthService] Resetting password...');
    const response = await apiClient.post('/auth/reset-password', {
      token,
      new_password: newPassword,
    }) as { msg: string };
    return response;
  }

  // Token management
  private setTokens(token: string, refreshToken: string): void {
    localStorage.setItem('token', token);
    localStorage.setItem('refreshToken', refreshToken);
    console.log('[AuthService] Tokens stored');
  }

  private setToken(token: string): void {
    localStorage.setItem('token', token);
    console.log('[AuthService] Access token updated');
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
