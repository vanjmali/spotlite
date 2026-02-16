import { Component, signal, output, inject, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import {
  applyFieldErrors,
  DialogComponent,
  PasswordInputComponent,
  MessageComponent,
  VALIDATION_MESSAGES,
} from '../../shared';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-change-password-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatIconModule,
    DialogComponent,
    PasswordInputComponent,
    MessageComponent,
  ],
  templateUrl: './change-password-dialog.html',
  styleUrls: ['./change-password-dialog.scss'],
})
export class ChangePasswordDialogComponent {
  readonly currentPasswordSg = signal<string>('');
  readonly newPasswordSg = signal<string>('');
  readonly confirmPasswordSg = signal<string>('');
  readonly errorSg = signal<string>('');
  readonly loadingSg = signal<boolean>(false);
  readonly currentPasswordInputSg = viewChild('currentPasswordInput', {
    read: PasswordInputComponent,
  });
  readonly newPasswordInputSg = viewChild('newPasswordInput', { read: PasswordInputComponent });
  readonly confirmPasswordInputSg = viewChild('confirmPasswordInput', {
    read: PasswordInputComponent,
  });

  // Output event
  readonly closed = output<void>();
  readonly saved = output<void>();

  private readonly authService = inject(AuthService);

  submit(): void {
    const currentPasswordInput = this.currentPasswordInputSg();
    const newPasswordInput = this.newPasswordInputSg();
    const confirmPasswordInput = this.confirmPasswordInputSg();
    if (!currentPasswordInput || !newPasswordInput || !confirmPasswordInput) return;

    const currentPasswordValidation = currentPasswordInput.validate();
    const newPasswordValidation = newPasswordInput.validate();
    const confirmPasswordValidation = confirmPasswordInput.validate();
    if (
      !currentPasswordValidation.isValid ||
      !newPasswordValidation.isValid ||
      !confirmPasswordValidation.isValid
    ) {
      return;
    }

    if (this.newPasswordSg() !== this.confirmPasswordSg()) {
      confirmPasswordInput.setExternalError(VALIDATION_MESSAGES.PASSWORDS_MISMATCH);
      return;
    }

    this.loadingSg.set(true);
    this.errorSg.set('');

    // Call the change password service
    this.authService
      .changePassword(this.currentPasswordSg(), this.newPasswordSg())
      .subscribe((result) => {
        this.loadingSg.set(false);
        if (result.success) {
          this.saved.emit();
          // Close dialog on success
          this.cancel();
        } else {
          if (result.code === 'invalid_current_password') {
            currentPasswordInput.setExternalError(result.error || 'Invalid current password.');
            return;
          }

          if (result.code === 'password_same_as_current') {
            newPasswordInput.setExternalError(
              result.error || 'New password must be different from current password.'
            );
            return;
          }

          if (result.fields) {
            applyFieldErrors(result.fields, {
              current_password: (msg) => currentPasswordInput.setExternalError(msg),
              new_password: (msg) => newPasswordInput.setExternalError(msg),
              confirm_password: (msg) => confirmPasswordInput.setExternalError(msg),
            });
            return;
          }

          this.errorSg.set(result.error || 'Failed to change password');
        }
      });
  }

  cancel(): void {
    this.closed.emit();
  }
}
