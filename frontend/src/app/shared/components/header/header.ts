import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '@app/services/auth.service';
import { UserProfileDropdownComponent } from './components';
import { Component, inject, input } from '@angular/core';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [CommonModule, RouterLink, UserProfileDropdownComponent],
  templateUrl: './header.html',
  styleUrls: ['./header.scss'],
})
export class HeaderComponent {
  private readonly router = inject(Router);
  readonly authService = inject(AuthService);
  readonly presentLogoSg = input<boolean>(false);

  navigateToHome(): void {
    this.router.navigate(['/']);
  }
}
