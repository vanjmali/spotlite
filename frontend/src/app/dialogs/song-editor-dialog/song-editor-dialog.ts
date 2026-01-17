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
import { SongService, Song, CreateSongDto } from '@app/services/song.service';
import { ArtistService } from '@app/services/artist.service';

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

  readonly song = input<Song | null>(null);
  readonly isOpen = input<boolean>(false);
  readonly saved = output<void>();
  readonly closed = output<void>();

  readonly isSavingSg = signal(false);
  readonly titleSg = signal('');
  readonly genreSg = signal('');
  readonly durationSg = signal('');
  readonly selectedArtistIdsSg = signal<string[]>([]);
  readonly artistOptionsSg = signal<SelectOption[]>([]);

  // Computed
  readonly dialogTitle = computed(() => 'Create Song');
  readonly submitButtonText = computed(() => 'Create');
  readonly isLoadingSg = signal(false);
  readonly isFormValid = computed(() => {
    const title = this.titleSg().trim();
    const genre = this.genreSg().trim();
    const duration = this.durationSg().trim();
    const artists = this.selectedArtistIdsSg();

    return (
      title.length > 0 &&
      genre.length > 0 &&
      duration.length > 0 &&
      artists.length > 0 &&
      !isNaN(parseInt(duration))
    );
  });

  constructor() {
    effect(() => {
      if (this.isOpen()) {
        this.loadArtists();
        if (this.song()) {
          // Edit mode - populate form
          const songData = this.song();
          if (songData) {
            this.titleSg.set(songData.title);
            this.genreSg.set(songData.genre);
            this.durationSg.set(songData.lengthSeconds.toString());
            // Set artists array
            const artistIds = songData.artists.map((a) => a.id);
            this.selectedArtistIdsSg.set(artistIds);
          }
        } else {
          // Create mode - clear form
          this.titleSg.set('');
          this.genreSg.set('');
          this.durationSg.set('');
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

  onSave(): void {
    this.isLoadingSg.set(true);
    const songDto: CreateSongDto = {
      title: this.titleSg(),
      genre: this.genreSg(),
      length_seconds: parseInt(this.durationSg()),
      artist_ids: this.selectedArtistIdsSg(),
    };

    this.songService.createSong(songDto).subscribe({
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
