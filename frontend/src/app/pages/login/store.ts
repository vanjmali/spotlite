import { inject, Injectable, signal } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { VALIDATION_MESSAGES } from '@app/shared';

export type LoginStep = 'credentials' | 'otp' | 'reset';

@Injectable()
export class LoginStore {
  private _auth = inject(AuthService);
  private readonly storageKey = 'spotlite.login';

  public readonly stepSg = signal<LoginStep>('credentials');
  public readonly emailSg = signal<string>('');
  public readonly passwordSg = signal<string>('');
  public readonly loadingSg = signal(false);
  public readonly errorSg = signal<string | null>(null);

  constructor() {
    this.restoreFromSession();
  }

  public clearError(): void {
    this.errorSg.set(null);
  }

  public reset(): void {
    this.stepSg.set('credentials');
    this.emailSg.set('');
    this.passwordSg.set('');
    this.errorSg.set(null);
    this.persistToSession();
  }

  public clearOtpSession(): void {
    this.stepSg.set('credentials');
    this.emailSg.set('');
    this.errorSg.set(null);
    this.persistToSession();
  }

  public async validateCredentials(
    email: string,
    password: string
  ): Promise<{ success: boolean; error?: string; fields?: Record<string, string> }> {
    this.errorSg.set(null);

    // Validate credentials against backend
    this.loadingSg.set(true);
    const result = await this._auth.login(email, password);
    this.loadingSg.set(false);

    if (!result.success) {
      return {
        success: false,
        error: result.error || VALIDATION_MESSAGES.INVALID_CREDENTIALS,
        fields: result.fields,
      };
    }

    // Credentials valid, OTP sent to email
    this.emailSg.set(email);
    this.passwordSg.set(password);
    this.errorSg.set(null);
    this.stepSg.set('otp');
    this.persistToSession();
    return { success: true };
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

    this.persistToSession();
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

  private persistToSession(): void {
    try {
      sessionStorage.setItem(
        this.storageKey,
        JSON.stringify({ step: this.stepSg(), email: this.emailSg() })
      );
    } catch {
      // ignore storage errors
    }
  }

  private restoreFromSession(): void {
    try {
      const raw = sessionStorage.getItem(this.storageKey);
      if (!raw) return;
      const parsed = JSON.parse(raw) as { step?: LoginStep; email?: string };
      if (parsed.email) this.emailSg.set(parsed.email);
      if (parsed.step === 'otp' && parsed.email) {
        this.stepSg.set('otp');
      }
    } catch {
      // ignore parse errors
    }
  }
}
