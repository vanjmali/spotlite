import { Component, input, output, inject, signal, effect, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { DialogComponent } from '@app/shared/components/dialog';
import {
  TextInputComponent,
  SelectInputComponent,
  DateInputComponent,
  type SelectOption,
} from '@app/shared/components/input';
import { AlbumService, Album, CreateAlbumDto, UpdateAlbumDto } from '@app/services/album.service';
import { ArtistService } from '@app/services/artist.service';
import { GenreService } from '@app/services/genre.service';

@Component({
  selector: 'app-album-editor-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatIconModule,
    DialogComponent,
    TextInputComponent,
    DateInputComponent,
    SelectInputComponent,
  ],
  templateUrl: './album-editor-dialog.html',
  styleUrl: './album-editor-dialog.scss',
})
export class AlbumEditorDialogComponent {
  private readonly albumService = inject(AlbumService);
  private readonly artistService = inject(ArtistService);
  private readonly genreService = inject(GenreService);

  readonly album = input<Album | null>(null);
  readonly isOpen = input<boolean>(false);
  readonly saved = output<void>();
  readonly closed = output<void>();

  readonly isSavingSg = signal(false);
  readonly titleSg = signal('');
  readonly releaseDateSg = signal('');
  readonly genreIdsSg = signal<string[]>([]);
  readonly selectedArtistIdsSg = signal<string[]>([]);
  readonly artistOptionsSg = signal<SelectOption[]>([]);
  readonly genreOptionsSg = signal<SelectOption[]>([]);

  // Computed
  readonly isCreateMode = computed(() => !this.album());
  readonly dialogTitle = computed(() => (this.isCreateMode() ? 'Create Album' : 'Edit Album'));
  readonly submitButtonText = computed(() => (this.isCreateMode() ? 'Create' : 'Save'));
  readonly isLoadingSg = signal(false);
  readonly isFormValid = computed(() => {
    const title = this.titleSg().trim();
    const releaseDate = this.releaseDateSg().trim();
    const genreIds = this.genreIdsSg();
    const artists = this.selectedArtistIdsSg();

    return (
      title.length >= 2 &&
      releaseDate.length === 10 &&
      genreIds.length >= 1 &&
      artists.length > 0
    );
  });

  constructor() {
    effect(() => {
      if (this.isOpen()) {
        this.loadArtists();
        this.loadGenres();
        if (this.album()) {
          // Edit mode - populate form
          const albumData = this.album();
          if (albumData) {
            this.titleSg.set(albumData.title);
            this.releaseDateSg.set(albumData.releaseDate);
            this.genreIdsSg.set(albumData.genres.map((genre) => genre.id));
            // Set artists array
            const artistIds = albumData.artists.map((a) => a.id);
            this.selectedArtistIdsSg.set(artistIds);
          }
        } else {
          // Create mode - clear form
          this.titleSg.set('');
          this.releaseDateSg.set('');
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
    if (this.isCreateMode()) {
      const albumDto: CreateAlbumDto = {
        title: this.titleSg().trim(),
        release_date: this.releaseDateSg().trim(),
        genre_ids: this.genreIdsSg(),
        artist_ids: this.selectedArtistIdsSg(),
      };

      this.albumService.createAlbum(albumDto).subscribe({
        next: () => {
          this.isLoadingSg.set(false);
          this.saved.emit();
          this.cancel();
        },
        error: (error) => {
          console.error('Failed to save album:', error);
          this.isLoadingSg.set(false);
        },
      });
      return;
    }

    const album = this.album();
    if (!album) {
      this.isLoadingSg.set(false);
      return;
    }

    const updateDto: UpdateAlbumDto = {
      title: this.titleSg().trim(),
      release_date: this.releaseDateSg().trim(),
      genre_ids: this.genreIdsSg(),
      artist_ids: this.selectedArtistIdsSg(),
    };

    this.albumService.updateAlbum(album.id, updateDto).subscribe({
      next: () => {
        this.isLoadingSg.set(false);
        this.saved.emit();
        this.cancel();
      },
      error: (error) => {
        console.error('Failed to save album:', error);
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
