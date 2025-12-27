import { Component, input, model, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import {
  ValidationResult,
  PASSWORD_PATTERNS,
  MIN_PASSWORD_LENGTH,
  VALIDATION_MESSAGES,
} from '@app/shared';
import { ErrorComponent } from '../error';

@Component({
  selector: 'app-password-input',
  standalone: true,
  imports: [CommonModule, FormsModule, MatIconModule, ErrorComponent],
  templateUrl: './password-input.html',
  styleUrls: ['./password-input.scss'],
})
export class PasswordInputComponent {
  public readonly labelSg = input<string>('Password', { alias: 'label' });
  public readonly requiredSg = input<boolean>(false, { alias: 'required' });
  public readonly showCriteriaSg = input<boolean>(false, { alias: 'showCriteria' });
  public readonly minSg = input<number | null>(null, { alias: 'min' });
  public readonly maxSg = input<number | null>(null, { alias: 'max' });

  // Two-way binding using model with aliases
  public readonly valueSg = model<string>('', {
    alias: 'value',
  });

  // Track password visibility
  public readonly showPasswordSg = signal(false);

  // Internal error state
  public readonly errorSg = signal<string>('');

  // Individual criteria signals
  public readonly hasUppercaseSg = computed(() => PASSWORD_PATTERNS.uppercase.test(this.valueSg()));
  public readonly hasLowercaseSg = computed(() => PASSWORD_PATTERNS.lowercase.test(this.valueSg()));
  public readonly hasDigitSg = computed(() => PASSWORD_PATTERNS.digit.test(this.valueSg()));
  public readonly hasSpecialSg = computed(() => PASSWORD_PATTERNS.special.test(this.valueSg()));
  public readonly hasMinLengthSg = computed(() => this.valueSg().length >= MIN_PASSWORD_LENGTH);

  // Check if all criteria are met
  public readonly allCriteriaMet = computed(
    () =>
      this.hasUppercaseSg() &&
      this.hasLowercaseSg() &&
      this.hasDigitSg() &&
      this.hasSpecialSg() &&
      this.hasMinLengthSg()
  );

  public onValueChange(newValue: string): void {
    this.valueSg.set(newValue);
    // Clear error when user starts typing
    this.errorSg.set('');
  }

  public togglePasswordVisibility(): void {
    this.showPasswordSg.update((show) => !show);
  }

  public getInputType(): string {
    return this.showPasswordSg() ? 'text' : 'password';
  }

  public validate(): ValidationResult {
    const password = this.valueSg();
    const min = this.minSg();
    const max = this.maxSg();

    if (this.requiredSg() && !password) {
      this.errorSg.set(VALIDATION_MESSAGES.PASSWORD_REQUIRED);
      return { isValid: false, error: VALIDATION_MESSAGES.PASSWORD_REQUIRED };
    }

    if (min !== null && password.length < min) {
      const error = `Password must be at least ${min} characters`;
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    if (max !== null && password.length > max) {
      const error = `Password must be at most ${max} characters`;
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    // If criteria are shown, validate all criteria
    if (this.showCriteriaSg()) {
      if (
        !this.hasUppercaseSg() ||
        !this.hasLowercaseSg() ||
        !this.hasDigitSg() ||
        !this.hasSpecialSg() ||
        !this.hasMinLengthSg()
      ) {
        this.errorSg.set(VALIDATION_MESSAGES.PASSWORD_CRITERIA);
        return { isValid: false, error: VALIDATION_MESSAGES.PASSWORD_CRITERIA };
      }
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
