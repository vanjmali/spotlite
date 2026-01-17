import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { WidgetComponent } from '@app/shared/components/widget';
import { AlbumEditorDialogComponent } from '@app/dialogs/album-editor-dialog';
import { AlbumService, Album } from '@app/services/album.service';

@Component({
  selector: 'app-albums-management',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    WidgetComponent,
    AlbumEditorDialogComponent,
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
  readonly isDialogOpenSg = signal(false);
  readonly selectedAlbumSg = signal<Album | null>(null);

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

  onDialogSaved(): void {
    this.isDialogOpenSg.set(false);
    this.selectedAlbumSg.set(null);
    this.loadAlbums();
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  }
}
