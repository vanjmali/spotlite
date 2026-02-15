import { Injectable, inject, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom, map, catchError, of } from 'rxjs';
import { getApiErrorInfo, VALIDATION_MESSAGES } from '@app/shared';
import { environment } from '../../environments/environment';
import { NotificationService } from './notification.service';

export interface LoginResponse {
  message: string;
}

export interface VerifyOtpResponse {
  access_token: string;
}

export type OtpVerificationStatus = 'success' | 'invalid' | 'expired';

@Injectable({ providedIn: 'root' })
export class AuthService {
  // Public signals for reactive state
  readonly currentEmailSg = signal<string | null>(null);
  readonly accessTokenSg = signal<string | null>(null);
  readonly isAuthenticatedSg = computed(() => !!this.accessTokenSg());
  readonly isAdminSg = computed(() => {
    const claims = this.getTokenClaims(this.accessTokenSg());
    if (!claims) {
      return false;
    }

    if (claims.is_admin === true) {
      return true;
    }

    if (typeof claims.role === 'string' && claims.role.toLowerCase() === 'admin') {
      return true;
    }

    if (
      Array.isArray(claims.roles) &&
      claims.roles.some((role) => role.toLowerCase() === 'admin')
    ) {
      return true;
    }

    return false;
  });
  readonly profileInitialSg = computed(() => {
    const email = this.currentEmailSg();
    if (email) {
      return email.charAt(0).toUpperCase();
    }

    const claims = this.getTokenClaims(this.accessTokenSg());
    const candidate = [claims?.name, claims?.username, claims?.email, claims?.sub]
      .find((value) => typeof value === 'string' && value.trim().length > 0)
      ?.toString()
      .trim();

    return candidate ? candidate.charAt(0).toUpperCase() : 'U';
  });

  readonly notificationService = inject(NotificationService);

  private readonly API_BASE = environment.apiBaseUrl;
  private readonly http = inject(HttpClient);

  // Register new user with backend
  async register(
    first_name: string,
    last_name: string,
    email: string,
    username: string,
    password: string
  ): Promise<{ success: boolean; error?: string; code?: string; fields?: Record<string, string> }> {
    try {
      await firstValueFrom(
        this.http.post(`${this.API_BASE}/users/register`, {
          first_name,
          last_name,
          email,
          username,
          password,
        })
      );

      return { success: true };
    } catch (error: unknown) {
      const info = getApiErrorInfo(error, VALIDATION_MESSAGES.REGISTRATION_FAILED);
      return {
        success: false,
        error: info.userMessage,
        code: info.code,
        fields: info.fields,
      };
    }
  }

  // Check if email already exists
  async checkEmailExists(email: string): Promise<boolean> {
    try {
      const response = await firstValueFrom(
        this.http.get<{ exists: boolean }>(
          `${this.API_BASE}/users/check-email?email=${encodeURIComponent(email)}`
        )
      );
      return response.exists;
    } catch {
      return false;
    }
  }

  // Login with email and password - backend sends OTP via email
  async login(
    email: string,
    password: string
  ): Promise<{ success: boolean; error?: string; code?: string; fields?: Record<string, string> }> {
    try {
      this.accessTokenSg.set(null);
      this.currentEmailSg.set(null);

      await firstValueFrom(
        this.http.post<LoginResponse>(`${this.API_BASE}/users/login`, { email, password })
      );

      // If we reach here, response was successful (2xx status)
      this.currentEmailSg.set(email);
      return { success: true };
    } catch (error: unknown) {
      const info = getApiErrorInfo(error, VALIDATION_MESSAGES.LOGIN_FAILED);
      return {
        success: false,
        error: info.userMessage,
        code: info.code,
        fields: info.fields,
      };
    }
  }

