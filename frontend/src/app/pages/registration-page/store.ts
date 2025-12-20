import { inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { VALIDATION_MESSAGES } from '@app/shared/validation';

@Injectable()
export class RegistrationStore {
  private _auth = inject(AuthService);
  private _router = inject(Router);

  public readonly loadingSg = signal(false);
  public readonly errorSg = signal<string | null>(null);

  // Store personal info for use across steps
  public readonly firstNameSg = signal<string>('');
  public readonly lastNameSg = signal<string>('');
  public readonly emailSg = signal<string>('');

  // Store credentials
  public readonly usernameSg = signal<string>('');
  public readonly passwordSg = signal<string>('');

  public clearError(): void {
    this.errorSg.set(null);
  }

  public async savePersonalInfo(firstName: string, lastName: string, email: string): Promise<void> {
    this.errorSg.set(null);
    this.loadingSg.set(true);

    // Check if email already exists
    const emailExists = await this._auth.checkEmailExists(email);
    this.loadingSg.set(false);

    if (emailExists) {
      this.errorSg.set(VALIDATION_MESSAGES.EMAIL_IN_USE);
      return;
    }

    this.firstNameSg.set(firstName);
    this.lastNameSg.set(lastName);
    this.emailSg.set(email);

    // Navigate to credentials step
    this._router.navigate(['/register/credentials']);
  }

  public async saveCredentials(username: string, password: string): Promise<void> {
    this.errorSg.set(null);
    this.loadingSg.set(true);

    // Check if username already exists
    const usernameExists = await this._auth.checkUsernameExists(username);
    this.loadingSg.set(false);

    if (usernameExists) {
      this.errorSg.set(VALIDATION_MESSAGES.USERNAME_IN_USE);
      return;
    }

    // All checks passed, register user
    const registered = await this._auth.register(
      this.firstNameSg(),
      this.lastNameSg(),
      this.emailSg(),
      username,
      password
    );

    if (!registered) {
      this.errorSg.set(VALIDATION_MESSAGES.REGISTRATION_FAILED);
      return;
    }

    this.usernameSg.set(username);
    this.passwordSg.set(password);

    // Navigate to login page after successful registration
    this._router.navigate(['/login']);
  }
}
