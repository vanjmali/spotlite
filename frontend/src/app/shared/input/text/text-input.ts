import { Component, input, model, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ValidationResult, VALIDATION_MESSAGES } from '../../validation';
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
  public readonly placeholderSg = input<string>('', { alias: 'placeholder' });

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
    const value = this.valueSg();

    if (this.requiredSg() && !value) {
      const error = VALIDATION_MESSAGES.TEXT_REQUIRED(this.labelSg());
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