  // Verify OTP code sent to email - returns access token
  async verifyOtp(
    email: string,
    code: string
  ): Promise<{ success: boolean; error?: string; status?: OtpVerificationStatus }> {
    try {
      const response = await firstValueFrom(
        this.http.post<VerifyOtpResponse>(
          `${this.API_BASE}/users/login/verify-otp`,
          { email, code },
          { withCredentials: true }
        )
      );

      if (response?.access_token) {
        this.storeAccessToken(response.access_token, email);
        return { success: true, status: 'success' };
      }
      return { success: false, status: 'invalid' };
    } catch (error: unknown) {
      const httpError = error as { status?: number };
      const status = httpError?.status === 401 ? 'invalid' : 'expired';
      return { success: false, status };
    }
  }

  // Store access token in memory only
  private storeAccessToken(access: string, email: string): void {
    this.accessTokenSg.set(access);
    this.currentEmailSg.set(email);
  }

  // Initialize auth on app startup - refresh access token using httpOnly cookie
  initializeAuth(): void {
    void this.refreshAccessToken();
  }

  // Refresh access token using httpOnly refresh cookie
  async refreshAccessToken(): Promise<boolean> {
    try {
      const response = await firstValueFrom(
        this.http.post<VerifyOtpResponse>(
          `${this.API_BASE}/users/refresh-token`,
          {},
          { withCredentials: true }
        )
      );
      this.accessTokenSg.set(response.access_token);
      return true;
    } catch {
      return false;
    }
  }

  // Resend OTP code to email during login
  async resendOtp(email: string): Promise<{ success: boolean; error?: string }> {
    try {
      await firstValueFrom(
        this.http.post<LoginResponse>(`${this.API_BASE}/users/login/resend-otp`, { email })
      );

      // If we reach here, response was successful (2xx status)
      return { success: true };
    } catch (error: unknown) {
      const httpError = error as { error?: { message?: string } };
      const errorMsg = httpError?.error?.message || VALIDATION_MESSAGES.OTP_RESEND_FAILED;
      return { success: false, error: errorMsg };
    }
  }

  // Verify account email token
  async verifyAccount(token: string): Promise<{ success: boolean; error?: string; code?: string }> {
    try {
      await firstValueFrom(
        this.http.post(`${this.API_BASE}/users/verify`, {
          token,
        })
      );
      return { success: true };
    } catch (error: unknown) {
      const info = getApiErrorInfo(error, VALIDATION_MESSAGES.VERIFICATION_FAILED);
      return {
        success: false,
        error: info.userMessage,
        code: info.code,
      };
    }
  }

  // Logout - clear tokens and signals
  logout(): void {
    void firstValueFrom(
      this.http.post(
        `${this.API_BASE}/users/logout`,
        {},
        {
          withCredentials: true,
        }
      )
    ).catch(() => undefined);

    this.accessTokenSg.set(null);
    this.currentEmailSg.set(null);
    this.notificationService.closeConnection();
  }

  // Change password
  changePassword(currentPassword: string, newPassword: string) {
    return this.http
      .patch(
        `${this.API_BASE}/users/change-password`,
        {
          current_password: currentPassword,
          new_password: newPassword,
        },
        {
          withCredentials: true,
        }
      )
      .pipe(
        map(() => ({
          success: true as const,
          error: undefined,
          code: undefined as string | undefined,
          fields: undefined as Record<string, string> | undefined,
        })),
        catchError((error: unknown) => {
          const info = getApiErrorInfo(error, 'Failed to change password');
          return of({
            success: false as const,
            error: info.userMessage,
            code: info.code,
            fields: info.fields,
          });
        })
      );
  }

  private getTokenClaims(token: string | null): {
    name?: string;
    username?: string;
    email?: string;
    sub?: string;
    role?: string;
    roles?: string[];
    is_admin?: boolean;
  } | null {
    if (!token) return null;

    const payload = token.split('.')[1];
    if (!payload) return null;

    try {
      const normalized = payload.replace(/-/g, '+').replace(/_/g, '/');
      const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');
      const decoded = atob(padded);
      return JSON.parse(decoded) as {
        name?: string;
        username?: string;
        email?: string;
        sub?: string;
      };
    } catch {
      return null;
    }
  }
}
