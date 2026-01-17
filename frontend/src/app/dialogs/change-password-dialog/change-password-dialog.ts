import { Component, signal, computed, output, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { DialogComponent, PasswordInputComponent, MessageComponent } from '../../shared';
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

  // Output event
  readonly closed = output<void>();

  private readonly authService = inject(AuthService);

  // Computed signal to check if passwords match
  readonly passwordsMatchSg = computed(() => {
    const newPassword = this.newPasswordSg();
    const confirmPassword = this.confirmPasswordSg();
    return newPassword && confirmPassword && newPassword === confirmPassword;
  });

  // Computed signal to show password mismatch error
  readonly showPasswordMismatchSg = computed(() => {
    const newPassword = this.newPasswordSg();
    const confirmPassword = this.confirmPasswordSg();
    return !this.passwordsMatchSg() && newPassword && confirmPassword;
  });

  // Computed signal to check if form is valid
  readonly isValidSg = computed(() => {
    const currentPassword = this.currentPasswordSg();
    const newPassword = this.newPasswordSg();
    const confirmPassword = this.confirmPasswordSg();
    return (
      currentPassword.length >= 10 &&
      newPassword.length >= 10 &&
      confirmPassword.length >= 10 &&
      newPassword === confirmPassword
    );
  });

  submit(): void {
    if (!this.isValidSg()) {
      this.errorSg.set('Please check all fields and ensure passwords match');
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
          // Close dialog on success
          this.cancel();
        } else {
          // Show error message
          this.errorSg.set(result.error || 'Failed to change password');
        }
      });
  }

  cancel(): void {
    this.closed.emit();
  }
}
