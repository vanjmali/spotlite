import { Component, effect, inject, signal, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { LoginStore } from '../../store';
import { EmailInputComponent } from '@app/shared/input';

@Component({
  selector: 'app-email-step',
  standalone: true,
  imports: [CommonModule, EmailInputComponent],
  templateUrl: './email-step.html',
  styleUrls: ['./email-step.scss'],
})
export class EmailStep {
  @ViewChild(EmailInputComponent) public emailInputSg!: EmailInputComponent;

  public store = inject(LoginStore);
  public emailSg = signal('');

  constructor() {
    // Clear store error when entering email step
    this.store.clearError();

    // Track store errors and sync to email input
    effect(() => {
      const storeError = this.store.errorSg();
      if (storeError && this.emailInputSg) {
        this.emailInputSg.errorSg.set(storeError);
      }
    });
  }

  public submit(): void {
    // Validate email input
    const validation = this.emailInputSg.validate();
    if (!validation.isValid) {
      return;
    }

    // Email is valid, proceed to next step
    const email = this.emailSg().trim();
    void this.store.checkEmailAndSendOtp(email);
  }
}
