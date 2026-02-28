import {
  Component,
  DestroyRef,
  ElementRef,
  HostListener,
  OnDestroy,
  computed,
  inject,
  signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AuthService } from '@app/services/auth.service';
import { AlbumService } from '@app/services/album.service';
import { Notification, NotificationService } from '@app/services/notification.service';
import { UserActivity, UserActivityService } from '@app/services/user-activity.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { ProfileEditorDialogComponent } from '@app/dialogs/profile-editor-dialog/profile-editor-dialog';
import { ChangePasswordDialogComponent } from '@app/dialogs/change-password-dialog/change-password-dialog';
import { firstValueFrom } from 'rxjs';

type NotificationFilter = 'all' | 'artists' | 'albums';

@Component({
  selector: 'app-user-profile-dropdown',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    CoverArtComponent,
    ProfileEditorDialogComponent,
    ChangePasswordDialogComponent,
  ],
  templateUrl: './user-profile-dropdown.html',
  styleUrls: ['./user-profile-dropdown.scss'],
})
export class UserProfileDropdownComponent implements OnDestroy {
  readonly authService = inject(AuthService);
  readonly notificationService = inject(NotificationService);
  readonly userActivityService = inject(UserActivityService);
  readonly albumService = inject(AlbumService);

  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly hostRef = inject(ElementRef<HTMLElement>);
  private flashTimeout: ReturnType<typeof setTimeout> | null = null;

  readonly menuOpenSg = signal(false);
  readonly activityMenuOpenSg = signal(false);
  readonly profileMenuOpenSg = signal(false);
  readonly profileDialogOpenSg = signal(false);
  readonly changePasswordDialogOpenSg = signal(false);
  readonly filterSg = signal<NotificationFilter>('all');
  readonly notificationsSg = signal<Notification[]>([]);
  readonly activitiesSg = signal<UserActivity[]>([]);
  readonly activityLoadingSg = signal(false);
  readonly activityLoadingMoreSg = signal(false);
  readonly activityPageSg = signal(0);
  readonly activityTotalSg = signal(0);
  readonly dismissingIdsSg = signal<Set<string>>(new Set());
  readonly shouldFlashBadgeSg = signal(false);
  readonly unreadCountSg = computed(
    () => this.notificationsSg().filter((item) => item.is_new).length
  );
  readonly hasUnreadSg = computed(() => this.unreadCountSg() > 0);
  readonly filteredNotificationsSg = computed(() => {
    const filter = this.filterSg();
    if (filter === 'all') {
      return this.notificationsSg();
    }

    return this.notificationsSg().filter(
      (item) => this.mapTypeToFilter(item.notification_type) === filter
    );
  });
  readonly hasMoreActivitiesSg = computed(
    () => this.activitiesSg().length < this.activityTotalSg()
  );
  private readonly songAlbumCache = new Map<string, string>();

