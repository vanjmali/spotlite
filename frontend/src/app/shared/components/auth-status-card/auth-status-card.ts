import { CommonModule } from '@angular/common';
import { Component, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-auth-status-card',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './auth-status-card.html',
  styleUrls: ['./auth-status-card.scss'],
})
export class AuthStatusCard {
  public readonly titleSg = input.required<string>({ alias: 'title' });
  public readonly messageSg = input.required<string>({ alias: 'message' });
  public readonly iconSg = input<string>('info', { alias: 'icon' });
  public readonly variantSg = input<'success' | 'error' | 'info'>('info', { alias: 'variant' });
}
