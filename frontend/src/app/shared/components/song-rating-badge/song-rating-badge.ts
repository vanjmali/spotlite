import { CommonModule } from '@angular/common';
import { Component, computed, effect, inject, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RatingService } from '@app/services/rating.service';

@Component({
  selector: 'app-song-rating-badge',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './song-rating-badge.html',
  styleUrl: './song-rating-badge.scss',
})
export class SongRatingBadgeComponent {
  private readonly ratingService = inject(RatingService);
  readonly songId = input<string | null | undefined>(null);
  readonly average = input<number | null | undefined>(null);
  readonly count = input<number | null | undefined>(null);
  readonly liveSummarySg = computed(() => {
    const songId = this.songId();
    if (!songId) {
      return null;
    }
    return this.ratingService.getLiveSummary(songId);
  });

  constructor() {
    effect(() => {
      const songId = this.songId();
      if (!songId) {
        return;
      }
      this.ratingService.refreshSongSummary(songId);
    });
  }

  readonly countSg = computed(() => {
    const value = this.liveSummarySg()?.count ?? this.count();
    return typeof value === 'number' && Number.isFinite(value) ? Math.max(0, Math.floor(value)) : 0;
  });

  readonly hasRatingsSg = computed(() => this.countSg() > 0);
  readonly averageTextSg = computed(() => {
    const value = this.liveSummarySg()?.avg ?? this.average();
    if (typeof value !== 'number' || !Number.isFinite(value)) {
      return '-';
    }
    return value.toFixed(1);
  });
}
