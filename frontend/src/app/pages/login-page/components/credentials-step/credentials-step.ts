import { Component, inject, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { EmailInputComponent, PasswordInputComponent, MessageComponent } from '@app/shared';
import { LoginStore } from '../../store';

@Component({
  selector: 'app-login-credentials-step',
  standalone: true,
  imports: [CommonModule, EmailInputComponent, PasswordInputComponent, MessageComponent],
  templateUrl: './credentials-step.html',
  styleUrls: ['./credentials-step.scss'],
})
export class CredentialsStep {
  public emailInputSg = viewChild(EmailInputComponent);
  public passwordInputSg = viewChild(PasswordInputComponent);

  public store = inject(LoginStore);

  public emailSg = signal<string>('');
  public passwordSg = signal<string>('');

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

    const email = this.emailSg();
    const password = this.passwordSg();

    // Validate credentials and send OTP
    await this.store.validateCredentials(email, password);
  }
}
