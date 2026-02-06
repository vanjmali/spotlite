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
import { ValidationResult, OTP_PATTERN, VALIDATION_MESSAGES } from '@app/shared';
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
    if (sanitized.length > 1) {
      const nextDigits = sanitized.slice(0, 6 - index).split('');
      this.digitsSg.update((digits) => {
        const next = [...digits];
        nextDigits.forEach((digit, offset) => {
          next[index + offset] = digit;
        });
        return next;
      });

      this.valueSg.set(this.digitsSg().join(''));
      this.errorSg.set('');

      const focusIndex = Math.min(index + nextDigits.length, 5);
      queueMicrotask(() => this.focusBox(focusIndex, false));
      return;
    }

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
    if (event.metaKey || event.ctrlKey || event.altKey) return;

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

      const isCurrentEmpty = !currentDigits[index];

      this.digitsSg.update((digits) => {
        const next = [...digits];
        if (isCurrentEmpty && index > 0) {
          next[index - 1] = '';
        } else {
          next[index] = '';
        }
        return next;
      });

      // Update the full value
      const fullValue = this.digitsSg().join('');
      this.valueSg.set(fullValue);

      // Clear error when user types
      this.errorSg.set('');

      // Move focus to previous box if current is empty and index > 0
      if (isCurrentEmpty && index > 0) {
        this.focusBox(index - 1, false);
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
    const code = this.valueSg();

    if (this.requiredSg() && !code) {
      this.errorSg.set(VALIDATION_MESSAGES.OTP_REQUIRED);
      return { isValid: false, error: VALIDATION_MESSAGES.OTP_REQUIRED };
    }

    if (code && !OTP_PATTERN.test(code)) {
      this.errorSg.set(VALIDATION_MESSAGES.OTP_INVALID);
      return { isValid: false, error: VALIDATION_MESSAGES.OTP_INVALID };
    }

    this.errorSg.set('');
    return { isValid: true };
  }

  public onBoxClick(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    input?.select();
  }

  public onPaste(index: number, event: ClipboardEvent): void {
    event.preventDefault();

    const text = event.clipboardData?.getData('text') ?? '';
    const digitsOnly = text.replace(/[^0-9]/g, '');
    if (!digitsOnly) return;

    const startIndex = typeof index === 'number' ? index : 0;
    const nextDigits = digitsOnly.slice(0, 6 - startIndex).split('');

    this.digitsSg.update((digits) => {
      const next = [...digits];
      nextDigits.forEach((digit, offset) => {
        next[startIndex + offset] = digit;
      });
      return next;
    });

    this.valueSg.set(this.digitsSg().join(''));
    this.errorSg.set('');

    const focusIndex = Math.min(startIndex + nextDigits.length, 5);
    queueMicrotask(() => this.focusBox(focusIndex, false));
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
