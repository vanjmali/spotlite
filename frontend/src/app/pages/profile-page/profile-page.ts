import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { signal } from '@angular/core';
import { MessageComponent, PageComponent } from '../../shared';
import { WidgetComponent } from '@app/shared/components/widget';
import { ChangePasswordDialogComponent, ProfileEditorDialogComponent } from '../../dialogs';

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
  readonly profileUpdateSuccessSg = signal<string>('');
  readonly isChangePasswordDialogOpenSg = signal<boolean>(false);

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
}
