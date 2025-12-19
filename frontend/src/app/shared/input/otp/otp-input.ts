import {
  Component,
  input,
  model,
  signal,
  AfterViewInit,
  ElementRef,
  QueryList,
  ViewChildren,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { ValidationResult } from '../../../pages/login-page/store';
import { ErrorComponent } from '../error';

@Component({
  selector: 'app-otp-input',
  standalone: true,
  imports: [CommonModule, ErrorComponent],
  templateUrl: './otp-input.html',
  styleUrls: ['./otp-input.scss'],
})
export class OtpInputComponent implements AfterViewInit {
  public readonly labelSg = input<string>('', { alias: 'label' });
  public readonly requiredSg = input<boolean>(false, { alias: 'required' });
  @ViewChildren('otpBox') private otpBoxes!: QueryList<ElementRef<HTMLInputElement>>;

  // Two-way binding using model with alias
  public readonly valueSg = model<string>('', {
    alias: 'value',
  });

  // Array of 6 digits for display
  public readonly digitsSg = signal<string[]>(['', '', '', '', '', '']);

  // Internal error state (exposed publicly for template)
  public readonly errorSg = signal<string>('');

  public ngAfterViewInit(): void {
    this.focusBox(0);
  }

  public onDigitChange(index: number, value: string): void {
    // Only allow numbers
    const sanitized = value.replace(/[^0-9]/g, '');
    this.digitsSg.update((digits) => {
      const next = [...digits];
      next[index] = sanitized.slice(0, 1);
      return next;
    });

    // Update the full value
    const fullValue = this.digitsSg().join('');
    this.valueSg.set(fullValue);

    // Clear error when user starts typing
    this.errorSg.set('');

    // Auto-focus next input if value entered
    if (sanitized && index < 5) {
      queueMicrotask(() => this.focusBox(index + 1, true));
    }
  }

  public onKeyDown(index: number, event: KeyboardEvent): void {
    // Prevent non-numeric characters from being entered
    if (
      !/[0-9]/.test(event.key) &&
      !['Backspace', 'ArrowLeft', 'ArrowRight', 'Tab', 'Enter', 'Delete'].includes(event.key)
    ) {
      event.preventDefault();
      return;
    }

    // Handle backspace
    if (event.key === 'Backspace') {
      event.preventDefault();

      const currentDigits = this.digitsSg();

      // Clear the current box
      this.digitsSg.update((digits) => {
        const next = [...digits];
        next[index] = '';
        return next;
      });

      // Update the full value
      const fullValue = this.digitsSg().join('');
      this.valueSg.set(fullValue);

      // Clear error when user types
      this.errorSg.set('');

      // Move focus to previous box if current is empty and index > 0
      if (index > 0 && !currentDigits[index]) {
        this.focusBox(index - 1, true);
      }
    }
    // Handle left arrow - move to previous box
    else if (event.key === 'ArrowLeft' && index > 0) {
      event.preventDefault();
      this.focusBox(index - 1, true);
    }
    // Handle right arrow - move to next box
    else if (event.key === 'ArrowRight' && index < 5) {
      event.preventDefault();
      this.focusBox(index + 1, true);
    }
  }

  public validate(): ValidationResult {
    const code = this.valueSg().trim();

    if (this.requiredSg() && !code) {
      const error = 'Please enter a 6-digit code';
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    if (code && code.length !== 6) {
      const error = 'Please enter a 6-digit code';
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    this.errorSg.set('');
    return { isValid: true };
  }

  public onBoxClick(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    input?.select();
  }

  private focusBox(index: number, select = false): void {
    const target = this.otpBoxes?.get(index)?.nativeElement;
    if (!target) return;
    target.focus();
    if (select) {
      target.select();
    }
  }
}
