import { Component, input, output, inject, signal, effect, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { DialogComponent } from '@app/shared/components/dialog';
import {
  TextInputComponent,
  SelectInputComponent,
  type SelectOption,
} from '@app/shared/components/input';
import { SongService, Song, CreateSongDto, UpdateSongDto } from '@app/services/song.service';
import { ArtistService } from '@app/services/artist.service';
import { GenreService } from '@app/services/genre.service';

@Component({
  selector: 'app-song-editor-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatIconModule,
    DialogComponent,
    TextInputComponent,
    SelectInputComponent,
  ],
  templateUrl: './song-editor-dialog.html',
  styleUrl: './song-editor-dialog.scss',
})
export class SongEditorDialogComponent {
  private readonly songService = inject(SongService);
  private readonly artistService = inject(ArtistService);
  private readonly genreService = inject(GenreService);

  readonly song = input<Song | null>(null);
  readonly isOpen = input<boolean>(false);
  readonly saved = output<void>();
  readonly closed = output<void>();

  readonly isSavingSg = signal(false);
  readonly titleSg = signal('');
  readonly genreIdsSg = signal<string[]>([]);
  readonly durationSg = signal('');
  readonly selectedArtistIdsSg = signal<string[]>([]);
  readonly artistOptionsSg = signal<SelectOption[]>([]);
  readonly genreOptionsSg = signal<SelectOption[]>([]);

  // Computed
  readonly isEditMode = computed(() => !!this.song());
  readonly dialogTitle = computed(() => (this.isEditMode() ? 'Edit Song' : 'Create Song'));
  readonly submitButtonText = computed(() => (this.isEditMode() ? 'Save' : 'Create'));
  readonly isLoadingSg = signal(false);
  readonly isFormValid = computed(() => {
    const title = this.titleSg().trim();
    const duration = this.durationSg().trim();
    const artists = this.selectedArtistIdsSg();
    const genreIds = this.genreIdsSg();
    const durationValue = parseInt(duration);

    return (
      title.length >= 2 &&
      duration.length > 0 &&
      genreIds.length > 0 &&
      artists.length > 0 &&
      !isNaN(durationValue) &&
      durationValue > 0
    );
  });

  constructor() {
    effect(() => {
      if (this.isOpen()) {
        this.loadArtists();
        this.loadGenres();
        if (this.song()) {
          // Edit mode - populate form
          const songData = this.song();
          if (songData) {
            this.titleSg.set(songData.title);
            this.durationSg.set(songData.lengthSeconds.toString());
            this.genreIdsSg.set(songData.genres.map((genre) => genre.id));
            // Set artists array
            const artistIds = songData.artists.map((a) => a.id);
            this.selectedArtistIdsSg.set(artistIds);
          }
        } else {
          // Create mode - clear form
          this.titleSg.set('');
          this.durationSg.set('');
          this.genreIdsSg.set([]);
          this.selectedArtistIdsSg.set([]);
        }
      }
    });
  }

  private loadArtists(): void {
    this.artistService.getArtists(1, 100).subscribe({
      next: (response) => {
        this.artistOptionsSg.set(
          response.items.map((artist) => ({
            label: artist.name,
            value: artist.id,
          }))
        );
      },
      error: (error) => {
        console.error('Failed to load artists:', error);
      },
    });
  }

  private loadGenres(): void {
    this.genreService.getGenres(1, 200).subscribe({
      next: (response) => {
        this.genreOptionsSg.set(
          response.items.map((genre) => ({
            label: genre.name,
            value: genre.id,
          }))
        );
      },
      error: (error) => {
        console.error('Failed to load genres:', error);
      },
    });
  }

  onSave(): void {
    this.isLoadingSg.set(true);
    const basePayload = {
      title: this.titleSg().trim(),
      genre_ids: this.genreIdsSg(),
      length_seconds: parseInt(this.durationSg()),
      artist_ids: this.selectedArtistIdsSg(),
    };

    if (this.isEditMode()) {
      const updateDto: UpdateSongDto = basePayload;
      this.songService.updateSong(this.song()!.id, updateDto).subscribe({
        next: () => {
          this.isLoadingSg.set(false);
          this.saved.emit();
          this.cancel();
        },
        error: (error) => {
          console.error('Failed to update song:', error);
          this.isLoadingSg.set(false);
        },
      });
      return;
    }

    const createDto: CreateSongDto = basePayload;
    this.songService.createSong(createDto).subscribe({
      next: () => {
        this.isLoadingSg.set(false);
        this.saved.emit();
        this.cancel();
      },
      error: (error) => {
        console.error('Failed to save song:', error);
        this.isLoadingSg.set(false);
      },
    });
  }

  cancel(): void {
    this.closed.emit();
  }

  submit(): void {
    this.onSave();
  }

  onClose(): void {
    this.cancel();
  }
}
