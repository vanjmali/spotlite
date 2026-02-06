import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-registration-check-email-step',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './check-email-step.html',
  styleUrls: ['./check-email-step.scss'],
})
export class RegistrationCheckEmailStep {}
