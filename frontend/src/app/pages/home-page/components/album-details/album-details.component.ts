import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { AlbumService, type Album } from '../../../../services/album.service';
import { RouterLink } from '@angular/router';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { MessageComponent } from '@app/shared/components/message';
import { SongTrailingMetaComponent } from '@app/shared/components/song-trailing-meta/song-trailing-meta';
import { distinctUntilChanged, map } from 'rxjs';

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
    SongTrailingMetaComponent,
  ],
  templateUrl: './album-details.component.html',
  styleUrl: './album-details.component.scss',
})
export class AlbumDetailsComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly destroyRef = inject(DestroyRef);
  private readonly albumService = inject(AlbumService);
  readonly playback = inject(PlaybackService);

  readonly albumSg = signal<Album | null>(null);
  readonly isLoadingSg = signal(false);
  readonly fetchErrorSg = signal<string>('');
  readonly activeSongIdSg = computed(() => this.playback.currentTrackSg()?.id ?? '');

  constructor() {
    this.route.paramMap
      .pipe(
        map((params) => params.get('id') ?? ''),
        distinctUntilChanged(),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe((albumId) => {
        if (!albumId) {
          return;
        }
        this.loadAlbum(albumId);
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

  onSongRowClick(songId: string, index: number): void {
    const isActiveSong = songId === this.activeSongIdSg();
    if (isActiveSong) {
      this.playback.togglePlayPause();
      return;
    }

    this.playSong(index);
  }

  toggleSongPlayback(songId: string, index: number, event: Event): void {
    event.stopPropagation();
    this.onSongRowClick(songId, index);
  }

  toggleAlbumPlayback(): void {
    const album = this.albumSg();
    if (!album) {
      return;
    }

    const currentTrack = this.playback.currentTrackSg();
    const isCurrentAlbum = currentTrack?.albumId === album.id;
    if (isCurrentAlbum) {
      this.playback.togglePlayPause();
      return;
    }

    this.playback.playAlbum(album);
  }

  isCurrentAlbumPlaying(): boolean {
    const album = this.albumSg();
    if (!album) {
      return false;
    }

    const currentTrack = this.playback.currentTrackSg();
    return currentTrack?.albumId === album.id && this.playback.isPlayingSg();
  }
}