  constructor() {
    this.notificationService.notifications$
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((items) => this.notificationsSg.set(items ?? []));

    this.notificationService.incomingNotification$
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(() => {
        this.shouldFlashBadgeSg.set(true);
        if (this.flashTimeout) {
          clearTimeout(this.flashTimeout);
        }
        this.flashTimeout = setTimeout(() => this.shouldFlashBadgeSg.set(false), 1800);
      });
  }

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: MouseEvent): void {
    if (!this.menuOpenSg() && !this.activityMenuOpenSg() && !this.profileMenuOpenSg()) {
      return;
    }

    const target = event.target as Node | null;
    if (!target) {
      return;
    }

    if (!this.hostRef.nativeElement.contains(target)) {
      this.menuOpenSg.set(false);
      this.activityMenuOpenSg.set(false);
      this.profileMenuOpenSg.set(false);
    }
  }

  ngOnDestroy(): void {
    if (this.flashTimeout) {
      clearTimeout(this.flashTimeout);
      this.flashTimeout = null;
    }
  }

  toggleNotificationsMenu(): void {
    this.activityMenuOpenSg.set(false);
    this.profileMenuOpenSg.set(false);
    this.menuOpenSg.set(!this.menuOpenSg());
    if (this.menuOpenSg()) {
      this.notificationService.markAllAsSeen();
      this.shouldFlashBadgeSg.set(false);
    }
  }

  toggleActivityMenu(): void {
    this.menuOpenSg.set(false);
    this.profileMenuOpenSg.set(false);
    const nextOpen = !this.activityMenuOpenSg();
    this.activityMenuOpenSg.set(nextOpen);
    if (nextOpen) {
      this.resetActivities();
      this.loadActivitiesPage(1, false);
    }
  }

  toggleProfileMenu(): void {
    this.menuOpenSg.set(false);
    this.activityMenuOpenSg.set(false);
    this.profileMenuOpenSg.set(!this.profileMenuOpenSg());
  }

  setFilter(filter: NotificationFilter): void {
    this.filterSg.set(filter);
  }

  dismissNotification(id: string, event: Event): void {
    event.stopPropagation();
    this.scheduleDismiss(id);
  }

  clearAll(event: Event): void {
    event.stopPropagation();
    const ids = this.notificationsSg().map((item) => item.notification_id);
    for (const id of ids) {
      this.scheduleDismiss(id);
    }
  }

  formatTimeAgo(isoDate: string): string {
    const timestamp = new Date(isoDate).getTime();
    if (!Number.isFinite(timestamp)) {
      return '';
    }

    const diffSeconds = Math.max(0, Math.floor((Date.now() - timestamp) / 1000));
    if (diffSeconds < 60) {
      return 'Just now';
    }
    if (diffSeconds < 3600) {
      return `${Math.floor(diffSeconds / 60)}m ago`;
    }
    if (diffSeconds < 86400) {
      return `${Math.floor(diffSeconds / 3600)}h ago`;
    }
    return `${Math.floor(diffSeconds / 86400)}d ago`;
  }

  notificationTitle(item: Notification): string {
    const filter = this.mapTypeToFilter(item.notification_type);
    if (filter === 'artists') {
      return 'New artist activity';
    }

    return 'New album available';
  }

  openNotification(item: Notification): void {
    this.menuOpenSg.set(false);
    const mapped = this.mapTypeToFilter(item.notification_type);
    if (mapped === 'artists' && item.entity_id) {
      this.router.navigate(['/artist', item.entity_id]);
      return;
    }
    if (mapped === 'albums' && item.entity_id) {
      this.router.navigate(['/album', item.entity_id]);
      return;
    }
    this.router.navigate(['/']);
  }

  async openActivity(item: UserActivity): Promise<void> {
    this.activityMenuOpenSg.set(false);
    const entityType = (item.entity_type ?? '').toUpperCase();
    const entityId = item.entity_id ?? '';

    if (entityType === 'SONG' && entityId) {
      await this.openSongActivity(entityId);
      return;
    }

    if (entityType === 'ARTIST' && entityId) {
      this.router.navigate(['/artist', entityId]);
      return;
    }
    if (entityType === 'GENRE' && entityId) {
      this.router.navigate(['/genre', entityId]);
      return;
    }

    this.router.navigate(['/']);
  }

  onActivityListScroll(event: Event): void {
    const target = event.target as HTMLElement | null;
    if (!target) {
      return;
    }

    const nearBottom = target.scrollTop + target.clientHeight >= target.scrollHeight - 72;
    if (!nearBottom) {
      return;
    }

    if (
      !this.hasMoreActivitiesSg() ||
      this.activityLoadingSg() ||
      this.activityLoadingMoreSg()
    ) {
      return;
    }

    this.loadActivitiesPage(this.activityPageSg() + 1, true);
  }

  notificationFilters(): NotificationFilter[] {
    return ['all', 'albums', 'artists'];
  }

  isFilterActive(filter: NotificationFilter): boolean {
    return this.filterSg() === filter;
  }

  filterLabel(filter: NotificationFilter): string {
    switch (filter) {
      case 'all':
        return 'All';
      case 'artists':
        return 'Artists';
      case 'albums':
        return 'Albums';
    }
  }

  navigateToAdmin(): void {
    this.menuOpenSg.set(false);
    this.router.navigate(['/admin']);
  }

  isAdminActive(): boolean {
    return this.router.url.startsWith('/admin');
  }

  isProfileActive(): boolean {
    return this.profileMenuOpenSg();
  }

  profileCoverName(): string {
    return this.authService.profileAvatarNameSg();
  }

  trackNotification(_: number, item: Notification): string {
    return item.notification_id;
  }

  trackFilter(_: number, filter: NotificationFilter): string {
    return filter;
  }

  trackActivity(_: number, item: UserActivity): string {
    return item.event_id;
  }

  openProfileDialog(): void {
    this.profileMenuOpenSg.set(false);
    this.profileDialogOpenSg.set(true);
  }

  closeProfileDialog(): void {
    this.profileDialogOpenSg.set(false);
  }

  openChangePasswordDialog(): void {
    this.profileMenuOpenSg.set(false);
    this.changePasswordDialogOpenSg.set(true);
  }

  closeChangePasswordDialog(): void {
    this.changePasswordDialogOpenSg.set(false);
  }

  logout(): void {
    this.menuOpenSg.set(false);
    this.activityMenuOpenSg.set(false);
    this.profileMenuOpenSg.set(false);
    this.authService.logout();
    this.router.navigate(['/']);
  }

  isDismissing(id: string): boolean {
    return this.dismissingIdsSg().has(id);
  }

  private mapTypeToFilter(type: string): NotificationFilter {
    const normalized = (type || '').toLowerCase();
    if (normalized.includes('artist')) return 'artists';
    return 'albums';
  }

  private scheduleDismiss(id: string): void {
    if (this.dismissingIdsSg().has(id)) {
      return;
    }

    const next = new Set(this.dismissingIdsSg());
    next.add(id);
    this.dismissingIdsSg.set(next);

    setTimeout(() => {
      this.notificationService.dismissNotification(id);
      const updated = new Set(this.dismissingIdsSg());
      updated.delete(id);
      this.dismissingIdsSg.set(updated);
    }, 220);
  }

  activityTitle(item: UserActivity): string {
    const name = item.entity_name?.trim();
    if (name) {
      return name;
    }

    const entityType = (item.entity_type ?? '').toUpperCase();
    if (entityType === 'ARTIST') return 'Unknown artist';
    if (entityType === 'GENRE') return 'Unknown genre';
    return 'Unknown song';
  }

  activityDescription(item: UserActivity): string {
    if (item.type === 'SUBSCRIPTION_CREATED') {
      const entityType = (item.entity_type ?? '').toUpperCase();
      if (entityType === 'ARTIST') {
        return 'Subscribed to new artist';
      }
      return 'Subscribed to new genre';
    }

    if (item.type === 'SUBSCRIPTION_DELETED') {
      const entityType = (item.entity_type ?? '').toUpperCase();
      if (entityType === 'ARTIST') {
        return 'Unsubscribed from artist';
      }
      return 'Unsubscribed from genre';
    }

    if (item.type === 'LISTEN') {
      return 'Played song';
    }

    return 'Activity';
  }

  activityWhen(item: UserActivity): string {
    return this.formatTimeAgo(item.occurred_at || item.created_at);
  }

  private resetActivities(): void {
    this.activitiesSg.set([]);
    this.activityPageSg.set(0);
    this.activityTotalSg.set(0);
  }

  private loadActivitiesPage(page: number, append: boolean): void {
    if (page <= 1 && !append) {
      this.activityLoadingSg.set(true);
    } else {
      this.activityLoadingMoreSg.set(true);
    }

    this.userActivityService.getMyActivities(page, 15).subscribe({
      next: (response) => {
        const items = response.items ?? [];
        this.activityTotalSg.set(response.total ?? 0);
        this.activityPageSg.set(response.page ?? page);

        if (!append) {
          this.activitiesSg.set(items);
        } else {
          this.activitiesSg.set(this.mergeActivities(this.activitiesSg(), items));
        }

        this.activityLoadingSg.set(false);
        this.activityLoadingMoreSg.set(false);
      },
      error: () => {
        if (!append) {
          this.activitiesSg.set([]);
        }
        this.activityLoadingSg.set(false);
        this.activityLoadingMoreSg.set(false);
      },
    });
  }

  private mergeActivities(existing: UserActivity[], incoming: UserActivity[]): UserActivity[] {
    const byId = new Map<string, UserActivity>();
    for (const item of existing) {
      byId.set(item.event_id, item);
    }
    for (const item of incoming) {
      byId.set(item.event_id, item);
    }
    return Array.from(byId.values());
  }

  private async openSongActivity(songId: string): Promise<void> {
    const cachedAlbumId = this.songAlbumCache.get(songId);
    if (cachedAlbumId) {
      this.router.navigate(['/album', cachedAlbumId]);
      return;
    }

    const size = 100;
    let page = 1;

    try {
      while (true) {
        const response = await firstValueFrom(this.albumService.getAlbums(page, size));
        const albums = response.items ?? [];
        const matched = albums.find((album) => (album.songs ?? []).some((song) => song.id === songId));
        if (matched) {
          this.songAlbumCache.set(songId, matched.id);
          this.router.navigate(['/album', matched.id]);
          return;
        }

        const total = response.total ?? 0;
        const reachedEnd = page * size >= total || albums.length === 0;
        if (reachedEnd || page >= 20) {
          break;
        }

        page += 1;
      }
    } catch {
      // no-op fallback below
    }

    this.router.navigate(['/']);
  }
}
