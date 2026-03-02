import { CommonModule } from '@angular/common';
import { Component, computed, inject } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink } from '@angular/router';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { SongRatingBadgeComponent } from '@app/shared/components/song-rating-badge/song-rating-badge';
import { formatDuration } from '@app/shared/utils/display';

@Component({
  selector: 'app-homepage-right-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, CoverArtComponent, MatIconModule, SongRatingBadgeComponent],
  templateUrl: './homepage-right-sidebar.html',
  styleUrl: './homepage-right-sidebar.scss',
})
export class HomepageRightSidebarComponent {
  readonly playback = inject(PlaybackService);

  readonly currentAlbumSg = computed(() => this.playback.currentAlbumSg());
  readonly queueSg = computed(() => this.playback.queueSg());
  readonly activeSongIdSg = computed(() => this.playback.currentTrackSg()?.id ?? '');
  readonly primaryAlbumArtistSg = computed(
    () => this.playback.currentAlbumSg()?.artists?.[0] ?? null
  );

  playQueueIndex(index: number): void {
    this.playback.playFromQueue(index);
  }

  onQueueRowClick(songId: string, index: number): void {
    if (songId === this.activeSongIdSg()) {
      this.playback.togglePlayPause();
      return;
    }
    this.playQueueIndex(index);
  }

  toggleQueueSongPlayback(songId: string, index: number, event: Event): void {
    event.stopPropagation();
    this.onQueueRowClick(songId, index);
  }

  formatSongDuration(seconds: number | null | undefined): string {
    return formatDuration(seconds);
  }
}
