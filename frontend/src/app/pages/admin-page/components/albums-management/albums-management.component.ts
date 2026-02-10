import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ItemTableComponent } from '@app/shared/components/item-table';
import { PaginationComponent } from '@app/shared/components/pagination';
import { AlbumEditorDialogComponent } from '@app/dialogs/album-editor-dialog';
import { AlbumSongsDialogComponent } from '@app/dialogs/album-songs-dialog';
import { AlbumService, Album } from '@app/services/album.service';

@Component({
  selector: 'app-albums-management',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    ItemTableComponent,
    PaginationComponent,
    AlbumEditorDialogComponent,
    AlbumSongsDialogComponent,
  ],
  templateUrl: './albums-management.component.html',
  styleUrl: './albums-management.component.scss',
})
export class AlbumsManagementComponent {
  private readonly albumService = inject(AlbumService);

  readonly albumsSg = signal<Album[]>([]);
  readonly isLoadingSg = signal(false);
  readonly currentPageSg = signal(1);
  readonly pageSizeSg = signal(10);
  readonly totalSg = signal(0);
  readonly isDialogOpenSg = signal(false);
  readonly selectedAlbumSg = signal<Album | null>(null);
  readonly isSongsDialogOpenSg = signal(false);
  readonly selectedSongsAlbumSg = signal<Album | null>(null);

  constructor() {
    effect(() => {
      this.loadAlbums();
    });
  }

  private loadAlbums(): void {
    this.isLoadingSg.set(true);
    this.albumService.getAlbums(this.currentPageSg(), this.pageSizeSg()).subscribe({
      next: (response) => {
        this.albumsSg.set(response.items || []);
        this.totalSg.set(response.total ?? 0);
        this.isLoadingSg.set(false);
      },
      error: (error) => {
        console.error('Failed to load albums:', error);
        this.isLoadingSg.set(false);
      },
    });
  }

  onAddNew(): void {
    this.selectedAlbumSg.set(null);
    this.isDialogOpenSg.set(true);
  }

  onEdit(album: Album): void {
    this.selectedAlbumSg.set(album);
    this.isDialogOpenSg.set(true);
  }

  onManageSongs(album: Album): void {
    this.selectedSongsAlbumSg.set(album);
    this.isSongsDialogOpenSg.set(true);
  }

  onDialogSaved(): void {
    this.isDialogOpenSg.set(false);
    this.selectedAlbumSg.set(null);
    this.loadAlbums();
  }

  onSongsDialogSaved(): void {
    this.loadAlbums();
  }

  onSongsDialogClosed(): void {
    this.isSongsDialogOpenSg.set(false);
    this.selectedSongsAlbumSg.set(null);
  }

  onPageChange(page: number): void {
    this.currentPageSg.set(page);
  }

  onPageSizeChange(size: number): void {
    this.pageSizeSg.set(size);
    this.currentPageSg.set(1);
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  }
}
