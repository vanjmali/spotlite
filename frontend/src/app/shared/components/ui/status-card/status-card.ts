import { CommonModule } from '@angular/common';
import { Component, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-status-card',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './status-card.html',
  styleUrls: ['./status-card.scss'],
})
export class StatusCard {
  public readonly titleSg = input.required<string>({ alias: 'title' });
  public readonly messageSg = input.required<string>({ alias: 'message' });
  public readonly iconSg = input<string>('info', { alias: 'icon' });
  public readonly variantSg = input<'success' | 'error' | 'info'>('info', { alias: 'variant' });
}
