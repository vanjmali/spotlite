import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MessageComponent, PageComponent } from '../../shared';
import { WidgetComponent } from '@app/shared/components/widget';
import { ChangePasswordDialogComponent, ProfileEditorDialogComponent } from '../../dialogs';
import { AuthService } from '@app/services/auth.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-profile-page',
  standalone: true,
  imports: [
    CommonModule,
    PageComponent,
    MessageComponent,
    MatIconModule,
    WidgetComponent,
    ChangePasswordDialogComponent,
    ProfileEditorDialogComponent,
  ],
  templateUrl: './profile-page.html',
  styleUrls: ['./profile-page.scss'],
})
export class ProfilePage {
  private readonly router = inject(Router);
  private readonly authService = inject(AuthService);

  readonly profileUpdateSuccessSg = signal<string>('');
  readonly isChangePasswordDialogOpenSg = signal<boolean>(false);
  readonly passwordUpdateSuccessSg = signal<string>('');

  onProfileSaved(): void {
    this.profileUpdateSuccessSg.set('Profile successfully updated.');
    setTimeout(() => this.profileUpdateSuccessSg.set(''), 3000);
  }

  openChangePasswordDialog(): void {
    this.isChangePasswordDialogOpenSg.set(true);
  }

  closeChangePasswordDialog(): void {
    this.isChangePasswordDialogOpenSg.set(false);
  }

  onPasswordSaved(): void {
    this.passwordUpdateSuccessSg.set('Password successfully updated.');
    setTimeout(() => this.passwordUpdateSuccessSg.set(''), 3000);
  }

  logout(): void {
    this.authService.logout();
    this.router.navigate(['/']);
  }
}
