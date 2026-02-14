import {
  Component,
  input,
  output,
  inject,
  signal,
  computed,
  ElementRef,
  ViewChild,
  HostListener,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatIconModule } from '@angular/material/icon';
import { DialogComponent } from '@app/shared/components/dialog';
import {
  TextInputComponent,
  SelectInputComponent,
  type SelectOption,
} from '@app/shared/components/input';
import { MessageComponent } from '@app/shared/components/message';
import { SongService, Song, CreateSongDto, UpdateSongDto } from '@app/services/song.service';
import { OptionsService } from '@app/shared/services/options.service';
import { runOnOpen } from '@app/shared/utils/dialog';
import { getHttpErrorMessage } from '@app/shared/utils/http-error';
import { focusFirstFocusable } from '@app/shared/utils/focus';

@Component({
  selector: 'app-song-editor-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatIconModule,
    DialogComponent,
    MessageComponent,
    TextInputComponent,
    SelectInputComponent,
  ],
  templateUrl: './song-editor-dialog.html',
  styleUrl: './song-editor-dialog.scss',
})
export class SongEditorDialogComponent {
  private readonly songService = inject(SongService);
  private readonly optionsService = inject(OptionsService);

  readonly song = input<Song | null>(null);
  readonly isOpen = input<boolean>(false);
  readonly saved = output<void>();
  readonly closed = output<void>();

  readonly isSavingSg = signal(false);
  readonly errorSg = signal('');
  readonly titleSg = signal('');
  readonly genreIdsSg = signal<string[]>([]);
  readonly durationSg = signal('');
  readonly selectedArtistIdsSg = signal<string[]>([]);
  readonly albumIdSg = signal<string[]>([]);
  readonly artistOptionsSg = signal<SelectOption[]>([]);
  readonly genreOptionsSg = signal<SelectOption[]>([]);
  readonly albumOptionsSg = signal<SelectOption[]>([]);

  // Computed
  readonly isEditMode = computed(() => !!this.song());
  readonly isCreateMode = computed(() => !this.song());
  readonly dialogTitle = computed(() => (this.isEditMode() ? 'Edit Song' : 'Create Song'));
  readonly submitButtonText = computed(() => (this.isEditMode() ? 'Save' : 'Create'));
  readonly isLoadingSg = signal(false);
  @ViewChild('dialogContent')
  private readonly dialogContentRef?: ElementRef<HTMLElement>;
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
      durationValue > 0 &&
      (!this.isCreateMode() || this.albumIdSg().length > 0)
    );
  });

  constructor() {
    runOnOpen(this.isOpen, () => {
      this.loadArtists();
      this.loadGenres();
      this.loadAlbums();
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
          this.albumIdSg.set([]);
        }
      } else {
        // Create mode - clear form
        this.titleSg.set('');
        this.durationSg.set('');
        this.genreIdsSg.set([]);
        this.selectedArtistIdsSg.set([]);
        this.albumIdSg.set([]);
      }
      this.errorSg.set('');
      setTimeout(() => focusFirstFocusable(this.dialogContentRef?.nativeElement ?? null), 0);
    });
  }

  private loadArtists(): void {
    this.optionsService.loadArtists(100).subscribe({
      next: (options) => {
        this.artistOptionsSg.set(options);
      },
      error: (error) => {
        console.error('Failed to load artists:', error);
      },
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

  private loadAlbums(): void {
    this.optionsService.loadAlbums(200).subscribe({
      next: (options) => {
        this.albumOptionsSg.set(options);
      },
      error: (error) => {
        console.error('Failed to load albums:', error);
      },
    });
  }

  onSave(): void {
    this.isLoadingSg.set(true);
    this.errorSg.set('');
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
          this.errorSg.set(getHttpErrorMessage(error, 'Failed to update song. Please try again.'));
          this.isLoadingSg.set(false);
        },
      });
      return;
    }

    const createDto: CreateSongDto = {
      ...basePayload,
      album_id: this.albumIdSg()[0] ?? '',
    };
    this.songService.createSong(createDto).subscribe({
      next: () => {
        this.isLoadingSg.set(false);
        this.saved.emit();
        this.cancel();
      },
      error: (error) => {
        console.error('Failed to save song:', error);
        this.errorSg.set(getHttpErrorMessage(error, 'Failed to create song. Please try again.'));
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

  @HostListener('document:keydown', ['$event'])
  onDocumentKeydown(event: KeyboardEvent): void {
    if (event.key !== 'Escape' || !this.isOpen()) return;
    event.preventDefault();
    this.cancel();
  }
}
