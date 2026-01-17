import { Component, inject, input, output, signal, effect, computed } from '@angular/core';
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
  ],
  templateUrl: './artist-editor-dialog.html',
  styleUrl: './artist-editor-dialog.scss',
})
export class ArtistEditorDialogComponent {
  private readonly artistService = inject(ArtistService);

  // Inputs
  readonly artist = input<Artist | null>(null);
  readonly isOpen = input<boolean>(false);

  // Outputs
  readonly closed = output<void>();
  readonly saved = output<void>();

  // State - Form fields as signals
  readonly nameSg = signal('');
  readonly descriptionSg = signal('');
  readonly genresSg = signal<string[]>([]);
  readonly currentGenreSg = signal('');

  // UI State
  readonly isLoadingSg = signal(false);
  readonly errorSg = signal<string>('');

  // Computed
  readonly dialogTitle = computed(() => (this.isEdit() ? 'Edit Artist' : 'Create Artist'));
  readonly submitButtonText = computed(() => (this.isEdit() ? 'Update' : 'Create'));
  readonly isEdit = computed(() => !!this.artist());
  readonly isFormValid = computed(() => {
    const name = this.nameSg().trim();
    const description = this.descriptionSg().trim();
    const genres = this.genresSg();

    return (
      name.length >= 2 &&
      description.length >= 2 &&
      genres.length >= 1 &&
      genres.every((g) => g.length >= 2 && g.length <= 30)
    );
  });

  constructor() {
    // Populate form when artist input changes
    effect(() => {
      const artist = this.artist();
      if (artist) {
        this.nameSg.set(artist.name);
        this.descriptionSg.set(artist.description);
        this.genresSg.set([...artist.genres]);
      } else {
        this.nameSg.set('');
        this.descriptionSg.set('');
        this.genresSg.set([]);
      }
      this.currentGenreSg.set('');
      this.errorSg.set('');
    });
  }

  onAddGenre(): void {
    const genre = this.currentGenreSg().trim();
    if (!genre) return;

    // Validate genre length (2-30 chars to match backend)
    if (genre.length < 2 || genre.length > 30) {
      this.errorSg.set('Genre must be between 2 and 30 characters');
      return;
    }

    // Check if genre already exists
    if (this.genresSg().includes(genre)) {
      this.errorSg.set('This genre is already added');
      return;
    }

    this.genresSg.update((genres) => [...genres, genre]);
    this.currentGenreSg.set('');
    this.errorSg.set('');
  }

  onRemoveGenre(index: number): void {
    this.genresSg.update((genres) => genres.filter((_, i) => i !== index));
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
      genres: this.genresSg(),
    };

    if (this.isEdit()) {
      // Update existing artist
      this.artistService.updateArtist(this.artist()!.id, payload).subscribe({
        next: () => {
          this.isLoadingSg.set(false);
          this.saved.emit();
          this.cancel();
        },
        error: (error) => {
          console.error('Failed to update artist:', error);
          this.errorSg.set('Failed to update artist. Please try again.');
          this.isLoadingSg.set(false);
        },
      });
    } else {
      // Create new artist
      this.artistService.createArtist(payload).subscribe({
        next: () => {
          this.isLoadingSg.set(false);
          this.saved.emit();
          this.cancel();
        },
        error: (error) => {
          console.error('Failed to create artist:', error);
          this.errorSg.set('Failed to create artist. Please try again.');
          this.isLoadingSg.set(false);
        },
      });
    }
  }

  cancel(): void {
    this.nameSg.set('');
    this.descriptionSg.set('');
    this.genresSg.set([]);
    this.currentGenreSg.set('');
    this.errorSg.set('');
    this.closed.emit();
  }
}
