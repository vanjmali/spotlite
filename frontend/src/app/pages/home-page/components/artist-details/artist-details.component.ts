import { Component, inject, signal, effect } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { ArtistService, type Artist } from '../../../../services/artist.service';
import { AlbumService, type Album } from '../../../../services/album.service';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';

@Component({
  selector: 'app-artist-details',
  standalone: true,
  imports: [CommonModule, MatIconModule, WidgetComponent, CoverArtComponent],
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
    this.artistService.getArtistById(artistId).subscribe({
      next: (artist) => {
        this.artistSg.set(artist || null);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.isLoadingSg.set(false);
      },
    });
  }

  private loadAlbums(artistId: string): void {
    this.albumService.getAlbums(1, 100, { artist_id: artistId }).subscribe({
      next: (response) => {
        this.albumsSg.set(response.items || []);
      },
      error: () => {
        this.albumsSg.set([]);
      },
    });
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  onAlbumClick(album: Album): void {
    this.router.navigate(['/albums', album.id]);
  }

  playAlbum(album: Album): void {
    this.playback.playAlbum(album);
  }

  // goBack(): void {
  //   this.router.navigate(['/']);
  // }
}
