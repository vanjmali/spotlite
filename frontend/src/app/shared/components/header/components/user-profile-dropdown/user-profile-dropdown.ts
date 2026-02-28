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
import { Notification, NotificationService } from '@app/services/notification.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { ProfileEditorDialogComponent } from '@app/dialogs/profile-editor-dialog/profile-editor-dialog';
import { ChangePasswordDialogComponent } from '@app/dialogs/change-password-dialog/change-password-dialog';

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

  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly hostRef = inject(ElementRef<HTMLElement>);
  private flashTimeout: ReturnType<typeof setTimeout> | null = null;

  readonly menuOpenSg = signal(false);
  readonly profileMenuOpenSg = signal(false);
  readonly profileDialogOpenSg = signal(false);
  readonly changePasswordDialogOpenSg = signal(false);
  readonly filterSg = signal<NotificationFilter>('all');
  readonly notificationsSg = signal<Notification[]>([]);
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
    if (!this.menuOpenSg() && !this.profileMenuOpenSg()) {
      return;
    }

    const target = event.target as Node | null;
    if (!target) {
      return;
    }

    if (!this.hostRef.nativeElement.contains(target)) {
      this.menuOpenSg.set(false);
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
    this.profileMenuOpenSg.set(false);
    this.menuOpenSg.set(!this.menuOpenSg());
    if (this.menuOpenSg()) {
      this.notificationService.markAllAsSeen();
      this.shouldFlashBadgeSg.set(false);
    }
  }

  toggleProfileMenu(): void {
    this.menuOpenSg.set(false);
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
}
