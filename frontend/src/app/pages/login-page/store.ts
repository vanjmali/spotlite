import { inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { VALIDATION_MESSAGES } from '@app/shared';

@Injectable()
export class LoginStore {
  private _auth = inject(AuthService);
  private _router = inject(Router);

  public readonly emailSg = signal<string | null>(null);
  public readonly loadingSg = signal(false);
  public readonly errorSg = signal<string | null>(null);

  public clearError(): void {
    this.errorSg.set(null);
  }

  public async validateCredentials(email: string, password: string): Promise<boolean> {
    this.errorSg.set(null);

    // Validate credentials against backend
    this.loadingSg.set(true);
    const result = await this._auth.login(email, password);
    this.loadingSg.set(false);

    if (!result.success) {
      this.errorSg.set(result.error || VALIDATION_MESSAGES.INVALID_CREDENTIALS);
      return false;
    }

    // Credentials valid, OTP sent to email
    this.emailSg.set(email);
    this.errorSg.set(null);
    this._router.navigate(['/login/otp']);
    return true;
  }

  public async verifyOtp(code: string): Promise<boolean> {
    this.errorSg.set(null);
    const email = this.emailSg();
    if (!email) {
      this.errorSg.set(VALIDATION_MESSAGES.MISSING_EMAIL);
      return false;
    }
    this.loadingSg.set(true);
    const result = await this._auth.verifyOtp(email, code);
    this.loadingSg.set(false);

    if (!result.success) {
      switch (result.status) {
        case 'expired':
          this.errorSg.set(VALIDATION_MESSAGES.OTP_EXPIRED);
          break;
        case 'invalid':
          this.errorSg.set(VALIDATION_MESSAGES.OTP_INVALID_CODE);
          break;
        default:
          this.errorSg.set(result.error || VALIDATION_MESSAGES.OTP_VERIFICATION_FAILED);
      }
      return false;
    }

    // OTP verified successfully - redirect to home
    this._router.navigate(['/home']);
    return true;
  }

  public async resendOtp(): Promise<void> {
    this.errorSg.set(null);
    const email = this.emailSg();
    if (!email) {
      this.errorSg.set(VALIDATION_MESSAGES.MISSING_EMAIL);
      return;
    }
    this.loadingSg.set(true);
    const result = await this._auth.resendOtp(email);
    this.loadingSg.set(false);

    if (!result.success) {
      this.errorSg.set(result.error || VALIDATION_MESSAGES.OTP_RESEND_FAILED);
    }
  }
}
