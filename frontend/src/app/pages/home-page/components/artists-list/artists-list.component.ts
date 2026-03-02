import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { AlbumService, type Album } from '@app/services/album.service';
import { ArtistService, type Artist } from '@app/services/artist.service';
import { GenreService, type Genre } from '@app/services/genre.service';
import { PlaybackService } from '@app/services/playback.service';
import {
  RecommendationService,
  type SongRecommendation,
} from '@app/services/recommendation.service';
import { SongService, type Song } from '@app/services/song.service';
import { AlbumCardComponent } from '@app/shared/components/album-card/album-card.component';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { MessageComponent } from '@app/shared/components/message';
import { SongTrailingMetaComponent } from '@app/shared/components/song-trailing-meta/song-trailing-meta';

@Component({
  selector: 'app-artists-list',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    CoverArtComponent,
    MessageComponent,
    SongTrailingMetaComponent,
    AlbumCardComponent,
  ],
  templateUrl: './artists-list.component.html',
  styleUrl: './artists-list.component.scss',
})
export class ArtistsListComponent {
  private readonly recommendationService = inject(RecommendationService);
  private readonly artistService = inject(ArtistService);
  private readonly albumService = inject(AlbumService);
  private readonly genreService = inject(GenreService);
  private readonly songService = inject(SongService);
  readonly playback = inject(PlaybackService);
  private readonly router = inject(Router);

  readonly isLoadingSg = signal(true);

  readonly subscriptionRecommendationsSg = signal<SongRecommendation[]>([]);
  readonly likeRecommendationsSg = signal<SongRecommendation[]>([]);
  readonly subscriptionErrorSg = signal<string>('');
  readonly likeErrorSg = signal<string>('');
  readonly recommendationPlayErrorSg = signal<string>('');
  private readonly recommendationLookupAlbumsSg = signal<Album[] | null>(null);

  readonly recommendedArtistsSg = signal<Artist[]>([]);
  readonly recommendedAlbumsSg = signal<Album[]>([]);
  readonly recommendedGenresSg = signal<Genre[]>([]);
  readonly recommendedSongsSg = signal<Song[]>([]);
  readonly artistsErrorSg = signal<string>('');
  readonly albumsErrorSg = signal<string>('');
  readonly genresErrorSg = signal<string>('');
  readonly songsErrorSg = signal<string>('');

  constructor() {
    this.loadCombinedData();
  }

  private loadCombinedData(): void {
    this.isLoadingSg.set(true);

    this.subscriptionErrorSg.set('');
    this.likeErrorSg.set('');
    this.recommendationPlayErrorSg.set('');
    this.artistsErrorSg.set('');
    this.albumsErrorSg.set('');
    this.genresErrorSg.set('');
    this.songsErrorSg.set('');

    forkJoin({
      subscription: this.recommendationService.getSubscriptionRecommendations().pipe(
        catchError(() => {
          this.subscriptionErrorSg.set(
            'Failed to load recommendation songs for your subscriptions.'
          );
          return of([]);
        })
      ),
      likes: this.recommendationService.getLikeBasedRecommendations().pipe(
        catchError(() => {
          this.likeErrorSg.set('Failed to load top liked recommendation songs.');
          return of([]);
        })
      ),
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
        this.subscriptionRecommendationsSg.set(result.subscription ?? []);
        this.likeRecommendationsSg.set(result.likes ?? []);

        this.recommendedArtistsSg.set(result.artists.items ?? []);
        this.recommendedAlbumsSg.set(result.albums.items ?? []);
        this.recommendedGenresSg.set(result.genres.items ?? []);
        this.recommendedSongsSg.set(result.songs.items ?? []);

        this.isLoadingSg.set(false);
      },
      error: () => {
        this.subscriptionErrorSg.set('Failed to load recommendation songs.');
        this.likeErrorSg.set('Failed to load recommendation songs.');
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
    this.router.navigate(['/genre', genre.id]);
  }

  artistNames(artists: string[] | Artist[] | undefined): string {
    if (!artists || artists.length === 0) {
      return 'Unknown artist';
    }

    if (typeof artists[0] === 'string') {
      return (artists as string[]).join(', ');
    }

    return (artists as Artist[]).map((artist) => artist.name).join(', ');
  }

  playSong(song: Song): void {
    const album = this.recommendedAlbumsSg().find((entry) =>
      entry.songs.some((item) => item.id === song.id)
    );

    if (!album) {
      return;
    }

    this.playback.playSingleSong(song, album);
  }

  playOrToggleSong(song: Song): void {
    const currentTrack = this.playback.currentTrackSg();
    if (currentTrack?.id === song.id) {
      this.playback.togglePlayPause();
      return;
    }

    this.playSong(song);
  }

  async playOrToggleRecommendationSong(song: SongRecommendation): Promise<void> {
    const currentTrack = this.playback.currentTrackSg();
    if (currentTrack?.id === song.song_id) {
      this.playback.togglePlayPause();
      return;
    }

    this.recommendationPlayErrorSg.set('');
    const resolved = await this.resolveRecommendedSong(song.song_id);
    if (!resolved) {
      this.recommendationPlayErrorSg.set(
        'Could not play this recommendation right now. Please try another song.'
      );
      return;
    }

    this.playback.playSingleSong(resolved.song, resolved.album);
  }

  toggleAlbumPlayback(album: Album, event: Event): void {
    event.preventDefault();
    event.stopPropagation();

    const currentTrack = this.playback.currentTrackSg();
    if (currentTrack?.albumId === album.id) {
      this.playback.togglePlayPause();
      return;
    }

    this.playback.playAlbum(album);
  }

  isAlbumPlaying(album: Album): boolean {
    const currentTrack = this.playback.currentTrackSg();
    return currentTrack?.albumId === album.id && this.playback.isPlayingSg();
  }

  private async resolveRecommendedSong(songId: string): Promise<{ song: Song; album: Album } | null> {
    const loadedMatch = this.findSongInAlbums(songId, this.recommendedAlbumsSg());
    if (loadedMatch) {
      return loadedMatch;
    }

    const cachedAlbums = this.recommendationLookupAlbumsSg();
    if (cachedAlbums) {
      return this.findSongInAlbums(songId, cachedAlbums);
    }

    try {
      const response = await firstValueFrom(this.albumService.getAlbums(1, 200));
      const albums = response.items ?? [];
      this.recommendationLookupAlbumsSg.set(albums);
      return this.findSongInAlbums(songId, albums);
    } catch {
      return null;
    }
  }

  private findSongInAlbums(songId: string, albums: Album[]): { song: Song; album: Album } | null {
    for (const album of albums) {
      const match = (album.songs ?? []).find((song) => song.id === songId);
      if (match) {
        return { song: match, album };
      }
    }

    return null;
  }
}
