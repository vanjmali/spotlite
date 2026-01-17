import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { WidgetComponent } from '@app/shared/components/widget';
import { ArtistEditorDialogComponent } from '@app/dialogs/artist-editor-dialog';
import { ArtistService, Artist } from '@app/services/artist.service';

@Component({
  selector: 'app-artists-management',
  standalone: true,
  imports: [
    CommonModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    WidgetComponent,
    ArtistEditorDialogComponent,
  ],
  templateUrl: './artists-management.component.html',
  styleUrl: './artists-management.component.scss',
})
export class ArtistsManagementComponent {
  private readonly artistService = inject(ArtistService);

  readonly artistsSg = signal<Artist[]>([]);
  readonly isLoadingSg = signal(false);
  // readonly isDeleteLoadingSg = signal<string | null>(null); // Delete disabled for now
  readonly currentPageSg = signal(1);
  readonly pageSizeSg = signal(10);
  readonly isDialogOpenSg = signal(false);
  readonly selectedArtistSg = signal<Artist | null>(null);

  constructor() {
    effect(() => {
      this.loadArtists();
    });
  }

  private loadArtists(): void {
    this.isLoadingSg.set(true);
    this.artistService.getArtists(this.currentPageSg(), this.pageSizeSg()).subscribe({
      next: (response) => {
        this.artistsSg.set(response.items || []);
        this.isLoadingSg.set(false);
      },
      error: (error) => {
        console.error('Failed to load artists:', error);
        this.isLoadingSg.set(false);
      },
    });
  }

  onAddNew(): void {
    this.selectedArtistSg.set(null);
    this.isDialogOpenSg.set(true);
  }

  onEdit(artist: Artist): void {
    this.selectedArtistSg.set(artist);
    this.isDialogOpenSg.set(true);
  }

  onDialogSaved(): void {
    this.isDialogOpenSg.set(false);
    this.selectedArtistSg.set(null);
    this.loadArtists();
  }
}
