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
import {
  DialogComponent,
  MessageComponent,
  TextInputComponent,
  TextareaInputComponent,
} from '../../shared';
import { ArtistService, Artist } from '../../services/artist.service';
import { SelectInputComponent, type SelectOption } from '@app/shared/components/input';
import { OptionsService } from '@app/shared/services/options.service';
import { runOnOpen } from '@app/shared/utils/dialog';
import { getHttpErrorMessage } from '@app/shared/utils/http-error';
import { focusFirstFocusable } from '@app/shared/utils/focus';

@Component({
  selector: 'app-artist-editor-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatIconModule,
    DialogComponent,
    MessageComponent,
    TextInputComponent,
    TextareaInputComponent,
    SelectInputComponent,
  ],
  templateUrl: './artist-editor-dialog.html',
  styleUrl: './artist-editor-dialog.scss',
})
export class ArtistEditorDialogComponent {
  private readonly artistService = inject(ArtistService);
  private readonly optionsService = inject(OptionsService);

  // Inputs
  readonly artist = input<Artist | null>(null);
  readonly isOpen = input<boolean>(false);

  // Outputs
  readonly closed = output<void>();
  readonly saved = output<void>();

  // State - Form fields as signals
  readonly nameSg = signal('');
  readonly descriptionSg = signal('');
  readonly genreIdsSg = signal<string[]>([]);
  readonly genreOptionsSg = signal<SelectOption[]>([]);

  // UI State
  readonly isLoadingSg = signal(false);
  @ViewChild('dialogContent')
  private readonly dialogContentRef?: ElementRef<HTMLElement>;
  readonly errorSg = signal<string>('');

  // Computed
  readonly dialogTitle = computed(() => (this.isEdit() ? 'Edit Artist' : 'Create Artist'));
  readonly submitButtonText = computed(() => (this.isEdit() ? 'Update' : 'Create'));
  readonly isEdit = computed(() => !!this.artist());
  readonly isFormValid = computed(() => {
    const name = this.nameSg().trim();
    const description = this.descriptionSg().trim();
    const genreIds = this.genreIdsSg();

    return name.length >= 2 && description.length >= 2 && genreIds.length >= 1;
  });

  constructor() {
    // Populate form when artist input changes
    runOnOpen(this.isOpen, () => {
      const artist = this.artist();
      if (artist) {
        this.nameSg.set(artist.name);
        this.descriptionSg.set(artist.description);
        this.genreIdsSg.set(artist.genres.map((genre) => genre.id));
      } else {
        this.nameSg.set('');
        this.descriptionSg.set('');
        this.genreIdsSg.set([]);
      }
      this.loadGenres();
      this.errorSg.set('');
      setTimeout(() => focusFirstFocusable(this.dialogContentRef?.nativeElement ?? null), 0);
    });
  }

  private loadGenres(): void {
    this.optionsService.loadGenres(200).subscribe({
      next: (options) => {
        this.genreOptionsSg.set(options);
      },
      error: (error) => {
        console.error('Failed to load genres:', error);
      },
    });
  }

  submit(): void {
    if (!this.isFormValid()) {
      this.errorSg.set('Please fill in all required fields correctly');
      return;
    }

    this.isLoadingSg.set(true);
    this.errorSg.set('');

    const payload = {
      name: this.nameSg().trim(),
      description: this.descriptionSg().trim(),
      genre_ids: this.genreIdsSg(),
    };

    if (this.isEdit()) {
      // Update existing artist
      this.artistService.updateArtist(this.artist()!.id, payload).subscribe({
        next: () => {
          this.isLoadingSg.set(false);
          this.optionsService.invalidateArtists();
          this.saved.emit();
          this.cancel();
        },
        error: (error) => {
          console.error('Failed to update artist:', error);
          this.errorSg.set(getHttpErrorMessage(error, 'Failed to update artist. Please try again.'));
          this.isLoadingSg.set(false);
        },
      });
    } else {
      // Create new artist
      this.artistService.createArtist(payload).subscribe({
        next: () => {
          this.isLoadingSg.set(false);
          this.optionsService.invalidateArtists();
          this.saved.emit();
          this.cancel();
        },
      error: (error) => {
        console.error('Failed to create artist:', error);
        this.errorSg.set(getHttpErrorMessage(error, 'Failed to create artist. Please try again.'));
        this.isLoadingSg.set(false);
      },
    });
    }
  }

  cancel(): void {
    this.nameSg.set('');
    this.descriptionSg.set('');
    this.genreIdsSg.set([]);
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
