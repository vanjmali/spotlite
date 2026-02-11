import {
  Component,
  inject,
  input,
  output,
  signal,
  computed,
  ElementRef,
  ViewChild,
  HostListener,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { DialogComponent, MessageComponent, TextInputComponent } from '../../shared';
import { GenreService, Genre } from '@app/services/genre.service';
import { runOnOpen } from '@app/shared/utils/dialog';
import { getHttpErrorMessage } from '@app/shared/utils/http-error';
import { OptionsService } from '@app/shared/services/options.service';
import { focusFirstFocusable } from '@app/shared/utils/focus';

@Component({
  selector: 'app-genre-editor-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatIconModule,
    DialogComponent,
    MessageComponent,
    TextInputComponent,
  ],
  templateUrl: './genre-editor-dialog.html',
  styleUrl: './genre-editor-dialog.scss',
})
export class GenreEditorDialogComponent {
  private readonly genreService = inject(GenreService);
  private readonly optionsService = inject(OptionsService);

  readonly genre = input<Genre | null>(null);
  readonly isOpen = input<boolean>(false);

  readonly closed = output<void>();
  readonly saved = output<void>();

  readonly nameSg = signal('');
  readonly isLoadingSg = signal(false);
  @ViewChild('dialogContent')
  private readonly dialogContentRef?: ElementRef<HTMLElement>;
  readonly errorSg = signal('');

  readonly isEdit = computed(() => !!this.genre());
  readonly dialogTitle = computed(() => (this.isEdit() ? 'Edit Genre' : 'Create Genre'));
  readonly submitButtonText = computed(() => (this.isEdit() ? 'Update' : 'Create'));
  readonly isFormValid = computed(() => this.nameSg().trim().length >= 2);

  constructor() {
    runOnOpen(this.isOpen, () => {
      const genre = this.genre();
      if (genre) {
        this.nameSg.set(genre.name);
      } else {
        this.nameSg.set('');
      }
      this.errorSg.set('');
      setTimeout(() => focusFirstFocusable(this.dialogContentRef?.nativeElement ?? null), 0);
    });
  }

  submit(): void {
    if (!this.isFormValid()) {
      this.errorSg.set('Genre name must be at least 2 characters');
      return;
    }

    this.isLoadingSg.set(true);
    this.errorSg.set('');

    const payload = {
      name: this.nameSg().trim(),
    };

    if (this.isEdit()) {
      this.genreService.updateGenre(this.genre()!.id, payload).subscribe({
        next: () => {
          this.isLoadingSg.set(false);
          this.optionsService.invalidateGenres();
          this.saved.emit();
          this.cancel();
        },
        error: (error) => {
          console.error('Failed to update genre:', error);
          this.errorSg.set(getHttpErrorMessage(error, 'Failed to update genre. Please try again.'));
          this.isLoadingSg.set(false);
        },
      });
      return;
    }

    this.genreService.createGenre(payload).subscribe({
      next: () => {
        this.isLoadingSg.set(false);
        this.optionsService.invalidateGenres();
        this.saved.emit();
        this.cancel();
      },
      error: (error) => {
        console.error('Failed to create genre:', error);
        this.errorSg.set(getHttpErrorMessage(error, 'Failed to create genre. Please try again.'));
        this.isLoadingSg.set(false);
      },
    });
  }

  cancel(): void {
    this.nameSg.set('');
    this.errorSg.set('');
    this.closed.emit();
  }

  @HostListener('document:keydown', ['$event'])
  onDocumentKeydown(event: KeyboardEvent): void {
    if (event.key !== 'Escape' || !this.isOpen()) return;
    event.preventDefault();
    this.cancel();
  }
}
