import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ItemTableComponent } from '@app/shared/components/item-table';
import { PaginationComponent } from '@app/shared/components/pagination';
import { SongEditorDialogComponent } from '@app/dialogs/song-editor-dialog';
import { DialogComponent } from '@app/shared/components/dialog';
import { SongService, Song } from '@app/services/song.service';

@Component({
  selector: 'app-songs-management',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    ItemTableComponent,
    PaginationComponent,
    SongEditorDialogComponent,
    DialogComponent,
  ],
  templateUrl: './songs-management.component.html',
  styleUrl: './songs-management.component.scss',
})
export class SongsManagementComponent {
  private readonly songService = inject(SongService);

  readonly songsSg = signal<Song[]>([]);
  readonly isLoadingSg = signal(false);
  readonly isDeletingSg = signal(false);
  readonly currentPageSg = signal(1);
  readonly pageSizeSg = signal(10);
  readonly totalSg = signal(0);
  readonly isDialogOpenSg = signal(false);
  readonly selectedSongSg = signal<Song | null>(null);
  readonly isDeleteDialogOpenSg = signal(false);
  readonly songToDeleteSg = signal<Song | null>(null);

  constructor() {
    effect(() => {
      this.loadSongs();
    });
  }

  private loadSongs(): void {
    this.isLoadingSg.set(true);
    this.songService.getSongs(this.currentPageSg(), this.pageSizeSg()).subscribe({
      next: (response) => {
        this.songsSg.set(response.items || []);
        this.totalSg.set(response.total ?? 0);
        this.isLoadingSg.set(false);
      },
      error: (error) => {
        console.error('Failed to load songs:', error);
        this.isLoadingSg.set(false);
      },
    });
  }

  onAddNew(): void {
    this.selectedSongSg.set(null);
    this.isDialogOpenSg.set(true);
  }

  onEdit(song: Song): void {
    this.selectedSongSg.set(song);
    this.isDialogOpenSg.set(true);
  }

  onDialogSaved(): void {
    this.isDialogOpenSg.set(false);
    this.selectedSongSg.set(null);
    this.loadSongs();
  }

  onPageChange(page: number): void {
    this.currentPageSg.set(page);
  }

  onPageSizeChange(size: number): void {
    this.pageSizeSg.set(size);
    this.currentPageSg.set(1);
  }

  onDelete(song: Song): void {
    this.songToDeleteSg.set(song);
    this.isDeleteDialogOpenSg.set(true);
  }

  cancelDelete(): void {
    if (this.isDeletingSg()) {
      return;
    }

    this.isDeleteDialogOpenSg.set(false);
    this.songToDeleteSg.set(null);
  }

  confirmDelete(): void {
    const song = this.songToDeleteSg();
    if (!song || this.isDeletingSg()) {
      return;
    }

    this.isDeletingSg.set(true);
    this.songService.deleteSong(song.id).subscribe({
      next: () => {
        this.songsSg.set(this.songsSg().filter((s) => s.id !== song.id));
        this.totalSg.set(Math.max(0, this.totalSg() - 1));
        this.isDeletingSg.set(false);
        this.isDeleteDialogOpenSg.set(false);
        this.songToDeleteSg.set(null);
      },
      error: (error) => {
        console.error('Failed to delete song:', error);
        this.isDeletingSg.set(false);
      },
    });
  }

  formatDuration(seconds: number): string {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  }
}
