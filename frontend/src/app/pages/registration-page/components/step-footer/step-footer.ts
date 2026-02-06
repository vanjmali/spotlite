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
  @Input() loading = false;
  @Output() action = new EventEmitter<void>();
}
