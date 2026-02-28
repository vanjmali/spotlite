import { CommonModule } from '@angular/common';
import { Component, effect, inject, OnDestroy } from '@angular/core';
import { Router, RouterModule, RouterOutlet } from '@angular/router';
import { PageComponent } from '@app/shared';
import { AuthService } from '../../services/auth.service';
import { NotificationService } from '@app/services/notification.service';

@Component({
  selector: 'app-home-page',
  standalone: true,
  imports: [PageComponent, RouterOutlet, CommonModule, RouterModule],
  templateUrl: './home-page.component.html',
  styleUrl: './home-page.component.scss',
})
export class HomePage implements OnDestroy {
  readonly authService = inject(AuthService);
  private readonly router = inject(Router);
  private readonly notificationService = inject(NotificationService);

  constructor() {
    effect(() => {
      const isAuthenticated = this.authService.isAuthenticatedSg();
      const token = this.authService.accessTokenSg();

      if (!isAuthenticated || !token) {
        this.notificationService.closeConnection();
        return;
      }

      this.notificationService.getAllNotifications();
      this.notificationService.openSSEConnection(token);
    });
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/']);
  }

  ngOnDestroy(): void {
    this.notificationService.closeConnection();
  }
}
