import { Component, effect, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { AuthService } from '@app/services/auth.service';
import { NotificationService } from '@app/services/notification.service';

@Component({
  selector: 'app-user-profile-dropdown',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './user-profile-dropdown.html',
  styleUrls: ['./user-profile-dropdown.scss'],
})
export class UserProfileDropdownComponent {
  readonly authService = inject(AuthService);
  readonly notificationService = inject(NotificationService);

  private readonly router = inject(Router);

  constructor() {
    effect(() => {
      const isAuthenticated = this.authService.isAuthenticatedSg();
      const token = this.authService.accessTokenSg();

      if (isAuthenticated && token) {
        this.notificationService.openSSEConnection(token);
      } else {
        this.notificationService.closeConnection();
      }
    });
  }

  navigateToProfile(): void {
    this.router.navigate(['/profile']);
  }

  navigateToInbox(): void {
    this.router.navigate(['/inbox']);
  }

  navigateToAdmin(): void {
    this.router.navigate(['/admin']);
  }

  isInboxActive(): boolean {
    return this.router.url.startsWith('/inbox');
  }

  isAdminActive(): boolean {
    return this.router.url.startsWith('/admin');
  }

  isProfileActive(): boolean {
    return this.router.url.startsWith('/profile');
  }
}
