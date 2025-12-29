import { Component, input, model, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ValidationResult, EMAIL_PATTERN, VALIDATION_MESSAGES } from '@app/shared';
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
    this.valueSg.set(newValue.trim());
    // Clear error when user starts typing
    this.errorSg.set('');
  }

  public validate(): ValidationResult {
    const email = this.valueSg();

    if (this.requiredSg() && !email) {
      this.errorSg.set(VALIDATION_MESSAGES.EMAIL_REQUIRED);
      return { isValid: false, error: VALIDATION_MESSAGES.EMAIL_REQUIRED };
    }

    if (email && !EMAIL_PATTERN.test(email)) {
      this.errorSg.set(VALIDATION_MESSAGES.EMAIL_INVALID);
      return { isValid: false, error: VALIDATION_MESSAGES.EMAIL_INVALID };
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
