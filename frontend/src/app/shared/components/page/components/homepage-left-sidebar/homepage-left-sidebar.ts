import { CommonModule } from '@angular/common';
import { Component, DestroyRef, computed, effect, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { AuthService } from '@app/services/auth.service';
import { SubscriptionService } from '@app/services/subscription.service';
import { UserActivity, UserActivityService } from '@app/services/user-activity.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';

type SidebarShortcut = { label: string; href: string };
type RecentActivity = { key: string; label: string; href: string; meta: string };

@Component({
  selector: 'app-homepage-left-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, CoverArtComponent],
  templateUrl: './homepage-left-sidebar.html',
  styleUrl: './homepage-left-sidebar.scss',
})
export class HomepageLeftSidebarComponent {
  private readonly authService = inject(AuthService);
  private readonly subscriptionService = inject(SubscriptionService);
  private readonly userActivityService = inject(UserActivityService);
  private readonly destroyRef = inject(DestroyRef);

  readonly followedArtistsSg = signal<SidebarShortcut[]>([]);
  readonly followedGenresSg = signal<SidebarShortcut[]>([]);
  private readonly userActivitiesSg = signal<UserActivity[]>([]);
  readonly recentActivitySg = computed<RecentActivity[]>(() => this.mapActivities(this.userActivitiesSg()));

  constructor() {
    this.subscriptionService.changes$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      if (this.authService.currentUserIdSg()) {
        this.loadSubscriptions();
        this.loadActivities();
      }
    });

    effect(() => {
      const userId = this.authService.currentUserIdSg();
      if (!userId) {
        this.followedArtistsSg.set([]);
        this.followedGenresSg.set([]);
        this.userActivitiesSg.set([]);
        return;
      }

      this.loadSubscriptions();
      this.loadActivities();
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

  private loadActivities(): void {
    this.userActivityService
      .getMyActivities(1, 12)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (response) => {
          this.userActivitiesSg.set(response.items ?? []);
        },
        error: () => {
          this.userActivitiesSg.set([]);
        },
      });
  }

  private mapActivities(activities: UserActivity[]): RecentActivity[] {
    return activities.slice(0, 6).map((activity) => {
      const entityType = (activity.entity_type ?? '').toUpperCase();
      const entityID = activity.entity_id ?? '';
      const label = activity.entity_name?.trim() || this.defaultLabel(activity);
      const href = this.entityHref(entityType, entityID);
      const meta = this.activityMeta(activity);
      const key = activity.event_id || `${activity.type}:${entityID}:${activity.occurred_at}`;

      return { key, label, href, meta };
    });
  }

  private defaultLabel(activity: UserActivity): string {
    switch (activity.type) {
      case 'LISTEN':
      case 'RATING_CREATED':
      case 'RATING_UPDATED':
        return 'Song';
      case 'SUBSCRIPTION_CREATED':
      case 'SUBSCRIPTION_DELETED':
        return activity.entity_type === 'ARTIST' ? 'Artist' : 'Genre';
      default:
        return 'Activity';
    }
  }

  private entityHref(entityType: string, entityID: string): string {
    if (!entityID) {
      return '/';
    }

    if (entityType === 'SONG') {
      return '/';
    }
    if (entityType === 'ARTIST') {
      return `/artist/${entityID}`;
    }
    if (entityType === 'GENRE') {
      return `/genre/${entityID}`;
    }

    return '/';
  }

  private activityMeta(activity: UserActivity): string {
    const when = this.relativeTime(activity.occurred_at || activity.created_at);

    switch (activity.type) {
      case 'LISTEN':
        return `Played ${when}`;
      case 'SUBSCRIPTION_CREATED':
        return `Subscribed ${when}`;
      case 'SUBSCRIPTION_DELETED':
        return `Unsubscribed ${when}`;
      case 'RATING_CREATED':
      case 'RATING_UPDATED':
        return `Rated ${activity.rating_value ?? 0}/5 ${when}`;
      default:
        return when;
    }
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
