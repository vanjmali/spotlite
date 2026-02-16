import { Component, input, model, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { VALIDATION_MESSAGES, ValidationResult } from '@app/shared';
import { ErrorComponent } from '../error';

@Component({
  selector: 'app-text-input',
  standalone: true,
  imports: [CommonModule, FormsModule, ErrorComponent],
  templateUrl: './text-input.html',
  styleUrls: ['./text-input.scss'],
})
export class TextInputComponent {
  public readonly labelSg = input<string>('Text', { alias: 'label' });
  public readonly requiredSg = input<boolean>(false, { alias: 'required' });
  public readonly showRequiredIndicatorSg = input<boolean>(true, {
    alias: 'showRequiredIndicator',
  });
  public readonly placeholderSg = input<string>('', { alias: 'placeholder' });
  public readonly minSg = input<number | null>(null, { alias: 'min' });
  public readonly maxSg = input<number | null>(null, { alias: 'max' });
  public readonly patternSg = input<RegExp | null>(null, { alias: 'pattern' });
  public readonly patternMessageSg = input<string>('', { alias: 'patternMessage' });

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

  public setExternalError(message: string): void {
    this.errorSg.set(message);
  }

  public validate(): ValidationResult {
    const value = this.valueSg();
    const min = this.minSg();
    const max = this.maxSg();

    if (this.requiredSg() && !value) {
      const error = VALIDATION_MESSAGES.TEXT_REQUIRED(this.labelSg());
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    if (min !== null && value.length < min) {
      const error = `${this.labelSg()} must be at least ${min} characters`;
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    if (max !== null && value.length > max) {
      const error = `${this.labelSg()} must be at most ${max} characters`;
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    const pattern = this.patternSg();
    if (pattern && value && !pattern.test(value)) {
      const error = this.patternMessageSg() || `${this.labelSg()} is invalid`;
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
