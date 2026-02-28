import { CommonModule } from '@angular/common';
import { Component, DestroyRef, effect, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { AuthService } from '@app/services/auth.service';
import { SubscriptionService } from '@app/services/subscription.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';

type SidebarShortcut = { label: string; href: string };

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
  private readonly destroyRef = inject(DestroyRef);

  readonly followedArtistsSg = signal<SidebarShortcut[]>([]);
  readonly followedGenresSg = signal<SidebarShortcut[]>([]);

  constructor() {
    this.subscriptionService.changes$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      if (this.authService.currentUserIdSg()) {
        this.loadSubscriptions();
      }
    });

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
}
