import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-auth-step-header',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './step-header.html',
  styleUrls: ['./step-header.scss'],
})
export class AuthStepHeader {
  @Input({ required: true }) title!: string;
  @Input() showBack = false;
  @Input() showStepCount = false;
  @Input() stepIndex: number | null = null;
  @Input() totalSteps: number | null = null;

  @Output() back = new EventEmitter<void>();
}
