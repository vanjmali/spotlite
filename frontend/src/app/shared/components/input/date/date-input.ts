import { Component, input, model, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { VALIDATION_MESSAGES, ValidationResult } from '@app/shared';
import { ErrorComponent } from '../error';

@Component({
  selector: 'app-date-input',
  standalone: true,
  imports: [CommonModule, FormsModule, ErrorComponent],
  templateUrl: './date-input.html',
  styleUrls: ['./date-input.scss'],
})
export class DateInputComponent {
  public readonly labelSg = input<string>('Date', { alias: 'label' });
  public readonly requiredSg = input<boolean>(false, { alias: 'required' });
  public readonly placeholderSg = input<string>('yyyy-mm-dd', { alias: 'placeholder' });
  public readonly minSg = input<string | null>(null, { alias: 'min' });
  public readonly maxSg = input<string | null>(null, { alias: 'max' });

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
    const value = this.valueSg();

    if (this.requiredSg() && !value) {
      const error = VALIDATION_MESSAGES.TEXT_REQUIRED(this.labelSg());
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    if (value) {
      const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
      if (!dateRegex.test(value)) {
        const error = `${this.labelSg()} must be in the format yyyy-mm-dd`;
        this.errorSg.set(error);
        return { isValid: false, error };
      }

      const date = new Date(value);
      if (isNaN(date.getTime())) {
        const error = `${this.labelSg()} is not a valid date`;
        this.errorSg.set(error);
        return { isValid: false, error };
      }

      const min = this.minSg();
      if (min && value < min) {
        const error = `${this.labelSg()} must be on or after ${min}`;
        this.errorSg.set(error);
        return { isValid: false, error };
      }

      const max = this.maxSg();
      if (max && value > max) {
        const error = `${this.labelSg()} must be on or before ${max}`;
        this.errorSg.set(error);
        return { isValid: false, error };
      }
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
