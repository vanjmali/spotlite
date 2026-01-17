import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { signal } from '@angular/core';
import { PageComponent } from '../../shared';
import { WidgetComponent } from '@app/shared/components/widget';
import { ChangePasswordDialogComponent } from '../../dialogs';

@Component({
  selector: 'app-profile-page',
  standalone: true,
  imports: [
    CommonModule,
    PageComponent,
    MatIconModule,
    WidgetComponent,
    ChangePasswordDialogComponent,
  ],
  templateUrl: './profile-page.html',
  styleUrls: ['./profile-page.scss'],
})
export class ProfilePage {
  private readonly router = inject(Router);
  readonly isChangePasswordDialogOpenSg = signal<boolean>(false);

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
}
