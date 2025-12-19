import { Component, inject, signal, effect, ViewChild, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { EmailInputComponent, PasswordInputComponent } from '@app/shared/input';
import { LoginStore } from '../../store';

@Component({
  selector: 'app-credentials-step',
  standalone: true,
  imports: [CommonModule, EmailInputComponent, PasswordInputComponent],
  templateUrl: './credentials-step.html',
  styleUrls: ['./credentials-step.scss'],
})
export class CredentialsStep implements OnDestroy {
  @ViewChild(EmailInputComponent) public emailInputSg!: EmailInputComponent;
  @ViewChild(PasswordInputComponent) public passwordInputSg!: PasswordInputComponent;

  public store = inject(LoginStore);

  public emailSg = signal<string>('');
  public passwordSg = signal<string>('');
  public loadingSg = this.store.loadingSg;

  private readonly errorSyncEffect = effect(() => {
    const storeError = this.store.errorSg();
    if (storeError && this.passwordInputSg) {
      this.passwordInputSg.errorSg.set(storeError);
    }
  });

  constructor() {
    // Clear store error when entering credentials step
    this.store.clearError();
  }

  public async submit(): Promise<void> {
    // Validate both inputs
    const emailValidation = this.emailInputSg.validate();
    const passwordValidation = this.passwordInputSg.validate();

    if (!emailValidation.isValid || !passwordValidation.isValid) {
      return;
    }

    const email = this.emailSg().trim();
    const password = this.passwordSg().trim();

    // Validate credentials and send OTP
    await this.store.validateCredentials(email, password);
  }

  public ngOnDestroy(): void {
    this.errorSyncEffect.destroy();
  }
}
