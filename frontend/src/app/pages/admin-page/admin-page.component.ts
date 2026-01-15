import { Component, inject, signal } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { PageComponent } from '@app/shared';
import { ConfigAsideComponent } from '@app/shared/components/page/components/config-aside';
import { CommonModule } from '@angular/common';

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

  readonly isConfigAsideSg = signal(false);
  readonly selectedConfigOption = signal<string | null>(null);

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }

  onConfigOptionSelected(optionKey: string): void {
    this.selectedConfigOption.set(optionKey);
    // Navigate to the corresponding route
    this.router.navigate(['admin', optionKey]);
  }
}
