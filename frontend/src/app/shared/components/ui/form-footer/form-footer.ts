import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-form-footer',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './form-footer.html',
  styleUrls: ['./form-footer.scss'],
})
export class FormFooter {
  @Input({ required: true }) label!: string;
  @Input() inProgress = false;
  @Input() stackActions = false;
  @Input() submit = false;
  @Input() formId?: string;
  @Input() secondaryLabel?: string;
  @Input() secondaryDisabled = false;
  @Input() hintText?: string;
  @Input() hintLinkText?: string;
  @Input() hintLinkUrl?: string;
  @Output() action = new EventEmitter<void>();
  @Output() secondaryAction = new EventEmitter<void>();
}
