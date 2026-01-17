import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { Router, RouterLink, RouterModule, RouterOutlet } from '@angular/router';
import { HeaderComponent, PageComponent } from '@app/shared';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-home-page',
  standalone: true,
  imports: [HeaderComponent, PageComponent, RouterLink, RouterOutlet, CommonModule, RouterModule],
  templateUrl: './home-page.component.html',
  styleUrl: './home-page.component.scss',
})
export class HomePage {
  readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
