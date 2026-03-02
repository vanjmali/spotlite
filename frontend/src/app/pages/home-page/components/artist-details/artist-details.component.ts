import { Component, DestroyRef, computed, inject, signal, effect } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { ArtistService, type Artist } from '../../../../services/artist.service';
import { AlbumService, type Album } from '../../../../services/album.service';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { MessageComponent } from '@app/shared/components/message';
import { AuthService } from '@app/services/auth.service';
import { SubscriptionService } from '@app/services/subscription.service';
import { EMPTY, catchError, distinctUntilChanged, finalize, map, of } from 'rxjs';

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
  private readonly destroyRef = inject(DestroyRef);
  private readonly router = inject(Router);
  private readonly artistService = inject(ArtistService);
  private readonly albumService = inject(AlbumService);
  private readonly authService = inject(AuthService);
  private readonly subscriptionService = inject(SubscriptionService);
  readonly playback = inject(PlaybackService);

  readonly artistSg = signal<Artist | null>(null);
  readonly albumsSg = signal<Album[]>([]);
  readonly isLoadingSg = signal(false);
  readonly artistIdSg = signal<string>('');
  readonly artistErrorSg = signal<string>('');
  readonly albumsErrorSg = signal<string>('');
  readonly isSubscribedSg = signal(false);
  readonly isSubscribeBusySg = signal(false);
  readonly subscribeErrorSg = signal('');
  readonly canManageSubscriptionSg = computed(() => this.authService.isAuthenticatedSg());
  readonly activeAlbumIdSg = computed(
    () => this.playback.currentAlbumSg()?.id ?? this.playback.currentTrackSg()?.albumId ?? ''
  );

  constructor() {
    this.route.paramMap
      .pipe(
        map((params) => params.get('id') ?? ''),
        distinctUntilChanged(),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe((artistId) => {
        if (!artistId) {
          return;
        }

        this.artistIdSg.set(artistId);
        this.loadArtist(artistId);
        this.loadAlbums(artistId);
        this.loadSubscriptionState(artistId);
      });

    effect(() => {
      const artistId = this.artistIdSg();
      if (!artistId) {
        return;
      }

      if (!this.authService.isAuthenticatedSg()) {
        this.isSubscribedSg.set(false);
        return;
      }

      this.loadSubscriptionState(artistId);
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

  toggleArtistSubscription(): void {
    const artistId = this.artistIdSg();
    if (!artistId || !this.canManageSubscriptionSg() || this.isSubscribeBusySg()) {
      return;
    }

    this.isSubscribeBusySg.set(true);
    this.subscribeErrorSg.set('');
    const request$ = this.isSubscribedSg()
      ? this.subscriptionService.unsubscribe(artistId).pipe(map(() => false))
      : this.subscriptionService.subscribe(artistId, 'ARTIST').pipe(map(() => true));

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

  toggleAlbumPlayback(album: Album, event: Event): void {
    event.stopPropagation();
    if (album.id === this.activeAlbumIdSg()) {
      this.playback.togglePlayPause();
      return;
    }
    this.playAlbum(album);
  }

  private loadSubscriptionState(artistId: string): void {
    if (!this.canManageSubscriptionSg()) {
      this.isSubscribedSg.set(false);
      return;
    }

    this.subscriptionService
      .isSubscribed(artistId)
      .pipe(
        map(() => true),
        catchError(() => of(false))
      )
      .subscribe((value) => this.isSubscribedSg.set(value));
  }

  // goBack(): void {
  //   this.router.navigate(['/']);
  // }
}
