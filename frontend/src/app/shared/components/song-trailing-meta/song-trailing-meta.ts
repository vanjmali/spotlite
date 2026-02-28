import { CommonModule } from '@angular/common';
import { Component, input } from '@angular/core';
import { SongRatingBadgeComponent } from '@app/shared/components/song-rating-badge/song-rating-badge';
import { formatDuration } from '@app/shared/utils/display';

@Component({
  selector: 'app-song-trailing-meta',
  standalone: true,
  imports: [CommonModule, SongRatingBadgeComponent],
  templateUrl: './song-trailing-meta.html',
  styleUrl: './song-trailing-meta.scss',
})
export class SongTrailingMetaComponent {
  readonly songIdSg = input.required<string>({ alias: 'songId' });
  readonly averageSg = input<number | null | undefined>(undefined, { alias: 'average' });
  readonly countSg = input<number | null | undefined>(undefined, { alias: 'count' });
  readonly durationSecondsSg = input<number | null | undefined>(undefined, {
    alias: 'durationSeconds',
  });

  formatSongDuration(seconds: number | null | undefined): string {
    return formatDuration(seconds);
  }
}
