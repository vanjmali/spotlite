import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-auth-form-footer',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './auth-form-footer.html',
  styleUrls: ['./auth-form-footer.scss'],
})
export class AuthFormFooter {
  @Input({ required: true }) label!: string;
  @Input() loadingLabel?: string;
  @Input() loading = false;
  @Input() stackActions = false;
  @Input() submit = false;
  @Input() secondaryLabel?: string;
  @Input() secondaryDisabled = false;
  @Input() hintText?: string;
  @Input() hintLinkText?: string;
  @Input() hintLinkUrl?: string;
  @Output() action = new EventEmitter<void>();
  @Output() secondaryAction = new EventEmitter<void>();
}
