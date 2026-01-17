import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { UserProfileDropdownComponent } from './components/user-profile-dropdown/user-profile-dropdown';
import { AuthService } from '@app/services/auth.service';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [CommonModule, RouterLink, UserProfileDropdownComponent],
  templateUrl: './header.html',
  styleUrls: ['./header.scss'],
})
export class HeaderComponent {
  readonly authService = inject(AuthService);
}
