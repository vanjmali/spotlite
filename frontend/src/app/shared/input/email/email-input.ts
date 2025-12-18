import { Component, input, model, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ValidationResult } from '../../../pages/login-page/store';
import { EMAIL_PATTERN } from '../../validation';
import { ErrorComponent } from '../error';

@Component({
  selector: 'app-email-input',
  standalone: true,
  imports: [CommonModule, FormsModule, ErrorComponent],
  templateUrl: './email-input.html',
  styleUrls: ['./email-input.scss'],
})
export class EmailInputComponent {
  public readonly labelSg = input<string>('Email address', { alias: 'label' });
  public readonly requiredSg = input<boolean>(false, { alias: 'required' });

  // Two-way binding using model with aliases
  public readonly valueSg = model<string>('', {
    alias: 'value',
  });

  // Internal error state
  public readonly errorSg = signal<string>('');

  public onValueChange(newValue: string): void {
    this.valueSg.set(newValue);
    // Clear error when user starts typing
    this.errorSg.set('');
  }

  public validate(): ValidationResult {
    const email = this.valueSg().trim();

    if (this.requiredSg() && !email) {
      const error = 'Email address is required';
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    if (email && !EMAIL_PATTERN.test(email)) {
      const error = 'Please enter a valid email address';
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
