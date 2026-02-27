import { CommonModule } from '@angular/common';
import { Component, computed, effect, inject, signal } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink, ActivatedRoute } from '@angular/router';
import { AlbumService, type Album } from '@app/services/album.service';
import { GenreService, type Genre } from '@app/services/genre.service';
import { PlaybackService } from '@app/services/playback.service';
import { SubscriptionService } from '@app/services/subscription.service';
import { AuthService } from '@app/services/auth.service';
import { MessageComponent } from '@app/shared/components/message';
import { SongRatingBadgeComponent } from '@app/shared/components/song-rating-badge/song-rating-badge';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { EMPTY, catchError, finalize, map, of } from 'rxjs';

type GenreSongRow = {
  songId: string;
  songTitle: string;
  artistsText: string;
  albumId: string;
  albumTitle: string;
};

@Component({
  selector: 'app-genre-details',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatIconModule,
    MessageComponent,
    SongRatingBadgeComponent,
    WidgetComponent,
  ],
  templateUrl: './genre-details.component.html',
  styleUrl: './genre-details.component.scss',
})
export class GenreDetailsComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly genreService = inject(GenreService);
  private readonly albumService = inject(AlbumService);
  private readonly authService = inject(AuthService);
  private readonly subscriptionService = inject(SubscriptionService);
  readonly playback = inject(PlaybackService);

  readonly genreSg = signal<Genre | null>(null);
  readonly isLoadingSg = signal(false);
  readonly errorSg = signal('');
  readonly albumsSg = signal<Album[]>([]);
  readonly isSubscribedSg = signal(false);
  readonly isSubscribeBusySg = signal(false);
  readonly subscribeErrorSg = signal('');
  readonly canManageSubscriptionSg = computed(() => this.authService.isAuthenticatedSg());
  readonly songsSg = computed<GenreSongRow[]>(() => {
    return this.albumsSg().flatMap((album) =>
      (album.songs ?? []).map((song) => ({
        songId: song.id,
        songTitle: song.title,
        artistsText:
          song.artists && song.artists.length > 0
            ? song.artists.map((artist) => artist.name).join(', ')
            : 'Unknown artist',
        albumId: album.id,
        albumTitle: album.title,
      }))
    );
  });

  readonly activeSongIdSg = computed(() => this.playback.currentTrackSg()?.id ?? '');

  constructor() {
    effect(() => {
      const genreId = this.route.snapshot.paramMap.get('id');
      if (!genreId) {
        return;
      }
      this.load(genreId);
      this.loadSubscriptionState(genreId);
    });

    effect(() => {
      const genreId = this.route.snapshot.paramMap.get('id');
      if (!genreId) {
        return;
      }

      if (!this.authService.isAuthenticatedSg()) {
        this.isSubscribedSg.set(false);
        return;
      }

      this.loadSubscriptionState(genreId);
    });
  }

  playFromRow(songId: string, albumId: string): void {
    const album = this.albumsSg().find((entry) => entry.id === albumId);
    if (!album) {
      return;
    }
    this.playback.playAlbum(album, songId);
  }

  toggleRowPlayback(songId: string, albumId: string, event: Event): void {
    event.stopPropagation();
    if (songId === this.activeSongIdSg()) {
      this.playback.togglePlayPause();
      return;
    }
    this.playFromRow(songId, albumId);
  }

  toggleGenreSubscription(): void {
    const genreId = this.route.snapshot.paramMap.get('id');
    if (!genreId || !this.canManageSubscriptionSg() || this.isSubscribeBusySg()) {
      return;
    }

    this.isSubscribeBusySg.set(true);
    this.subscribeErrorSg.set('');

    const request$ = this.isSubscribedSg()
      ? this.subscriptionService.unsubscribe(genreId).pipe(map(() => false))
      : this.subscriptionService.subscribe(genreId, 'GENRE').pipe(map(() => true));

    request$
      .pipe(
        catchError(() => {
          this.subscribeErrorSg.set('Could not update subscription right now.');
          return EMPTY;
        }),
        finalize(() => this.isSubscribeBusySg.set(false))
      )
      .subscribe((nextState) => {
        this.isSubscribedSg.set(nextState);
      });
  }

  private load(genreId: string): void {
    this.isLoadingSg.set(true);
    this.errorSg.set('');

    this.genreService.getGenreById(genreId).subscribe({
      next: (genre) => this.genreSg.set(genre),
      error: () => {
        this.genreSg.set(null);
        this.errorSg.set('Failed to load genre.');
        this.isLoadingSg.set(false);
      },
    });

    this.albumService.getAlbums(1, 100, { genre_id: genreId }).subscribe({
      next: (response) => {
        this.albumsSg.set(response.items ?? []);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.albumsSg.set([]);
        this.errorSg.set('Failed to load songs for this genre.');
        this.isLoadingSg.set(false);
      },
    });
  }

  private loadSubscriptionState(genreId: string): void {
    if (!this.canManageSubscriptionSg()) {
      this.isSubscribedSg.set(false);
      return;
    }

    this.subscriptionService
      .isSubscribed(genreId)
      .pipe(
        map(() => true),
        catchError(() => of(false))
      )
      .subscribe((value) => this.isSubscribedSg.set(value));
  }
}
