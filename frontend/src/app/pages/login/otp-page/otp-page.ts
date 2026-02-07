import { Component, inject, OnInit, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthLayout, OtpInputComponent, MessageComponent } from '@app/shared';
import { AuthFormFooter } from '@app/pages/registration-page/components';
import { LoginStore } from '../store';

@Component({
  selector: 'app-login-otp-page',
  standalone: true,
  imports: [CommonModule, AuthLayout, OtpInputComponent, MessageComponent, AuthFormFooter],
  templateUrl: './otp-page.html',
  styleUrls: ['./otp-page.scss'],
})
export class OtpPage implements OnInit {
  public otpInputSg = viewChild(OtpInputComponent);

  public store = inject(LoginStore);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  public codeSg = signal<string>('');
  public loadingSg = this.store.loadingSg;
  public successSg = signal<string | null>(null);
  public readonly title = this.route.snapshot.data?.['authTitle'] ?? 'Verify your code';

  public ngOnInit(): void {
    if (!this.store.emailSg()) {
      this.router.navigate(['/login']);
    }
  }

  public async verify(): Promise<void> {
    // Clear store error at start
    this.store.clearError();
    this.successSg.set(null);

    const otpInput = this.otpInputSg();
    if (!otpInput) return;

    // Validate OTP input
    const validation = otpInput.validate();
    if (!validation.isValid) {
      return;
    }

    // Send trimmed code to store
    const code = this.codeSg().trim();
    const ok = await this.store.verifyOtp(code);
    if (!ok) {
      return;
    }
    this.router.navigate(['/']);
  }

  public async resend(): Promise<void> {
    this.successSg.set(null);
    await this.store.resendOtp();
    // Show success message if no error
    if (!this.store.errorSg()) {
      this.successSg.set('OTP resent successfully. Check your email.');
      // Clear success message after 4 seconds
      setTimeout(() => this.successSg.set(null), 4000);
    }
  }
}
