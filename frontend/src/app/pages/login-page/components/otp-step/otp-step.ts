import { Component, inject, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { OtpInputComponent } from '@app/shared/input';
import { LoginStore } from '../../store';

@Component({
  selector: 'app-login-otp-step',
  standalone: true,
  imports: [CommonModule, OtpInputComponent],
  templateUrl: './otp-step.html',
  styleUrls: ['./otp-step.scss'],
})
export class OtpStep {
  public otpInputSg = viewChild(OtpInputComponent);

  public store = inject(LoginStore);

  public codeSg = signal<string>('');
  public loadingSg = this.store.loadingSg;

  public async verify(): Promise<void> {
    // Clear store error at start
    this.store.clearError();

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
    await this.store.resendOtp();
  }
}
