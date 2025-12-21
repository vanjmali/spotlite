import { inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { VALIDATION_MESSAGES } from '@app/shared/validation';

// TODO: Consider using Angular validators or a validation library for better scalability

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
    const isValid = await this._auth.login(email, password);
    this.loadingSg.set(false);

    if (!isValid) {
      this.errorSg.set(VALIDATION_MESSAGES.INVALID_CREDENTIALS);
      return false;
    }

    // Credentials valid, send OTP
    await this._auth.sendOtp(email);
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
    const status = await this._auth.verifyOtp(email, code);
    this.loadingSg.set(false);

    switch (status) {
      case 'success':
        return true;
      case 'expired':
        this.errorSg.set(VALIDATION_MESSAGES.OTP_EXPIRED);
        return false;
      case 'invalid':
        this.errorSg.set(VALIDATION_MESSAGES.OTP_INVALID_CODE);
        return false;
    }
  }

  public async resendOtp(): Promise<void> {
    const email = this.emailSg();
    if (!email) return;
    await this._auth.sendOtp(email);
  }
}
