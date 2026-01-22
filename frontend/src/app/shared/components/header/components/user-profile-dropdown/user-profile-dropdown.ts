import { Component, computed, effect, inject, signal } from '@angular/core';
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

  readonly isDropdownOpenSg = signal(false);

  // Computed signal to get first letter of email
  readonly firstLetterSg = computed((email = this.authService.currentEmailSg()) => {
    return email ? email.charAt(0).toUpperCase() : '';
  });

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

  navigateToHome(): void {
    this.router.navigate(['/home']);
  }

  toggleDropdown(): void {
    this.isDropdownOpenSg.update((isOpen) => !isOpen);
  }

  closeDropdown(): void {
    this.isDropdownOpenSg.set(false);
  }

  navigateToProfile(): void {
    this.closeDropdown();
    this.router.navigate(['/profile']);
  }

  navigateToInbox(): void {
    this.closeDropdown();
    this.router.navigate(['/inbox']);
  }

  navigateToAdmin(): void {
    this.closeDropdown();
    this.router.navigate(['/admin']);
  }

  logout(): void {
    this.closeDropdown();
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
