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
import { SongService, type Song } from '@app/services/song.service';

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
  private readonly songService = inject(SongService);

  readonly album = input<Album | null>(null);
  readonly isOpen = input<boolean>(false);
  readonly saved = output<void>();
  readonly closed = output<void>();

  readonly isSavingSg = signal(false);
  readonly titleSg = signal('');
  readonly releaseDateSg = signal('');
  readonly genresSg = signal<string[]>([]);
  readonly currentGenreSg = signal('');
  readonly selectedArtistIdsSg = signal<string[]>([]);
  readonly artistOptionsSg = signal<SelectOption[]>([]);
  readonly availableSongsSg = signal<Song[]>([]);
  readonly selectedSongIdsSg = signal<string[]>([]);
  readonly isLoadingSongsSg = signal(false);

  // Computed
  readonly isCreateMode = computed(() => !this.album());
  readonly dialogTitle = computed(() => (this.isCreateMode() ? 'Create Album' : 'Edit Album'));
  readonly submitButtonText = computed(() => (this.isCreateMode() ? 'Create' : 'Save'));
  readonly isLoadingSg = signal(false);
  readonly isSongsValid = computed(() => !this.isCreateMode() || this.selectedSongIdsSg().length > 0);
  readonly isFormValid = computed(() => {
    const title = this.titleSg().trim();
    const releaseDate = this.releaseDateSg().trim();
    const genres = this.genresSg();
    const artists = this.selectedArtistIdsSg();

    return (
      title.length > 0 &&
      releaseDate.length > 0 &&
      genres.length >= 1 &&
      genres.every((g) => g.length > 0) &&
      artists.length > 0 &&
      this.isSongsValid()
    );
  });

  constructor() {
    effect(() => {
      if (this.isOpen()) {
        this.loadArtists();
        if (this.album()) {
          // Edit mode - populate form
          const albumData = this.album();
          if (albumData) {
            this.titleSg.set(albumData.title);
            this.releaseDateSg.set(albumData.releaseDate);
            this.genresSg.set([...albumData.genres]);
            this.currentGenreSg.set('');
            // Set artists array
            const artistIds = albumData.artists.map((a) => a.id);
            this.selectedArtistIdsSg.set(artistIds);
            this.selectedSongIdsSg.set(albumData.songs.map((song) => song.id));
          }
        } else {
          // Create mode - clear form
          this.titleSg.set('');
          this.releaseDateSg.set('');
          this.genresSg.set([]);
          this.currentGenreSg.set('');
          this.selectedArtistIdsSg.set([]);
          this.selectedSongIdsSg.set([]);
          this.availableSongsSg.set([]);
        }
      }
    });

    // Load songs when selected artists change
    effect(() => {
      const selectedArtistIds = this.selectedArtistIdsSg();
      if (selectedArtistIds.length > 0) {
        this.loadSongsForArtists(selectedArtistIds);
      } else {
        this.availableSongsSg.set([]);
        this.selectedSongIdsSg.set([]);
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

  private loadSongsForArtists(artistIds: string[]): void {
    this.isLoadingSongsSg.set(true);
    const allSongs: Song[] = [];
    let loadedCount = 0;

    artistIds.forEach((artistId) => {
      this.songService.getSongs(1, 100, { artist_id: artistId }).subscribe({
        next: (response) => {
          allSongs.push(...response.items);
          loadedCount++;
          if (loadedCount === artistIds.length) {
            // Remove duplicates by song id
            const uniqueSongs = Array.from(
              new Map(allSongs.map((song) => [song.id, song])).values()
            );
            this.availableSongsSg.set(uniqueSongs);
            this.isLoadingSongsSg.set(false);
          }
        },
        error: (error) => {
          console.error('Failed to load songs for artist:', error);
          loadedCount++;
          if (loadedCount === artistIds.length) {
            this.isLoadingSongsSg.set(false);
          }
        },
      });
    });
  }

  toggleSongSelection(songId: string): void {
    this.selectedSongIdsSg.update((ids) => {
      if (ids.includes(songId)) {
        return ids.filter((id) => id !== songId);
      } else {
        return [...ids, songId];
      }
    });
  }

  isSongSelected(songId: string): boolean {
    return this.selectedSongIdsSg().includes(songId);
  }

  addGenre(): void {
    const genre = this.currentGenreSg().trim();
    if (genre && !this.genresSg().includes(genre)) {
      this.genresSg.update((genres) => [...genres, genre]);
      this.currentGenreSg.set('');
    }
  }

  removeGenre(genre: string): void {
    this.genresSg.update((genres) => genres.filter((g) => g !== genre));
  }

  onSave(): void {
    this.isLoadingSg.set(true);
    if (this.isCreateMode()) {
      const albumDto: CreateAlbumDto = {
        title: this.titleSg(),
        release_date: this.releaseDateSg(),
        genres: this.genresSg(),
        song_ids: this.selectedSongIdsSg(),
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
      title: this.titleSg(),
      release_date: this.releaseDateSg(),
      genres: this.genresSg(),
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
