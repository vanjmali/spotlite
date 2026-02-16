import { Component, inject, signal, effect } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { WidgetComponent } from '@app/shared/components/widget/widget.component';
import { AlbumService, type Album } from '../../../../services/album.service';

@Component({
  selector: 'app-album-details',
  standalone: true,
  imports: [CommonModule, MatIconModule, WidgetComponent],
  templateUrl: './album-details.component.html',
  styleUrl: './album-details.component.scss',
})
export class AlbumDetailsComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly albumService = inject(AlbumService);

  readonly albumSg = signal<Album | null>(null);
  readonly isLoadingSg = signal(false);

  constructor() {
    effect(() => {
      const albumId = this.route.snapshot.paramMap.get('id');
      if (albumId) {
        this.loadAlbum(albumId);
      }
    });
  }

  private loadAlbum(albumId: string): void {
    this.isLoadingSg.set(true);
    this.albumService.getAlbumById(albumId).subscribe({
      next: (album) => {
        this.albumSg.set(album || null);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.isLoadingSg.set(false);
      },
    });
  }

  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  // TODO: implement to go page back instead of home
  // goBack(): void {
  //   this.router.navigate(['/']);
  // }
}
