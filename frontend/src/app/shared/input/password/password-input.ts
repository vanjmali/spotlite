import { Component, input, model, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import {
  ValidationResult,
  PASSWORD_PATTERNS,
  MIN_PASSWORD_LENGTH,
  VALIDATION_MESSAGES,
} from '../../validation';
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

  // Two-way binding using model with aliases
  public readonly valueSg = model<string>('', {
    alias: 'value',
  });

  // Track password visibility
  public readonly showPasswordSg = signal(false);

  // Internal error state
  public readonly errorSg = signal<string>('');

  // Individual criteria signals
  public readonly hasLetterSg = computed(() => PASSWORD_PATTERNS.letter.test(this.valueSg()));
  public readonly hasNumberOrSpecialSg = computed(() =>
    PASSWORD_PATTERNS.numberOrSpecial.test(this.valueSg())
  );
  public readonly hasMinLengthSg = computed(() => this.valueSg().length >= MIN_PASSWORD_LENGTH);

  // Check if all criteria are met
  public readonly allCriteriaMet = computed(
    () => this.hasLetterSg() && this.hasNumberOrSpecialSg() && this.hasMinLengthSg()
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

    if (this.requiredSg() && !password) {
      this.errorSg.set(VALIDATION_MESSAGES.PASSWORD_REQUIRED);
      return { isValid: false, error: VALIDATION_MESSAGES.PASSWORD_REQUIRED };
    }

    // If criteria are shown, validate all criteria
    if (this.showCriteriaSg()) {
      if (!this.hasLetterSg() || !this.hasNumberOrSpecialSg() || !this.hasMinLengthSg()) {
        this.errorSg.set(VALIDATION_MESSAGES.PASSWORD_CRITERIA);
        return { isValid: false, error: VALIDATION_MESSAGES.PASSWORD_CRITERIA };
      }
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
