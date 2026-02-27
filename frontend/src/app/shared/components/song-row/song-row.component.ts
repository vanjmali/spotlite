import { CommonModule } from '@angular/common';
import { Component, computed, input, output } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink } from '@angular/router';
import { Artist } from '@app/services/artist.service';
import { Genre } from '@app/services/genre.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { SongTrailingMetaComponent } from '@app/shared/components/song-trailing-meta/song-trailing-meta';

export interface SongRowSong {
  id: string;
  title: string;
  artists?: Artist[];
  genres?: Genre[];
  lengthSeconds?: number | null;
  rating?: {
    average: number;
    count: number;
  };
}

@Component({
  selector: 'app-song-row',
  standalone: true,
  imports: [CommonModule, MatIconModule, RouterLink, CoverArtComponent, SongTrailingMetaComponent],
  templateUrl: './song-row.component.html',
  styleUrl: './song-row.component.scss',
})
export class SongRowComponent {
  readonly songSg = input.required<SongRowSong>({ alias: 'song' });
  readonly compactSg = input<boolean>(false, { alias: 'compact' });
  readonly activeSg = input<boolean>(false, { alias: 'active' });
  readonly isPlayingSg = input<boolean>(false, { alias: 'isPlaying' });
  readonly showIndexSg = input<boolean>(false, { alias: 'showIndex' });
  readonly showToggleWhenActiveSg = input<boolean>(false, { alias: 'showToggleWhenActive' });
  readonly showGenresSg = input<boolean>(false, { alias: 'showGenres' });
  readonly showArtistLinkSg = input<boolean>(false, { alias: 'showArtistLink' });
  readonly indexSg = input<number | null>(null, { alias: 'index' });
  readonly coverSizeSg = input<number>(42, { alias: 'coverSize' });
  readonly coverFontSizeSg = input<number>(14, { alias: 'coverFontSize' });

  readonly rowClick = output<void>();
  readonly toggleClick = output<Event>();

  readonly artistsTextSg = computed(() => {
    const artists = this.songSg().artists ?? [];
    if (artists.length === 0) {
      return 'Unknown Artist';
    }

    return artists.map((artist) => artist.name).join(', ');
  });

  onRowClick(): void {
    this.rowClick.emit();
  }

  onToggle(event: Event): void {
    event.stopPropagation();
    this.toggleClick.emit(event);
  }
}
