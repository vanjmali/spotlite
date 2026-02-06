import { Component, inject, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import {
  EmailInputComponent,
  PasswordInputComponent,
  MessageComponent,
  applyFieldErrors,
} from '@app/shared';
import { LoginStore } from '../../store';
import { RegistrationStepFooter } from '../../../registration-page/components';

@Component({
  selector: 'app-login-credentials-step',
  standalone: true,
  imports: [CommonModule, EmailInputComponent, PasswordInputComponent, MessageComponent, RegistrationStepFooter],
  templateUrl: './credentials-step.html',
  styleUrls: ['./credentials-step.scss'],
})
export class CredentialsStep {
  public emailInputSg = viewChild(EmailInputComponent);
  public passwordInputSg = viewChild(PasswordInputComponent);

  public store = inject(LoginStore);
  private router = inject(Router);

  public emailSg = this.store.emailSg;
  public passwordSg = this.store.passwordSg;
  public loadingSg = this.store.loadingSg;

  public async submit(): Promise<void> {
    this.store.clearError();

    const emailInput = this.emailInputSg();
    const passwordInput = this.passwordInputSg();

    if (!emailInput || !passwordInput) return;

    const emailValidation = emailInput.validate();
    const passwordValidation = passwordInput.validate();

    if (!emailValidation.isValid || !passwordValidation.isValid) {
      return;
    }

    const email = emailInput.valueSg();
    const password = passwordInput.valueSg();

    // Validate credentials and send OTP
    const result = await this.store.validateCredentials(email, password);
    if (!result.success) {
      if (result.fields) {
        applyFieldErrors(result.fields, {
          email: (msg) => emailInput.setExternalError(msg),
          password: (msg) => passwordInput.setExternalError(msg),
        });
        return;
      }

      this.store.errorSg.set(result.error || 'Login failed');
      return;
    }

    this.router.navigate(['/login/otp']);
  }
}
