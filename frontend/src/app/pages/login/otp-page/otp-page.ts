import { Component, effect, inject, OnInit, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthLayout, OtpInputComponent, MessageComponent, OTP_PATTERN } from '@app/shared';
import { FormFooter } from '@app/shared';
import { LoginStore } from '../store';

@Component({
  selector: 'app-login-otp-page',
  standalone: true,
  imports: [CommonModule, AuthLayout, OtpInputComponent, MessageComponent, FormFooter],
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
  private readonly lastAutoSubmitCodeSg = signal<string>('');
  private readonly autoSubmitEffect = effect(() => {
    if (!this.store.emailSg()) return;

    const code = this.codeSg().trim();
    if (code.length !== 6) return;
    if (!OTP_PATTERN.test(code)) return;
    if (this.loadingSg()) return;
    if (code === this.lastAutoSubmitCodeSg()) return;

    this.lastAutoSubmitCodeSg.set(code);
    queueMicrotask(() => this.verify());
  });

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
