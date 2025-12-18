import { Component, input, model, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { ValidationResult } from '../../../pages/login-page/store';
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
  private readonly MIN_PASSWORD_LENGTH = 8;

  // Two-way binding using model with aliases
  public readonly valueSg = model<string>('', {
    alias: 'value',
  });

  // Track password visibility
  public readonly showPasswordSg = signal(false);

  // Internal error state
  public readonly errorSg = signal<string>('');

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
    const password = this.valueSg().trim();

    if (this.requiredSg() && !password) {
      const error = 'Password is required';
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    if (password && password.length < this.MIN_PASSWORD_LENGTH) {
      const error = `Password must be at least ${this.MIN_PASSWORD_LENGTH} characters`;
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
