import { Component, inject, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  TextInputComponent,
  PasswordInputComponent,
  MessageComponent,
  USERNAME_PATTERN,
  VALIDATION_MESSAGES,
} from '@app/shared';
import { RegistrationStore } from '../../store';

@Component({
  selector: 'app-registration-credentials-step',
  standalone: true,
  imports: [CommonModule, TextInputComponent, PasswordInputComponent, MessageComponent],
  templateUrl: './credentials-step.html',
  styleUrls: ['./credentials-step.scss'],
})
export class RegistrationCredentialsStep {
  public usernameInputSg = viewChild('usernameInput', { read: TextInputComponent });
  public passwordInputSg = viewChild('passwordInput', { read: PasswordInputComponent });
  public confirmPasswordInputSg = viewChild('confirmPasswordInput', {
    read: PasswordInputComponent,
  });

  public store = inject(RegistrationStore);

  public usernameSg = signal<string>('');
  public passwordSg = signal<string>('');
  public confirmPasswordSg = signal<string>('');
  public readonly usernamePattern = USERNAME_PATTERN;
  public readonly validationMessages = VALIDATION_MESSAGES;

  public async submit(): Promise<void> {
    this.store.clearError();

    const usernameInput = this.usernameInputSg();
    const passwordInput = this.passwordInputSg();
    const confirmPasswordInput = this.confirmPasswordInputSg();

    if (!usernameInput || !passwordInput || !confirmPasswordInput) return;

    // Validate all inputs
    const usernameValidation = usernameInput.validate();
    const passwordValidation = passwordInput.validate();
    const confirmPasswordValidation = confirmPasswordInput.validate();

    if (
      !usernameValidation.isValid ||
      !passwordValidation.isValid ||
      !confirmPasswordValidation.isValid
    ) {
      return;
    }

    // Check if passwords match
    const password = this.passwordSg();
    const confirmPassword = this.confirmPasswordSg();

    if (password !== confirmPassword) {
      this.store.errorSg.set(VALIDATION_MESSAGES.PASSWORDS_MISMATCH);
      return;
    }

    // Save credentials and navigate
    const username = this.usernameSg();
    this.store.saveCredentials(username, password);
  }
}
