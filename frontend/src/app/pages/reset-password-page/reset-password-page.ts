import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, viewChild } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import {
  AuthLayout,
  EmailInputComponent,
  LoaderComponent,
  MessageComponent,
  PasswordInputComponent,
  VALIDATION_MESSAGES,
} from '@app/shared';
import { PasswordRecoveryService } from '../../services/password-recovery.service';
import { AuthCheckEmail, AuthFormFooter } from '../registration-page/components';

@Component({
  selector: 'app-reset-password-page',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    AuthLayout,
    PasswordInputComponent,
    MessageComponent,
    LoaderComponent,
    EmailInputComponent,
    AuthCheckEmail,
    AuthFormFooter,
  ],
  templateUrl: './reset-password-page.html',
  styleUrls: ['./reset-password-page.scss'],
})
export class ResetPasswordPage implements OnInit {
  private readonly _route = inject(ActivatedRoute);
  private readonly _recoveryService = inject(PasswordRecoveryService);
  private readonly _router = inject(Router);

  public emailInputSg = viewChild(EmailInputComponent);
  public passwordInputSg = viewChild<PasswordInputComponent>('passwordInput');
  public confirmPasswordInputSg = viewChild<PasswordInputComponent>('confirmPasswordInput');

  public tokenSg = signal<string | null>(null);
  public emailSg = signal<string>('');
  public passwordSg = signal<string>('');
  public confirmPasswordSg = signal<string>('');
  public isValidatingSg = signal(true);
  public isTokenValidSg = signal(false);
  public isLoadingSg = signal(false);
  public isSuccessSg = signal(false);
  public isRequestSuccessSg = signal(false);
  public errorSg = signal<string | null>(null);
  public readonly title = this._route.snapshot.data?.['authTitle'] ?? 'Reset your password';


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
    if (!passwordValidation.isValid) {
      return;
    }

    const password = this.passwordSg();
    const confirmPassword = this.confirmPasswordSg();

    if (password !== confirmPassword) {
      confirmInput.setExternalError(VALIDATION_MESSAGES.PASSWORDS_MISMATCH);
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
        this.isLoadingSg.set(false);
      },
      complete: () => this.isLoadingSg.set(false),
    });
  }

  public requestReset(): void {
    this.clearError();

    const emailInput = this.emailInputSg();
    if (!emailInput) return;

    const validation = emailInput.validate();
    if (!validation.isValid) {
      return;
    }

    const email = this.emailSg();
    this.isLoadingSg.set(true);

    this._recoveryService.requestReset(email).subscribe({
      next: () => {
        this.isRequestSuccessSg.set(true);
        this.emailSg.set('');
      },
      error: (err: Error) => {
        this.isLoadingSg.set(false);
        this.errorSg.set(err.message);
      },
      complete: () => this.isLoadingSg.set(false),
    });
  }
}
