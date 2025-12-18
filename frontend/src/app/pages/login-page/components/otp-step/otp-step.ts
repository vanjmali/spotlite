import { Component, inject, signal, effect, ViewChild, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { OtpInputComponent } from '@app/shared/input';
import { LoginStore } from '../../store';

@Component({
  selector: 'app-otp-step',
  standalone: true,
  imports: [CommonModule, OtpInputComponent],
  templateUrl: './otp-step.html',
  styleUrls: ['./otp-step.scss'],
})
export class OtpStep implements OnDestroy {
  @ViewChild(OtpInputComponent) public otpInputSg!: OtpInputComponent;

  public store = inject(LoginStore);

  public codeSg = signal<string>('');
  public loadingSg = this.store.loadingSg;

  private readonly errorSyncEffect = effect(() => {
    const storeError = this.store.errorSg();
    if (storeError && this.otpInputSg) {
      this.otpInputSg.errorSg.set(storeError);
    }
  });

  constructor() {
    // Clear store error when entering OTP step
    this.store.clearError();
  }

  public async verify(): Promise<void> {
    // Validate OTP input
    const validation = this.otpInputSg.validate();
    if (!validation.isValid) {
      return;
    }

    const code = this.codeSg();
    const ok = await this.store.verifyOtp(code);
    if (!ok) return;
    // success: store sets authenticated state; nothing else here
  }

  public async resend(): Promise<void> {
    await this.store.resendOtp();
  }

  public ngOnDestroy(): void {
    this.errorSyncEffect.destroy();
  }
}
