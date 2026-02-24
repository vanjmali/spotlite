import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { AlbumService, type Album } from '@app/services/album.service';
import { ArtistService, type Artist } from '@app/services/artist.service';
import { GenreService, type Genre } from '@app/services/genre.service';
import { SongService, type Song } from '@app/services/song.service';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { MessageComponent } from '@app/shared/components/message';

@Component({
  selector: 'app-artists-list',
  standalone: true,
  imports: [CommonModule, RouterLink, CoverArtComponent, MessageComponent],
  templateUrl: './artists-list.component.html',
  styleUrl: './artists-list.component.scss',
})
export class ArtistsListComponent {
  private readonly artistService = inject(ArtistService);
  private readonly albumService = inject(AlbumService);
  private readonly genreService = inject(GenreService);
  private readonly songService = inject(SongService);
  private readonly playback = inject(PlaybackService);
  private readonly router = inject(Router);

  readonly isLoadingSg = signal(true);
  readonly recommendedArtistsSg = signal<Artist[]>([]);
  readonly recommendedAlbumsSg = signal<Album[]>([]);
  readonly recommendedGenresSg = signal<Genre[]>([]);
  readonly recommendedSongsSg = signal<Song[]>([]);
  readonly artistsErrorSg = signal<string>('');
  readonly albumsErrorSg = signal<string>('');
  readonly genresErrorSg = signal<string>('');
  readonly songsErrorSg = signal<string>('');

  constructor() {
    this.loadRecommendations();
  }

  private loadRecommendations(): void {
    this.isLoadingSg.set(true);
    this.artistsErrorSg.set('');
    this.albumsErrorSg.set('');
    this.genresErrorSg.set('');
    this.songsErrorSg.set('');

    forkJoin({
      artists: this.artistService.getArtists(1, 12).pipe(
        catchError(() => {
          this.artistsErrorSg.set('Failed to load artists.');
          return of({ items: [] });
        })
      ),
      albums: this.albumService.getAlbums(1, 12).pipe(
        catchError(() => {
          this.albumsErrorSg.set('Failed to load albums.');
          return of({ items: [] });
        })
      ),
      genres: this.genreService.getGenres(1, 12).pipe(
        catchError(() => {
          this.genresErrorSg.set('Failed to load genres.');
          return of({ items: [] });
        })
      ),
      songs: this.songService.getSongs(1, 12).pipe(
        catchError(() => {
          this.songsErrorSg.set('Failed to load songs.');
          return of({ items: [] });
        })
      ),
    }).subscribe({
      next: (result) => {
        this.recommendedArtistsSg.set(result.artists.items ?? []);
        this.recommendedAlbumsSg.set(result.albums.items ?? []);
        this.recommendedGenresSg.set(result.genres.items ?? []);
        this.recommendedSongsSg.set(result.songs.items ?? []);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.artistsErrorSg.set('Failed to load artists.');
        this.albumsErrorSg.set('Failed to load albums.');
        this.genresErrorSg.set('Failed to load genres.');
        this.songsErrorSg.set('Failed to load songs.');
        this.isLoadingSg.set(false);
      },
    });
  }

  shortDescription(value: string, maxLength: number = 150): string {
    const normalized = value.trim();
    if (normalized.length <= maxLength) {
      return normalized;
    }

    return `${normalized.slice(0, maxLength).trimEnd()}...`;
  }

  openGenre(genre: Genre): void {
    this.router.navigate(['/search'], { queryParams: { q: genre.name } });
  }

  artistNames(artists: Artist[] | undefined): string {
    if (!artists || artists.length === 0) {
      return 'Unknown artist';
    }

    return artists.map((artist) => artist.name).join(', ');
  }

  playSong(song: Song): void {
    const album = this.recommendedAlbumsSg().find((entry) =>
      entry.songs.some((it) => it.id === song.id)
    );
    if (!album) {
      return;
    }

    this.playback.playSingleSong(song, album);
  }

  selectArtist(artist: Artist): void {
    this.router.navigate(['/artist', artist.id]);
  }
}
