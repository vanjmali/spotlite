import { inject, Injectable, signal } from '@angular/core';
import { VALIDATION_MESSAGES } from '@app/shared';
import { AuthService } from '../../services/auth.service';
import type { RegistrationStep } from './registration-page';

@Injectable()
export class RegistrationStore {
  private _auth = inject(AuthService);

  public readonly loadingSg = signal(false);
  public readonly errorSg = signal<string | null>(null);
  public readonly stepSg = signal<RegistrationStep>('personal');

  // Store personal info for use across steps
  public readonly firstNameSg = signal<string>('');
  public readonly lastNameSg = signal<string>('');
  public readonly emailSg = signal<string>('');

  // Store credentials
  public readonly usernameSg = signal<string>('');
  public readonly passwordSg = signal<string>('');
  public readonly confirmPasswordSg = signal<string>('');

  public clearError(): void {
    this.errorSg.set(null);
  }

  public setStep(step: RegistrationStep): void {
    this.stepSg.set(step);
  }

  public async savePersonalInfo(
    firstName: string,
    lastName: string,
    email: string
  ): Promise<{ success: boolean; error?: string; errorField?: 'email' }> {
    this.errorSg.set(null);
    this.loadingSg.set(true);

    // Check if email already exists
    const emailExists = await this._auth.checkEmailExists(email);
    this.loadingSg.set(false);

    if (emailExists) {
      return {
        success: false,
        error: VALIDATION_MESSAGES.EMAIL_IN_USE,
        errorField: 'email',
      };
    }

    this.firstNameSg.set(firstName);
    this.lastNameSg.set(lastName);
    this.emailSg.set(email);

    this.stepSg.set('credentials');

    return { success: true };
  }

  public async saveCredentials(
    username: string,
    password: string
  ): Promise<{ success: boolean; error?: string; code?: string }> {
    this.errorSg.set(null);
    this.loadingSg.set(true);

    // Register user - backend will check if username exists
    const result = await this._auth.register(
      this.firstNameSg(),
      this.lastNameSg(),
      this.emailSg(),
      username,
      password
    );
    this.loadingSg.set(false);

    if (!result.success) {
      return {
        success: false,
        error: result.error || VALIDATION_MESSAGES.REGISTRATION_FAILED,
        code: result.code,
      };
    }

    this.usernameSg.set(username);
    this.passwordSg.set(password);
    this.confirmPasswordSg.set('');
    this.stepSg.set('verify');

    return { success: true };
  }
}
