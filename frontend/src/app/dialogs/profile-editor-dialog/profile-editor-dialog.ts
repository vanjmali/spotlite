import {
  NAME_PATTERN,
  USERNAME_PATTERN,
  VALIDATION_MESSAGES,
  applyFieldErrors,
  DialogComponent,
  LoaderComponent,
  MessageComponent,
  TextInputComponent,
} from '@app/shared';
import {
  Component,
  ElementRef,
  HostListener,
  ViewChild,
  computed,
  inject,
  output,
  signal,
  viewChild,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { UserProfileService, UserProfile } from '@app/services/user-profile.service';
import { focusFirstFocusable } from '@app/shared/utils/focus';

@Component({
  selector: 'app-profile-editor-dialog',
  standalone: true,
  imports: [CommonModule, DialogComponent, TextInputComponent, MessageComponent, LoaderComponent],
  templateUrl: './profile-editor-dialog.html',
  styleUrl: './profile-editor-dialog.scss',
})
export class ProfileEditorDialogComponent {
  private readonly userProfileService = inject(UserProfileService);

  readonly closed = output<void>();
  readonly saved = output<void>();

  readonly usernameSg = signal('');
  readonly firstNameSg = signal('');
  readonly lastNameSg = signal('');

  readonly isLoadingProfileSg = signal(false);
  readonly isSavingSg = signal(false);
  readonly errorSg = signal('');

  readonly usernameStatusSg = signal<'idle' | 'checking' | 'available' | 'taken' | 'error'>('idle');

  readonly usernamePattern = USERNAME_PATTERN;
  readonly namePattern = NAME_PATTERN;
  readonly usernamePatternMessage = VALIDATION_MESSAGES.USERNAME_INVALID;
  readonly firstNamePatternMessage = VALIDATION_MESSAGES.NAME_INVALID('First name');
  readonly lastNamePatternMessage = VALIDATION_MESSAGES.NAME_INVALID('Last name');

  readonly canSubmitSg = computed(() => {
    if (this.isLoadingProfileSg() || this.isSavingSg()) return false;
    if (this.usernameStatusSg() === 'taken' || this.usernameStatusSg() === 'checking') return false;

    return (
      this.usernameSg().trim().length >= 4 &&
      this.firstNameSg().trim().length >= 2 &&
      this.lastNameSg().trim().length >= 2
    );
  });

  readonly profileSg = signal<UserProfile | null>(null);

  readonly usernameInputSg = viewChild('usernameInput', { read: TextInputComponent });
  readonly firstNameInputSg = viewChild('firstNameInput', { read: TextInputComponent });
  readonly lastNameInputSg = viewChild('lastNameInput', { read: TextInputComponent });

  @ViewChild('dialogContent')
  private readonly dialogContentRef?: ElementRef<HTMLElement>;

  private usernameCheckTimer?: ReturnType<typeof setTimeout>;
  private usernameCheckRun = 0;

  constructor() {
    void this.loadProfile();
    setTimeout(() => focusFirstFocusable(this.dialogContentRef?.nativeElement ?? null), 0);
  }

  async loadProfile(): Promise<void> {
    this.isLoadingProfileSg.set(true);
    this.errorSg.set('');
    this.usernameStatusSg.set('idle');

    const result = await this.userProfileService.getProfile();
    this.isLoadingProfileSg.set(false);

    if (!result.success || !result.profile) {
      this.errorSg.set(result.error || 'Failed to load profile');
      return;
    }

    this.profileSg.set(result.profile);
    this.usernameSg.set(result.profile.username);
    this.firstNameSg.set(result.profile.first_name);
    this.lastNameSg.set(result.profile.last_name);
  }

  onUsernameChange(username: string): void {
    this.usernameSg.set(username);
    this.queueUsernameAvailabilityCheck(username);
  }

  onFirstNameChange(firstName: string): void {
    this.firstNameSg.set(firstName);
  }

  onLastNameChange(lastName: string): void {
    this.lastNameSg.set(lastName);
  }

  async submit(): Promise<void> {
    const usernameInput = this.usernameInputSg();
    const firstNameInput = this.firstNameInputSg();
    const lastNameInput = this.lastNameInputSg();
    if (!usernameInput || !firstNameInput || !lastNameInput) return;

    const usernameValidation = usernameInput.validate();
    const firstNameValidation = firstNameInput.validate();
    const lastNameValidation = lastNameInput.validate();
    if (
      !usernameValidation.isValid ||
      !firstNameValidation.isValid ||
      !lastNameValidation.isValid
    ) {
      return;
    }

    if (this.usernameStatusSg() === 'taken') {
      usernameInput.setExternalError(VALIDATION_MESSAGES.USERNAME_IN_USE);
      return;
    }

    this.isSavingSg.set(true);
    this.errorSg.set('');

    const result = await this.userProfileService.updateProfile({
      username: this.usernameSg().trim(),
      first_name: this.firstNameSg().trim(),
      last_name: this.lastNameSg().trim(),
    });

    this.isSavingSg.set(false);

    if (!result.success) {
      if (result.code === 'username_taken') {
        usernameInput.setExternalError(VALIDATION_MESSAGES.USERNAME_IN_USE);
      }

      if (result.fields) {
        applyFieldErrors(result.fields, {
          username: (message) => usernameInput.setExternalError(message),
          first_name: (message) => firstNameInput.setExternalError(message),
          last_name: (message) => lastNameInput.setExternalError(message),
        });
      }

      this.errorSg.set(result.error || 'Failed to update profile');
      return;
    }

    this.saved.emit();
    this.cancel();
  }

  cancel(): void {
    this.clearPendingUsernameCheck();
    this.errorSg.set('');
    this.usernameStatusSg.set('idle');
    this.closed.emit();
  }

  @HostListener('document:keydown', ['$event'])
  onDocumentKeydown(event: KeyboardEvent): void {
    if (event.key !== 'Escape') return;
    event.preventDefault();
    this.cancel();
  }

  private queueUsernameAvailabilityCheck(username: string): void {
    this.clearPendingUsernameCheck();

    const normalized = username.trim();
    const originalUsername = this.profileSg()?.username;
    if (!normalized || normalized === originalUsername) {
      this.usernameStatusSg.set('idle');
      return;
    }

    if (!this.usernamePattern.test(normalized)) {
      this.usernameStatusSg.set('idle');
      return;
    }

    this.usernameStatusSg.set('checking');

    const run = ++this.usernameCheckRun;
    this.usernameCheckTimer = setTimeout(async () => {
      const exists = await this.userProfileService.checkUsernameExists(normalized);
      if (run !== this.usernameCheckRun) {
        return;
      }

      if (exists === null) {
        this.usernameStatusSg.set('error');
        return;
      }

      this.usernameStatusSg.set(exists ? 'taken' : 'available');
    }, 350);
  }

  private clearPendingUsernameCheck(): void {
    this.usernameCheckRun++;
    if (this.usernameCheckTimer) {
      clearTimeout(this.usernameCheckTimer);
      this.usernameCheckTimer = undefined;
    }
  }
}
