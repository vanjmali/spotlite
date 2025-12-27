import { Component, inject, signal, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TextInputComponent, EmailInputComponent, MessageComponent } from '@app/shared';
import { RegistrationStore } from '../../store';

@Component({
  selector: 'app-registration-personal-info-step',
  standalone: true,
  imports: [CommonModule, TextInputComponent, EmailInputComponent, MessageComponent],
  templateUrl: './personal-info-step.html',
  styleUrls: ['./personal-info-step.scss'],
})
export class PersonalInfoStep {
  public firstNameInputSg = viewChild('firstNameInput', { read: TextInputComponent });
  public lastNameInputSg = viewChild('lastNameInput', { read: TextInputComponent });
  public emailInputSg = viewChild('emailInput', { read: EmailInputComponent });

  public store = inject(RegistrationStore);

  public firstNameSg = signal<string>('');
  public lastNameSg = signal<string>('');
  public emailSg = signal<string>('');

  public async submit(): Promise<void> {
    const firstNameInput = this.firstNameInputSg();
    const lastNameInput = this.lastNameInputSg();
    const emailInput = this.emailInputSg();

    if (!firstNameInput || !lastNameInput || !emailInput) return;

    // Validate all inputs
    const firstNameValidation = firstNameInput.validate();
    const lastNameValidation = lastNameInput.validate();
    const emailValidation = emailInput.validate();

    if (!firstNameValidation.isValid || !lastNameValidation.isValid || !emailValidation.isValid) {
      return;
    }

    const firstName = this.firstNameSg();
    const lastName = this.lastNameSg();
    const email = this.emailSg();

    this.store.savePersonalInfo(firstName, lastName, email);
  }
}
