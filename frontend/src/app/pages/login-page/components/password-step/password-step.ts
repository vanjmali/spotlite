import { Component, inject, signal, effect, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { EmailInputComponent, PasswordInputComponent } from '@app/shared';
import { LoginStore } from '../../store';

@Component({
  selector: 'app-password-step',
  standalone: true,
  imports: [CommonModule, RouterLink, EmailInputComponent, PasswordInputComponent],
  templateUrl: './password-step.html',
  styleUrls: ['./password-step.scss'],
})
export class PasswordStep {
  @ViewChild(EmailInputComponent) public emailInputSg!: EmailInputComponent;
  @ViewChild(PasswordInputComponent) public passwordInputSg!: PasswordInputComponent;

  public store = inject(LoginStore);

  public emailSg = signal<string>('');
  public passwordSg = signal<string>('');
  public loadingSg = this.store.loadingSg;

  constructor() {
    // Initialize email from store
    const storedEmail = this.store.emailSg();
    if (storedEmail) {
      this.emailSg.set(storedEmail);
    }

    // Clear store error when entering password step
    this.store.clearError();

    // Track store errors and set them on the password input
    effect(() => {
      const storeError = this.store.errorSg();
      if (storeError && this.passwordInputSg) {
        this.passwordInputSg.errorSg.set(storeError);
      }
    });
  }

  public submit(): void {
    // Validate both inputs
    const emailValidation = this.emailInputSg.validate();
    const passwordValidation = this.passwordInputSg.validate();

    if (!emailValidation.isValid || !passwordValidation.isValid) {
      return;
    }

    const password = this.passwordSg().trim();
    void this.store.loginWithPassword(password);
  }
}
