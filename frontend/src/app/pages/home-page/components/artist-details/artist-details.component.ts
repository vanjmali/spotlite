import { Component, inject, signal, effect } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { ArtistService, type Artist } from '../../../../services/artist.service';

@Component({
  selector: 'app-artist-details',
  standalone: true,
  imports: [CommonModule, MatIconModule, WidgetComponent],
  templateUrl: './artist-details.component.html',
  styleUrl: './artist-details.component.scss',
})
export class ArtistDetailsComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly artistService = inject(ArtistService);

  readonly artistSg = signal<Artist | null>(null);
  readonly isLoadingSg = signal(false);

  constructor() {
    effect(() => {
      const artistId = this.route.snapshot.paramMap.get('id');
      if (artistId) {
        this.loadArtist(artistId);
      }
    });
  }

  private loadArtist(artistId: string): void {
    this.isLoadingSg.set(true);
    this.artistService.getArtistById(artistId).subscribe({
      next: (artist) => {
        this.artistSg.set(artist || null);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.isLoadingSg.set(false);
      },
    });
  }

  goBack(): void {
    this.router.navigate(['/home']);
  }
}
