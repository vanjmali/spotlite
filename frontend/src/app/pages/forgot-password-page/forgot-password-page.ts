import { Component, inject, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, Router } from '@angular/router';
import { EmailInputComponent, MessageComponent, LoaderComponent, BrandLogo } from '@app/shared';
import { PasswordRecoveryService } from '../../services/password-recovery.service';

@Component({
  selector: 'app-forgot-password-page',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    EmailInputComponent,
    MessageComponent,
    LoaderComponent,
    BrandLogo,
  ],
  templateUrl: './forgot-password-page.html',
  styleUrls: ['./forgot-password-page.scss'],
})
export class ForgotPasswordPage {
  private readonly _recoveryService = inject(PasswordRecoveryService);
  private readonly _router = inject(Router);

  public emailInputSg = viewChild(EmailInputComponent);

  public emailSg = signal<string>('');
  public isLoadingSg = signal(false);
  public errorSg = signal<string | null>(null);
  public isSuccessSg = signal(false);

  public submit(): void {
    this.errorSg.set(null);

    const emailInput = this.emailInputSg();
    if (!emailInput) return;

    const emailValidation = emailInput.validate();
    if (!emailValidation.isValid) {
      return;
    }

    const email = this.emailSg();
    this.isLoadingSg.set(true);

    this._recoveryService.requestReset(email).subscribe({
      next: () => {
        this.isSuccessSg.set(true);
        this.emailSg.set('');
        // Auto-redirect to login page after 3 seconds
        setTimeout(() => {
          this._router.navigate(['/login']);
        }, 3000);
      },
      error: (err: Error) => {
        this.isLoadingSg.set(false);
        this.errorSg.set(err.message);
      },
      complete: () => this.isLoadingSg.set(false),
    });
  }
}
