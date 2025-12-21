import { Injectable, inject, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { VALIDATION_MESSAGES } from '@app/shared';

export interface LoginResponse {
  message: string;
}

export interface VerifyOtpResponse {
  access_token: string;
  refresh_token: string;
}

export type OtpVerificationStatus = 'success' | 'invalid' | 'expired';

@Injectable({ providedIn: 'root' })
export class AuthService {
  // Public signals for reactive state
  readonly currentEmailSg = signal<string | null>(null);
  readonly accessTokenSg = signal<string | null>(null);
  readonly refreshTokenSg = signal<string | null>(null);
  readonly isAuthenticatedSg = computed(() => !!this.accessTokenSg());

  private readonly API_BASE = 'http://localhost:3000/api/users'; // Traefik API Gateway on port 3000
  private readonly http = inject(HttpClient);

  // Register new user with backend
  async register(
    firstName: string,
    lastName: string,
    email: string,
    username: string,
    password: string
  ): Promise<{ success: boolean; error?: string }> {
    try {
      await firstValueFrom(
        this.http.post(`${this.API_BASE}/register`, {
          firstName,
          lastName,
          email,
          username,
          password,
        })
      );

      return { success: true };
    } catch (error: unknown) {
      const httpError = error as {
        error?: { message?: string; errors?: Array<{ message: string }> };
      };
      const errorMsg =
        httpError?.error?.message ||
        httpError?.error?.errors?.[0]?.message ||
        VALIDATION_MESSAGES.REGISTRATION_FAILED;
      return { success: false, error: errorMsg };
    }
  }

  // Check if email already exists
  async checkEmailExists(email: string): Promise<boolean> {
    try {
      const response = await firstValueFrom(
        this.http.get<{ exists: boolean }>(
          `${this.API_BASE}/check-email/${encodeURIComponent(email)}`
        )
      );
      return response.exists;
    } catch {
      return false;
    }
  }

  // Login with email and password - backend sends OTP via email
  async login(email: string, password: string): Promise<{ success: boolean; error?: string }> {
    try {
      // Clear any previous session state before new login attempt
      this.logout();

      await firstValueFrom(
        this.http.post<LoginResponse>(`${this.API_BASE}/login`, { email, password })
      );

      // If we reach here, response was successful (2xx status)
      this.currentEmailSg.set(email);
      return { success: true };
    } catch (error: unknown) {
      const httpError = error as { error?: { message?: string } };
      const errorMsg = httpError?.error?.message || VALIDATION_MESSAGES.LOGIN_FAILED;
      return { success: false, error: errorMsg };
    }
  }

  // Verify OTP code sent to email - returns JWT and refresh token
  async verifyOtp(
    email: string,
    code: string
  ): Promise<{ success: boolean; error?: string; status?: OtpVerificationStatus }> {
    try {
      const response = await firstValueFrom(
        this.http.post<VerifyOtpResponse>(`${this.API_BASE}/login/verify-otp`, { email, code })
      );

      if (response?.access_token) {
        this.storeTokens(response.access_token, response.refresh_token, email);
        return { success: true, status: 'success' };
      }
      return { success: false, status: 'invalid' };
    } catch (error: unknown) {
      const httpError = error as { status?: number };
      const status = httpError?.status === 401 ? 'invalid' : 'expired';
      return { success: false, status };
    }
  }

  // Store tokens securely - all tokens and email in localStorage
  private storeTokens(access: string, refresh: string, email: string): void {
    this.accessTokenSg.set(access);
    this.refreshTokenSg.set(refresh);
    this.currentEmailSg.set(email);
    localStorage.setItem('access_token', access);
    localStorage.setItem('refresh_token', refresh);
    localStorage.setItem('user_email', email);
  }

  // Check if access token is expired
  private isTokenExpired(token: string | null = this.accessTokenSg()): boolean {
    if (!token) return true;

    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      // exp is in seconds, Date.now() is in milliseconds
      return payload.exp * 1000 < Date.now();
    } catch {
      return true;
    }
  }

  // Initialize auth on app startup - restore session from localStorage
  initializeAuth(): void {
    const accessToken = localStorage.getItem('access_token');
    const refreshToken = localStorage.getItem('refresh_token');
    const userEmail = localStorage.getItem('user_email');

    // Check if access token is expired
    if (accessToken && this.isTokenExpired(accessToken)) {
      this.logout();
      return;
    }

    if (accessToken) {
      this.accessTokenSg.set(accessToken);
    }
    if (refreshToken) {
      this.refreshTokenSg.set(refreshToken);
    }
    if (userEmail) {
      this.currentEmailSg.set(userEmail);
    }
  }

  // Refresh access token using refresh token
  async refreshAccessToken(): Promise<boolean> {
    const refreshToken = this.refreshTokenSg();
    if (!refreshToken) return false;

    try {
      const response = await firstValueFrom(
        this.http.post<VerifyOtpResponse>(`${this.API_BASE}/refresh-token`, {
          refresh_token: refreshToken,
        })
      );
      this.accessTokenSg.set(response.access_token);
      localStorage.setItem('access_token', response.access_token);
      return true;
    } catch {
      this.logout();
      return false;
    }
  }

  // Resend OTP code to email during login
  async resendOtp(email: string): Promise<{ success: boolean; error?: string }> {
    try {
      await firstValueFrom(
        this.http.post<LoginResponse>(`${this.API_BASE}/login/resend-otp`, { email })
      );

      // If we reach here, response was successful (2xx status)
      return { success: true };
    } catch (error: unknown) {
      const httpError = error as { error?: { message?: string } };
      const errorMsg = httpError?.error?.message || VALIDATION_MESSAGES.OTP_RESEND_FAILED;
      return { success: false, error: errorMsg };
    }
  }

  // Logout - clear tokens and signals
  logout(): void {
    this.accessTokenSg.set(null);
    this.refreshTokenSg.set(null);
    this.currentEmailSg.set(null);
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_email');
  }
}
