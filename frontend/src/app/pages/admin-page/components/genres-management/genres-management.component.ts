import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { WidgetComponent } from '@app/shared/components/widget';
import { GenreEditorDialogComponent } from '@app/dialogs/genre-editor-dialog';
import { GenreService, Genre } from '@app/services/genre.service';

@Component({
  selector: 'app-genres-management',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    WidgetComponent,
    GenreEditorDialogComponent,
  ],
  templateUrl: './genres-management.component.html',
  styleUrl: './genres-management.component.scss',
})
export class GenresManagementComponent {
  private readonly genreService = inject(GenreService);

  readonly genresSg = signal<Genre[]>([]);
  readonly isLoadingSg = signal(false);
  readonly currentPageSg = signal(1);
  readonly pageSizeSg = signal(20);
  readonly isDialogOpenSg = signal(false);
  readonly selectedGenreSg = signal<Genre | null>(null);

  constructor() {
    effect(() => {
      this.loadGenres();
    });
  }

  private loadGenres(): void {
    this.isLoadingSg.set(true);
    this.genreService.getGenres(this.currentPageSg(), this.pageSizeSg()).subscribe({
      next: (response) => {
        this.genresSg.set(response.items || []);
        this.isLoadingSg.set(false);
      },
      error: (error) => {
        console.error('Failed to load genres:', error);
        this.isLoadingSg.set(false);
      },
    });
  }

  onAddNew(): void {
    this.selectedGenreSg.set(null);
    this.isDialogOpenSg.set(true);
  }

  onEdit(genre: Genre): void {
    this.selectedGenreSg.set(genre);
    this.isDialogOpenSg.set(true);
  }

  onDelete(genre: Genre): void {
    if (!confirm(`Delete genre "${genre.name}"?`)) {
      return;
    }

    this.isLoadingSg.set(true);
    this.genreService.deleteGenre(genre.id).subscribe({
      next: () => {
        this.isLoadingSg.set(false);
        this.loadGenres();
      },
      error: (error) => {
        console.error('Failed to delete genre:', error);
        this.isLoadingSg.set(false);
      },
    });
  }

  onDialogSaved(): void {
    this.isDialogOpenSg.set(false);
    this.selectedGenreSg.set(null);
    this.loadGenres();
  }
}
