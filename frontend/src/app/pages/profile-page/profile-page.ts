import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { signal } from '@angular/core';
import { PageComponent } from '../../shared';
import { WidgetComponent } from '@app/shared/components/widget';
import { ChangePasswordDialogComponent } from '../../dialogs';
import { ProfileSuccessCalloutComponent } from './components/profile-success-callout';
import { AuthService } from '@app/services/auth.service';

@Component({
  selector: 'app-profile-page',
  standalone: true,
  imports: [
    CommonModule,
    PageComponent,
    MatIconModule,
    WidgetComponent,
    ChangePasswordDialogComponent,
    ProfileSuccessCalloutComponent,
  ],
  templateUrl: './profile-page.html',
  styleUrls: ['./profile-page.scss'],
})
export class ProfilePage {
  private readonly router = inject(Router);
  private readonly authService = inject(AuthService);
  readonly isChangePasswordDialogOpenSg = signal<boolean>(false);
  readonly passwordUpdateSuccessSg = signal<string>('');

  navigateToEditProfile(): void {
    // TODO: Implement edit profile navigation when component is created
    console.log('Navigate to edit profile');
  }

  navigateToChangePassword(): void {
    this.isChangePasswordDialogOpenSg.set(true);
  }

  closeChangePasswordDialog(): void {
    this.isChangePasswordDialogOpenSg.set(false);
  }

  onPasswordSaved(): void {
    this.passwordUpdateSuccessSg.set('Password successfully updated.');
  }

  dismissPasswordSaved(): void {
    this.passwordUpdateSuccessSg.set('');
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/']);
  }
}
