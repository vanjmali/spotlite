import { CommonModule } from '@angular/common';
import { Component, input, output } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-profile-success-callout',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './profile-success-callout.html',
  styleUrls: ['./profile-success-callout.scss'],
})
export class ProfileSuccessCalloutComponent {
  readonly messageSg = input.required<string>({ alias: 'message' });
  readonly closed = output<void>();
}
