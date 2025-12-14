import { inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

// TODO: Consider using Angular validators or a validation library for better scalability
export interface ValidationResult {
  isValid: boolean;
  error?: string;
}

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

  public async checkEmailAndSendOtp(rawEmail: string): Promise<boolean> {
    this.errorSg.set(null);
    const email = (rawEmail ?? '').trim();

    if (!email) {
      this.errorSg.set('Please enter your email address.');
      return false;
    }

    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailPattern.test(email)) {
      this.errorSg.set('Invalid email address.');
      return false;
    }

    this.loadingSg.set(true);
    const res = await this._auth.checkEmail(email);
    this.loadingSg.set(false);
    if (!res.exists) {
      this.errorSg.set('User not found.');
      return false;
    }

    await this._auth.sendOtp(email);
    this.emailSg.set(email);
    this.errorSg.set(null); // Clear error before navigating to OTP
    this._router.navigate(['/login/otp']);
    return true;
  }

  public async validateCredentials(email: string, password: string): Promise<boolean> {
    this.errorSg.set(null);
    const trimmedEmail = (email ?? '').trim();
    const trimmedPassword = (password ?? '').trim();

    if (!trimmedEmail) {
      this.errorSg.set('Please enter your email address.');
      return false;
    }

    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailPattern.test(trimmedEmail)) {
      this.errorSg.set('Invalid email address.');
      return false;
    }

    if (!trimmedPassword) {
      this.errorSg.set('Please enter your password.');
      return false;
    }

    // Validate credentials against backend
    this.loadingSg.set(true);
    const isValid = await this._auth.login(trimmedEmail, trimmedPassword);
    this.loadingSg.set(false);

    if (!isValid) {
      this.errorSg.set('Invalid email or password.');
      return false;
    }

    // Credentials valid, send OTP
    await this._auth.sendOtp(trimmedEmail);
    this.emailSg.set(trimmedEmail);
    this.errorSg.set(null);
    this._router.navigate(['/login/otp']);
    return true;
  }

  public async verifyOtp(code: string): Promise<boolean> {
    this.errorSg.set(null);
    if (!/^[0-9]{6}$/.test(code)) {
      this.errorSg.set('Enter the 6-digit code');
      return false;
    }
    const email = this.emailSg();
    if (!email) {
      this.errorSg.set('Missing email');
      return false;
    }
    this.loadingSg.set(true);
    const status = await this._auth.verifyOtp(email, code);
    this.loadingSg.set(false);

    switch (status) {
      case 'success':
        return true;
      case 'expired':
        this.errorSg.set('Verification code has expired');
        return false;
      case 'invalid':
        this.errorSg.set('Invalid code');
        return false;
    }
  }

  public async resendOtp(): Promise<void> {
    const email = this.emailSg();
    if (!email) return;
    await this._auth.sendOtp(email);
  }
}
