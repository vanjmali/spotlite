import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ItemTableComponent } from '@app/shared/components/item-table';
import { PaginationComponent } from '@app/shared/components/pagination';
import { SongEditorDialogComponent } from '@app/dialogs/song-editor-dialog';
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
  ],
  templateUrl: './songs-management.component.html',
  styleUrl: './songs-management.component.scss',
})
export class SongsManagementComponent {
  private readonly songService = inject(SongService);

  readonly songsSg = signal<Song[]>([]);
  readonly isLoadingSg = signal(false);
  readonly currentPageSg = signal(1);
  readonly pageSizeSg = signal(10);
  readonly totalSg = signal(0);
  readonly isDialogOpenSg = signal(false);
  readonly selectedSongSg = signal<Song | null>(null);

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

  formatDuration(seconds: number): string {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  }
}
