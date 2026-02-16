import { Component, inject, viewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  TextInputComponent,
  EmailInputComponent,
  MessageComponent,
  NAME_PATTERN,
  VALIDATION_MESSAGES,
  applyFieldErrors,
} from '@app/shared';
import { RegistrationStore } from '../../store';
import { FormFooter } from '@app/shared';

@Component({
  selector: 'app-registration-personal-info-step',
  standalone: true,
  imports: [CommonModule, TextInputComponent, EmailInputComponent, MessageComponent, FormFooter],
  templateUrl: './personal-info-step.html',
  styleUrls: ['./personal-info-step.scss'],
})
export class PersonalInfoStep {
  public firstNameInputSg = viewChild('firstNameInput', { read: TextInputComponent });
  public lastNameInputSg = viewChild('lastNameInput', { read: TextInputComponent });
  public emailInputSg = viewChild('emailInput', { read: EmailInputComponent });

  public store = inject(RegistrationStore);
  public firstNameSg = this.store.firstNameSg;
  public lastNameSg = this.store.lastNameSg;
  public emailSg = this.store.emailSg;
  public loadingSg = this.store.loadingSg;
  public readonly namePattern = NAME_PATTERN;
  public readonly namePatternMessage = VALIDATION_MESSAGES.NAME_INVALID;

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

    const result = await this.store.savePersonalInfo(firstName, lastName, email);
    if (!result.success) {
      if (result.error && result.errorField === 'email') {
        emailInput.setExternalError(result.error);
      }

      if (result.fields) {
        applyFieldErrors(result.fields, {
          email: (msg) => emailInput.setExternalError(msg),
          first_name: (msg) => firstNameInput.setExternalError(msg),
          last_name: (msg) => lastNameInput.setExternalError(msg),
        });
      }
    }
  }
}
