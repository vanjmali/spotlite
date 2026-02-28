import { Component, inject, signal } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { PageComponent } from '@app/shared';
import { ConfigAsideComponent } from '@app/shared/components/page/components/config-aside';
import { CommonModule } from '@angular/common';
import { PlaybackService } from '@app/services/playback.service';

@Component({
  selector: 'app-admin-page',
  standalone: true,
  imports: [PageComponent, ConfigAsideComponent, CommonModule, RouterModule],
  templateUrl: './admin-page.component.html',
  styleUrl: './admin-page.component.scss',
})
export class AdminPage {
  readonly authService = inject(AuthService);
  private readonly router = inject(Router);
  private readonly playback = inject(PlaybackService);

  readonly isConfigAsideSg = signal(false);

  constructor() {
    this.playback.pause();
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/']);
  }
}
