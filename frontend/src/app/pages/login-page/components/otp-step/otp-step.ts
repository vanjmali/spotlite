import { Component, inject, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { OtpInputComponent, MessageComponent } from '@app/shared';
import { LoginStore } from '../../store';

@Component({
  selector: 'app-login-otp-step',
  standalone: true,
  imports: [CommonModule, OtpInputComponent, MessageComponent],
  templateUrl: './otp-step.html',
  styleUrls: ['./otp-step.scss'],
})
export class OtpStep {
  public otpInputSg = viewChild(OtpInputComponent);

  public store = inject(LoginStore);

  public codeSg = signal<string>('');
  public loadingSg = this.store.loadingSg;
  public successSg = signal<string | null>(null);

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
    // success: store sets authenticated state; nothing else here
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
