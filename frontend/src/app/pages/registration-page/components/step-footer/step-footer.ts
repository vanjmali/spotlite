import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-registration-step-footer',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './step-footer.html',
  styleUrls: ['./step-footer.scss'],
})
export class RegistrationStepFooter {
  @Input({ required: true }) label!: string;
  @Input() loadingLabel?: string;
  @Input() loading = false;
  @Input() stackActions = false;
  @Input() secondaryLabel?: string;
  @Input() secondaryDisabled = false;
  @Input() hintText?: string;
  @Input() hintLinkText?: string;
  @Input() hintLinkUrl?: string;
  @Output() action = new EventEmitter<void>();
  @Output() secondaryAction = new EventEmitter<void>();
}
