import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, viewChild } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { AuthFormFooter } from '@app/pages/registration-page/components';
import {
  AuthLayout,
  EmailInputComponent,
  MessageComponent,
  PasswordInputComponent,
  applyFieldErrors,
} from '@app/shared';
import { LoginStore } from '../store';

@Component({
  selector: 'app-login-credentials-page',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatIconModule,
    AuthLayout,
    EmailInputComponent,
    PasswordInputComponent,
    MessageComponent,
    AuthFormFooter,
  ],
  templateUrl: './credentials-page.html',
  styleUrls: ['./credentials-page.scss'],
})
export class CredentialsPage implements OnInit {
  public emailInputSg = viewChild(EmailInputComponent);
  public passwordInputSg = viewChild(PasswordInputComponent);

  public store = inject(LoginStore);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  public emailSg = this.store.emailSg;
  public passwordSg = this.store.passwordSg;
  public loadingSg = this.store.loadingSg;
  public readonly title = this.route.snapshot.data?.['authTitle'] ?? 'Welcome back';
  public verificationNoticeSg = signal(false);

  public ngOnInit(): void {
    this.store.clearOtpSession();
  }

  public async submit(): Promise<void> {
    this.store.clearError();
    this.verificationNoticeSg.set(false);

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
      if (result.code === 'verification_required') {
        this.store.errorSg.set(null);
        this.verificationNoticeSg.set(true);
        return;
      }

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
