import { Component, input, model, signal, effect, ElementRef, ViewChild, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { VALIDATION_MESSAGES, ValidationResult } from '@app/shared';
import { ErrorComponent } from '../error';

export interface SelectOption {
  label: string;
  value: string;
}

@Component({
  selector: 'app-select-input',
  standalone: true,
  imports: [CommonModule, FormsModule, ErrorComponent, MatIconModule],
  templateUrl: './select-input.html',
  styleUrls: ['./select-input.scss'],
})
export class SelectInputComponent {
  @ViewChild('container', { static: true })
  private readonly containerRef?: ElementRef<HTMLElement>;

  @ViewChild('dropdown')
  private readonly dropdownRef?: ElementRef<HTMLElement>;

  public readonly labelSg = input<string>('Select', { alias: 'label' });
  public readonly requiredSg = input<boolean>(false, { alias: 'required' });
  public readonly placeholderSg = input<string>('Search...', { alias: 'placeholder' });
  public readonly optionsSg = input<SelectOption[]>([], { alias: 'options' });

  // Two-way binding for multiple selections
  public readonly valueSg = model<string[]>([], {
    alias: 'value',
  });

  // Internal state
  public readonly errorSg = signal<string>('');
  public readonly isOpenSg = signal<boolean>(false);
  public readonly searchQuerySg = signal<string>('');
  public readonly filteredOptionsSg = signal<SelectOption[]>([]);
  public readonly dropdownStyleSg = signal<Record<string, string>>({});

  constructor() {
    effect(() => {
      // Filter options based on search query
      const query = this.searchQuerySg().toLowerCase();
      const filtered = this.optionsSg().filter((option) =>
        option.label.toLowerCase().includes(query)
      );
      this.filteredOptionsSg.set(filtered);
    });

    effect(() => {
      if (!this.isOpenSg()) return;
      setTimeout(() => this.updateDropdownPosition(), 0);
    });
  }

  public toggleDropdown(): void {
    this.isOpenSg.set(!this.isOpenSg());
  }

  public openDropdown(): void {
    this.isOpenSg.set(true);
    setTimeout(() => this.updateDropdownPosition(), 0);
  }

  public closeDropdown(): void {
    this.isOpenSg.set(false);
    this.searchQuerySg.set('');
  }

  public onOptionToggle(optionValue: string): void {
    const currentValues = this.valueSg();
    let newValues: string[];

    if (currentValues.includes(optionValue)) {
      // Remove from selection
      newValues = currentValues.filter((v) => v !== optionValue);
    } else {
      // Add to selection
      newValues = [...currentValues, optionValue];
    }

    this.valueSg.set(newValues);
    this.errorSg.set('');
  }

  public removeSelection(optionValue: string, event: Event): void {
    event.stopPropagation();
    const newValues = this.valueSg().filter((v) => v !== optionValue);
    this.valueSg.set(newValues);
  }

  public isSelected(optionValue: string): boolean {
    return this.valueSg().includes(optionValue);
  }

  public getLabel(value: string): string {
    return this.optionsSg().find((o) => o.value === value)?.label ?? '';
  }

  @HostListener('window:resize')
  onWindowResize(): void {
    if (this.isOpenSg()) {
      this.updateDropdownPosition();
    }
  }

  @HostListener('window:scroll')
  onWindowScroll(): void {
    if (this.isOpenSg()) {
      this.updateDropdownPosition();
    }
  }

  private updateDropdownPosition(): void {
    const container = this.containerRef?.nativeElement;
    const dropdown = this.dropdownRef?.nativeElement;
    if (!container || !dropdown) return;

    const rect = container.getBoundingClientRect();
    const dropdownHeight = dropdown.offsetHeight || 0;
    const viewportHeight = window.innerHeight;
    const spaceBelow = viewportHeight - rect.bottom;
    const shouldOpenAbove = spaceBelow < dropdownHeight + 12;

    const top = shouldOpenAbove ? rect.top - dropdownHeight - 6 : rect.bottom + 6;
    const left = rect.left;
    const width = rect.width;

    this.dropdownStyleSg.set({
      top: `${Math.max(8, top)}px`,
      left: `${Math.max(8, left)}px`,
      width: `${Math.max(160, width)}px`,
    });
  }

  public validate(): ValidationResult {
    const values = this.valueSg();

    if (this.requiredSg() && values.length === 0) {
      const error = VALIDATION_MESSAGES.TEXT_REQUIRED(this.labelSg());
      this.errorSg.set(error);
      return { isValid: false, error };
    }

    this.errorSg.set('');
    return { isValid: true };
  }
}
