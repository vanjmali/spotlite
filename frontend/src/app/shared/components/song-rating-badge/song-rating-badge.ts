import { CommonModule } from '@angular/common';
import { Component, computed, inject, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { PlaybackService } from '@app/services/playback.service';

@Component({
  selector: 'app-song-rating-badge',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './song-rating-badge.html',
  styleUrl: './song-rating-badge.scss',
})
export class SongRatingBadgeComponent {
  readonly songId = input.required<string>();
  private readonly playback = inject(PlaybackService);

  readonly ratingValueSg = computed(() => {
    const id = this.songId();
    return this.playback.ratingsBySongSg()[id] ?? 0;
  });
}
