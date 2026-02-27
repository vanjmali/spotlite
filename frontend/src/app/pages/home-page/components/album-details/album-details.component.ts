import { Component, computed, inject, signal, effect } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { AlbumService, type Album } from '../../../../services/album.service';
import { RouterLink } from '@angular/router';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { MessageComponent } from '@app/shared/components/message';
import { SongRatingBadgeComponent } from '@app/shared/components/song-rating-badge/song-rating-badge';

@Component({
  selector: 'app-album-details',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    WidgetComponent,
    RouterLink,
    CoverArtComponent,
    MessageComponent,
    SongRatingBadgeComponent,
  ],
  templateUrl: './album-details.component.html',
  styleUrl: './album-details.component.scss',
})
export class AlbumDetailsComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly albumService = inject(AlbumService);
  readonly playback = inject(PlaybackService);

  readonly albumSg = signal<Album | null>(null);
  readonly isLoadingSg = signal(false);
  readonly fetchErrorSg = signal<string>('');
  readonly activeSongIdSg = computed(() => this.playback.currentTrackSg()?.id ?? '');

  constructor() {
    effect(() => {
      const albumId = this.route.snapshot.paramMap.get('id');
      if (albumId) {
        this.loadAlbum(albumId);
      }
    });
  }

  private loadAlbum(albumId: string): void {
    this.isLoadingSg.set(true);
    this.fetchErrorSg.set('');
    this.albumService.getAlbumById(albumId).subscribe({
      next: (album) => {
        this.albumSg.set(album || null);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.fetchErrorSg.set('Failed to load album details. Please try again.');
        this.isLoadingSg.set(false);
      },
    });
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  playAlbum(): void {
    const album = this.albumSg();
    if (!album) {
      return;
    }

    this.playback.playAlbum(album);
  }

  playSong(index: number): void {
    const album = this.albumSg();
    if (!album || index < 0 || index >= (album.songs?.length ?? 0)) {
      return;
    }

    this.playback.playAlbum(album, album.songs[index].id);
  }

  toggleSongPlayback(songId: string, index: number, event: Event): void {
    event.stopPropagation();
    const isActiveSong = songId === this.activeSongIdSg();
    if (isActiveSong) {
      this.playback.togglePlayPause();
      return;
    }

    this.playSong(index);
  }
}
