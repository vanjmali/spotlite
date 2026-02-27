import { CommonModule } from '@angular/common';
import { Component, DestroyRef, computed, effect, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink } from '@angular/router';
import { AuthService } from '@app/services/auth.service';
import { PlaybackService } from '@app/services/playback.service';
import { RatingService } from '@app/services/rating.service';
import { SubscriptionService } from '@app/services/subscription.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { SongRatingBadgeComponent } from '@app/shared/components/song-rating-badge/song-rating-badge';
import { formatDuration } from '@app/shared/utils/display';

type SidebarStat = { label: string; value: string };

@Component({
  selector: 'app-homepage-right-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, CoverArtComponent, MatIconModule, SongRatingBadgeComponent],
  templateUrl: './homepage-right-sidebar.html',
  styleUrl: './homepage-right-sidebar.scss',
})
export class HomepageRightSidebarComponent {
  private readonly authService = inject(AuthService);
  readonly playback = inject(PlaybackService);
  private readonly subscriptionService = inject(SubscriptionService);
  private readonly ratingService = inject(RatingService);
  private readonly destroyRef = inject(DestroyRef);

  readonly currentAlbumSg = computed(() => this.playback.currentAlbumSg());
  readonly queueSg = computed(() => this.playback.queueSg());
  readonly activeSongIdSg = computed(() => this.playback.currentTrackSg()?.id ?? '');
  readonly primaryAlbumArtistSg = computed(
    () => this.playback.currentAlbumSg()?.artists?.[0] ?? null
  );
  readonly subscribedArtistsCountSg = signal(0);
  readonly subscribedGenresCountSg = signal(0);
  readonly avgRatingValueSg = signal<number | null>(null);
  readonly statsSg = computed<SidebarStat[]>(() => [
    { label: 'Subscribed artists', value: String(this.subscribedArtistsCountSg()) },
    { label: 'Subscribed genres', value: String(this.subscribedGenresCountSg()) },
    { label: 'Songs played this session', value: String(this.playback.playHistorySg().length) },
    {
      label: 'Avg rating given',
      value: this.avgRatingValueSg() === null ? '-' : this.avgRatingValueSg()!.toFixed(1),
    },
  ]);

  constructor() {
    this.subscriptionService.changes$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      const userId = this.authService.currentUserIdSg();
      if (userId) {
        this.loadStats(userId);
      }
    });

    effect(() => {
      const userId = this.authService.currentUserIdSg();
      if (!userId) {
        this.subscribedArtistsCountSg.set(0);
        this.subscribedGenresCountSg.set(0);
        this.avgRatingValueSg.set(null);
        return;
      }

      this.loadStats(userId);
    });
  }

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

  private loadStats(userId: string): void {
    this.subscriptionService
      .getMySubscriptions(1, 200)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (response) => {
          const items = response.items ?? [];
          this.subscribedArtistsCountSg.set(
            items.filter((entry) => entry.sub_type === 'ARTIST').length
          );
          this.subscribedGenresCountSg.set(
            items.filter((entry) => entry.sub_type === 'GENRE').length
          );
        },
        error: () => {
          this.subscribedArtistsCountSg.set(0);
          this.subscribedGenresCountSg.set(0);
        },
      });

    this.ratingService
      .getRatingsByUser(userId, 1, 200)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (response) => {
          const items = response.items ?? [];
          if (items.length === 0) {
            this.avgRatingValueSg.set(null);
            return;
          }

          const total = items.reduce((sum, item) => sum + (item.value || 0), 0);
          this.avgRatingValueSg.set(total / items.length);
        },
        error: () => {
          this.avgRatingValueSg.set(null);
        },
      });
  }
}
