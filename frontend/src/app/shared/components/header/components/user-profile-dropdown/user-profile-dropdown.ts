import { Component, computed, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { AuthService } from '@app/services/auth.service';

@Component({
  selector: 'app-user-profile-dropdown',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './user-profile-dropdown.html',
  styleUrls: ['./user-profile-dropdown.scss'],
})
export class UserProfileDropdownComponent {
  readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  readonly isDropdownOpenSg = signal(false);

  // Computed signal to get first letter of email
  readonly firstLetterSg = computed((email = this.authService.currentEmailSg()) => {
    return email ? email.charAt(0).toUpperCase() : '';
  });

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
