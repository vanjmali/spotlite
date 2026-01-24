import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { DialogComponent } from '@app/shared/components/dialog';
import { AlbumService, Album } from '@app/services/album.service';
import { SongService, type Song } from '@app/services/song.service';

@Component({
  selector: 'app-album-songs-dialog',
  standalone: true,
  imports: [CommonModule, MatIconModule, DialogComponent],
  templateUrl: './album-songs-dialog.html',
  styleUrl: './album-songs-dialog.scss',
})
export class AlbumSongsDialogComponent {
  private readonly albumService = inject(AlbumService);
  private readonly songService = inject(SongService);

  readonly album = input<Album | null>(null);
  readonly isOpen = input<boolean>(false);
  readonly saved = output<void>();
  readonly closed = output<void>();

  readonly albumSongsSg = signal<Song[]>([]);
  readonly availableSongsSg = signal<Song[]>([]);
  readonly selectedAddSongIdsSg = signal<string[]>([]);
  readonly isLoadingAlbumSongsSg = signal(false);
  readonly isLoadingAvailableSongsSg = signal(false);
  readonly isAddingSg = signal(false);
  readonly removingSongIdsSg = signal<Set<string>>(new Set());

  readonly dialogTitle = computed(() => {
    const album = this.album();
    return album ? `Manage Songs — ${album.title}` : 'Manage Album Songs';
  });

  readonly canAddSongs = computed(
    () =>
      this.selectedAddSongIdsSg().length > 0 &&
      !this.isAddingSg() &&
      !this.isLoadingAvailableSongsSg()
  );

  constructor() {
    effect(() => {
      if (this.isOpen()) {
        this.selectedAddSongIdsSg.set([]);
        this.loadAlbumSongs();
      }
    });
  }

  private loadAlbumSongs(): void {
    const album = this.album();
    if (!album) {
      this.albumSongsSg.set([]);
      this.availableSongsSg.set([]);
      this.isLoadingAlbumSongsSg.set(false);
      this.isLoadingAvailableSongsSg.set(false);
      return;
    }

    this.isLoadingAlbumSongsSg.set(true);
    this.albumService.getAlbumSongs(album.id).subscribe({
      next: (songs) => {
        const albumSongs = songs || [];
        this.albumSongsSg.set(albumSongs);
        this.isLoadingAlbumSongsSg.set(false);
        this.loadAvailableSongsForAlbum(album, albumSongs);
      },
      error: (error) => {
        console.error('Failed to load album songs:', error);
        this.isLoadingAlbumSongsSg.set(false);
      },
    });
  }

  private loadAvailableSongsForAlbum(album: Album, albumSongs: Song[]): void {
    const artistIds = album.artists?.map((artist) => artist.id).filter(Boolean) ?? [];
    if (artistIds.length === 0) {
      this.availableSongsSg.set([]);
      this.isLoadingAvailableSongsSg.set(false);
      return;
    }

    this.isLoadingAvailableSongsSg.set(true);
    const allSongs: Song[] = [];
    let loadedCount = 0;

    artistIds.forEach((artistId) => {
      this.songService.getSongs(1, 200, { artist_id: artistId }).subscribe({
        next: (response) => {
          allSongs.push(...(response.items || []));
          loadedCount++;
          if (loadedCount === artistIds.length) {
            this.setAvailableSongs(allSongs, albumSongs);
          }
        },
        error: (error) => {
          console.error('Failed to load songs for artist:', error);
          loadedCount++;
          if (loadedCount === artistIds.length) {
            this.setAvailableSongs(allSongs, albumSongs);
          }
        },
      });
    });
  }

  private setAvailableSongs(allSongs: Song[], albumSongs: Song[]): void {
    const albumSongIds = new Set(albumSongs.map((song) => song.id));
    const uniqueSongs = Array.from(new Map(allSongs.map((song) => [song.id, song])).values())
      .filter((song) => !albumSongIds.has(song.id))
      .sort((a, b) => a.title.localeCompare(b.title));

    this.availableSongsSg.set(uniqueSongs);
    this.isLoadingAvailableSongsSg.set(false);
  }

  toggleAddSongSelection(songId: string): void {
    this.selectedAddSongIdsSg.update((ids) => {
      if (ids.includes(songId)) {
        return ids.filter((id) => id !== songId);
      }
      return [...ids, songId];
    });
  }

  isAddSongSelected(songId: string): boolean {
    return this.selectedAddSongIdsSg().includes(songId);
  }

  addSelectedSongs(): void {
    const album = this.album();
    const selectedIds = this.selectedAddSongIdsSg();
    if (!album || selectedIds.length === 0) {
      return;
    }

    this.isAddingSg.set(true);
    this.albumService.addAlbumSongs(album.id, { ids: selectedIds }).subscribe({
      next: () => {
        this.isAddingSg.set(false);
        this.selectedAddSongIdsSg.set([]);
        this.saved.emit();
        this.loadAlbumSongs();
      },
      error: (error) => {
        console.error('Failed to add songs to album:', error);
        this.isAddingSg.set(false);
      },
    });
  }

  removeSong(songId: string): void {
    const album = this.album();
    if (!album) {
      return;
    }

    this.setRemoving(songId, true);
    this.albumService.removeAlbumSong(album.id, songId).subscribe({
      next: () => {
        this.setRemoving(songId, false);
        this.saved.emit();
        this.loadAlbumSongs();
      },
      error: (error) => {
        console.error('Failed to remove song from album:', error);
        this.setRemoving(songId, false);
      },
    });
  }

  isRemovingSong(songId: string): boolean {
    return this.removingSongIdsSg().has(songId);
  }

  private setRemoving(songId: string, isRemoving: boolean): void {
    this.removingSongIdsSg.update((current) => {
      const next = new Set(current);
      if (isRemoving) {
        next.add(songId);
      } else {
        next.delete(songId);
      }
      return next;
    });
  }

  cancel(): void {
    this.closed.emit();
  }

  onClose(): void {
    this.cancel();
  }

  formatGenres(genres: Song['genres'] | undefined): string {
    if (!genres || genres.length === 0) return '—';
    return genres.map((genre) => genre.name).join(', ');
  }
}
