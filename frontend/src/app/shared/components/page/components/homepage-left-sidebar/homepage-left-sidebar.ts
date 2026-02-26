import { CommonModule } from '@angular/common';
import { Component, DestroyRef, computed, effect, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { AuthService } from '@app/services/auth.service';
import { PlaybackService } from '@app/services/playback.service';
import { SubscriptionService } from '@app/services/subscription.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';

type SidebarShortcut = { label: string; href: string };
type RecentActivity = { label: string; href: string; meta: string };

@Component({
  selector: 'app-homepage-left-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, CoverArtComponent],
  templateUrl: './homepage-left-sidebar.html',
  styleUrl: './homepage-left-sidebar.scss',
})
export class HomepageLeftSidebarComponent {
  private readonly authService = inject(AuthService);
  private readonly playback = inject(PlaybackService);
  private readonly subscriptionService = inject(SubscriptionService);
  private readonly destroyRef = inject(DestroyRef);

  readonly followedArtistsSg = signal<SidebarShortcut[]>([]);
  readonly followedGenresSg = signal<SidebarShortcut[]>([]);
  readonly recentActivitySg = computed<RecentActivity[]>(() =>
    this.playback
      .playHistorySg()
      .slice(0, 6)
      .map((entry) => ({
        label: entry.title,
        href: `/albums/${entry.albumId}`,
        meta: `Played ${this.relativeTime(entry.playedAt)}`,
      }))
  );

  constructor() {
    effect(() => {
      const userId = this.authService.currentUserIdSg();
      if (!userId) {
        this.followedArtistsSg.set([]);
        this.followedGenresSg.set([]);
        return;
      }

      this.loadSubscriptions();
    });
  }

  private loadSubscriptions(): void {
    this.subscriptionService
      .getMySubscriptions(1, 200)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (response) => {
          const all = response.items ?? [];
          this.followedArtistsSg.set(
            all
              .filter((entry) => entry.sub_type === 'ARTIST')
              .map((entry) => ({
                label: entry.entity_name || 'Unknown artist',
                href: `/artist/${entry.entity_id}`,
              }))
          );
          this.followedGenresSg.set(
            all
              .filter((entry) => entry.sub_type === 'GENRE')
              .map((entry) => ({
                label: entry.entity_name || 'Unknown genre',
                href: `/genre/${entry.entity_id}`,
              }))
          );
        },
        error: () => {
          this.followedArtistsSg.set([]);
          this.followedGenresSg.set([]);
        },
      });
  }

  private relativeTime(isoDate: string): string {
    const timestamp = new Date(isoDate).getTime();
    if (!Number.isFinite(timestamp)) {
      return 'recently';
    }

    const diffMs = Date.now() - timestamp;
    const diffMinutes = Math.floor(diffMs / 60000);
    if (diffMinutes <= 0) {
      return 'just now';
    }
    if (diffMinutes < 60) {
      return `${diffMinutes}m ago`;
    }
    const diffHours = Math.floor(diffMinutes / 60);
    if (diffHours < 24) {
      return `${diffHours}h ago`;
    }
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  }
}
