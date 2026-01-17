import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { ArtistService, type Artist } from '../../../../services/artist.service';

@Component({
  selector: 'app-artists-list',
  standalone: true,
  imports: [CommonModule, WidgetComponent],
  templateUrl: './artists-list.component.html',
  styleUrl: './artists-list.component.scss',
})
export class ArtistsListComponent {
  private readonly router = inject(Router);
  private readonly artistService = inject(ArtistService);

  readonly artistsSg = signal<Artist[]>([]);
  readonly isLoadingSg = signal(false);

  constructor() {
    this.loadArtists();
  }

  private loadArtists(): void {
    this.isLoadingSg.set(true);
    this.artistService.getArtists().subscribe({
      next: (response) => {
        this.artistsSg.set(response.items || []);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.isLoadingSg.set(false);
      },
    });
  }

  selectArtist(artist: Artist): void {
    this.router.navigate(['/home', 'artist', artist.id]);
  }
}
