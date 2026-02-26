import { Component, computed, inject, signal, effect } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { ArtistService, type Artist } from '../../../../services/artist.service';
import { AlbumService, type Album } from '../../../../services/album.service';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { MessageComponent } from '@app/shared/components/message';

@Component({
  selector: 'app-artist-details',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    WidgetComponent,
    CoverArtComponent,
    MessageComponent,
    RouterLink,
  ],
  templateUrl: './artist-details.component.html',
  styleUrl: './artist-details.component.scss',
})
export class ArtistDetailsComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly artistService = inject(ArtistService);
  private readonly albumService = inject(AlbumService);
  private readonly playback = inject(PlaybackService);

  readonly artistSg = signal<Artist | null>(null);
  readonly albumsSg = signal<Album[]>([]);
  readonly isLoadingSg = signal(false);
  readonly artistIdSg = signal<string>('');
  readonly artistErrorSg = signal<string>('');
  readonly albumsErrorSg = signal<string>('');
  readonly activeAlbumIdSg = computed(
    () => this.playback.currentAlbumSg()?.id ?? this.playback.currentTrackSg()?.albumId ?? ''
  );

  constructor() {
    effect(() => {
      const artistId = this.route.snapshot.paramMap.get('id');
      if (artistId) {
        this.artistIdSg.set(artistId);
        this.loadArtist(artistId);
        this.loadAlbums(artistId);
      }
    });
  }

  private loadArtist(artistId: string): void {
    this.isLoadingSg.set(true);
    this.artistErrorSg.set('');
    this.artistService.getArtistById(artistId).subscribe({
      next: (artist) => {
        this.artistSg.set(artist || null);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.artistErrorSg.set('Failed to load artist details. Please try again.');
        this.isLoadingSg.set(false);
      },
    });
  }

  private loadAlbums(artistId: string): void {
    this.albumsErrorSg.set('');
    this.albumService.getAlbums(1, 100, { artist_id: artistId }).subscribe({
      next: (response) => {
        this.albumsSg.set(response.items || []);
      },
      error: () => {
        this.albumsErrorSg.set('Failed to load artist albums.');
        this.albumsSg.set([]);
      },
    });
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  onAlbumClick(album: Album): void {
    this.router.navigate(['/album', album.id]);
  }

  playAlbum(album: Album): void {
    this.playback.playAlbum(album);
  }

  // goBack(): void {
  //   this.router.navigate(['/']);
  // }
}
