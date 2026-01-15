import { Component, inject, signal, viewChild, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, ActivatedRoute, RouterLink } from '@angular/router';
import { PasswordInputComponent, MessageComponent, LoaderComponent } from '@app/shared';
import { PasswordRecoveryService } from '../../services/password-recovery.service';
import { VALIDATION_MESSAGES } from '@app/shared';

@Component({
  selector: 'app-reset-password-page',
  standalone: true,
  imports: [CommonModule, RouterLink, PasswordInputComponent, MessageComponent, LoaderComponent],
  templateUrl: './reset-password-page.html',
  styleUrls: ['./reset-password-page.scss'],
})
export class ResetPasswordPage implements OnInit {
  private readonly _route = inject(ActivatedRoute);
  private readonly _recoveryService = inject(PasswordRecoveryService);
  private readonly _router = inject(Router);

  public passwordInputSg = viewChild(PasswordInputComponent);
  public confirmPasswordInputSg = viewChild(PasswordInputComponent);

  public tokenSg = signal<string | null>(null);
  public passwordSg = signal<string>('');
  public confirmPasswordSg = signal<string>('');
  public isValidatingSg = signal(true);
  public isTokenValidSg = signal(false);
  public isLoadingSg = signal(false);
  public isSuccessSg = signal(false);
  public errorSg = signal<string | null>(null);

  ngOnInit(): void {
    this._route.queryParams.subscribe((params) => {
      const token = params['token'];
      if (!token) {
        this.isValidatingSg.set(false);
        return;
      }

      this.tokenSg.set(token);
      this.validateToken(token);
    });
  }

  public clearError(): void {
    this.errorSg.set(null);
  }

  public validateToken(token: string): void {
    this._recoveryService.validateToken(token).subscribe({
      next: () => {
        this.isTokenValidSg.set(true);
        this.isValidatingSg.set(false);
      },
      error: () => {
        this.isTokenValidSg.set(false);
        this.isValidatingSg.set(false);
      },
    });
  }

  public submit(): void {
    this.clearError();

    const passwordInput = this.passwordInputSg();
    const confirmInput = this.confirmPasswordInputSg();

    if (!passwordInput || !confirmInput) return;

    const passwordValidation = passwordInput.validate();
    const confirmValidation = confirmInput.validate();

    if (!passwordValidation.isValid || !confirmValidation.isValid) {
      return;
    }

    const password = this.passwordSg();
    const confirmPassword = this.confirmPasswordSg();

    if (password !== confirmPassword) {
      this.errorSg.set(VALIDATION_MESSAGES.PASSWORDS_MISMATCH);
      return;
    }

    const token = this.tokenSg();
    if (!token) {
      this.errorSg.set(VALIDATION_MESSAGES.INVALID_RECOVERY_LINK);
      return;
    }

    this.isLoadingSg.set(true);

    this._recoveryService.resetPassword(token, password).subscribe({
      next: () => {
        this.isSuccessSg.set(true);
        // Redirect after 3 seconds
        setTimeout(() => {
          this._router.navigate(['/login']);
        }, 3000);
      },
      error: (err: Error) => {
        this.errorSg.set(err.message);
      },
      complete: () => this.isLoadingSg.set(false),
    });
  }
}
