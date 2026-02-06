import { Component, inject, OnInit, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { OtpInputComponent, MessageComponent } from '@app/shared';
import { LoginStore } from '../../store';
import { RegistrationStepFooter } from '../../../registration-page/components';

@Component({
  selector: 'app-login-otp-step',
  standalone: true,
  imports: [CommonModule, OtpInputComponent, MessageComponent, RegistrationStepFooter],
  templateUrl: './otp-step.html',
  styleUrls: ['./otp-step.scss'],
})
export class OtpStep implements OnInit {
  public otpInputSg = viewChild(OtpInputComponent);

  public store = inject(LoginStore);
  private router = inject(Router);

  public codeSg = signal<string>('');
  public loadingSg = this.store.loadingSg;
  public successSg = signal<string | null>(null);

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
