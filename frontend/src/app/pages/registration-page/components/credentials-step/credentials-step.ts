import { Component, inject, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  TextInputComponent,
  PasswordInputComponent,
  MessageComponent,
  USERNAME_PATTERN,
  VALIDATION_MESSAGES,
  applyFieldErrors,
} from '@app/shared';
import { RegistrationStore } from '../../store';
import { RegistrationStepFooter } from '../step-footer';

@Component({
  selector: 'app-registration-credentials-step',
  standalone: true,
  imports: [
    CommonModule,
    TextInputComponent,
    PasswordInputComponent,
    MessageComponent,
    RegistrationStepFooter,
  ],
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

  public usernameSg = this.store.usernameSg;
  public passwordSg = this.store.passwordSg;
  public confirmPasswordSg = this.store.confirmPasswordSg;
  public loadingSg = this.store.loadingSg;
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

    if (!usernameValidation.isValid || !passwordValidation.isValid) {
      return;
    }

    // Check if passwords match
    const password = this.passwordSg();
    const confirmPassword = this.confirmPasswordSg();

    if (password !== confirmPassword) {
      confirmPasswordInput.setExternalError(VALIDATION_MESSAGES.PASSWORDS_MISMATCH);
      return;
    }

    // Save credentials and navigate
    const username = this.usernameSg();
    const result = await this.store.saveCredentials(username, password);
    if (!result.success) {
      if (result.code === 'username_taken') {
        usernameInput.setExternalError(result.error || VALIDATION_MESSAGES.USERNAME_IN_USE);
        return;
      }

      if (result.fields) {
        applyFieldErrors(result.fields, {
          username: (msg) => usernameInput.setExternalError(msg),
          password: (msg) => passwordInput.setExternalError(msg),
        });
        return;
      }

      this.store.errorSg.set(result.error || VALIDATION_MESSAGES.REGISTRATION_FAILED);
    }
  }

}
