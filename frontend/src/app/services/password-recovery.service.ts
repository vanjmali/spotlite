import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom, Observable, throwError } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { getApiErrorInfo, VALIDATION_MESSAGES } from '@app/shared';

@Injectable({ providedIn: 'root' })
export class PasswordRecoveryService {
  private readonly API_BASE = environment.apiBaseUrl;
  private readonly http = inject(HttpClient);

  /**
   * Request password reset link
   */
  requestReset(email: string): Observable<void> {
    return this.http
      .post<{ message: string }>(`${this.API_BASE}/users/password-recovery/request`, { email })
      .pipe(
        map(() => undefined),
        catchError((error) => {
          const errorMsg = error?.error?.message || VALIDATION_MESSAGES.PASSWORD_RESET_FAILED;
          return throwError(() => new Error(errorMsg));
        })
      );
  }

  /**
   * Validate recovery token
   */
  validateToken(token: string): Observable<boolean> {
    return this.http
      .post<{ message: string }>(`${this.API_BASE}/users/password-recovery/validate`, { token })
      .pipe(
        map(() => true),
        catchError(() => throwError(() => new Error(VALIDATION_MESSAGES.INVALID_RECOVERY_LINK)))
      );
  }

  /**
   * Reset password with token
   */
  async resetPassword(
    token: string,
    newPassword: string
  ): Promise<{ success: boolean; error?: string; code?: string; fields?: Record<string, string> }> {
    try {
      await firstValueFrom(
        this.http.post<{ message: string }>(`${this.API_BASE}/users/password-recovery/reset`, {
          token,
          new_password: newPassword,
        })
      );
      return { success: true };
    } catch (error: unknown) {
      const info = getApiErrorInfo(error, VALIDATION_MESSAGES.PASSWORD_RESET_FAILED);
      return {
        success: false,
        error: info.userMessage,
        code: info.code,
        fields: info.fields,
      };
    }
  }
}
