import { CommonModule } from '@angular/common';
import { Component, computed, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-song-rating-badge',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './song-rating-badge.html',
  styleUrl: './song-rating-badge.scss',
})
export class SongRatingBadgeComponent {
  readonly average = input<number | null | undefined>(null);
  readonly count = input<number | null | undefined>(null);

  readonly countSg = computed(() => {
    const value = this.count();
    return typeof value === 'number' && Number.isFinite(value) ? Math.max(0, Math.floor(value)) : 0;
  });

  readonly hasRatingsSg = computed(() => this.countSg() > 0);
  readonly averageTextSg = computed(() => {
    const value = this.average();
    if (typeof value !== 'number' || !Number.isFinite(value)) {
      return '-';
    }
    return value.toFixed(1);
  });
}
