import { CommonModule } from '@angular/common';
import { Component, input, output } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink } from '@angular/router';
import { Album } from '@app/services/album.service';
import { Artist } from '@app/services/artist.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';

@Component({
  selector: 'app-album-card',
  standalone: true,
  imports: [CommonModule, RouterLink, MatIconModule, CoverArtComponent],
  templateUrl: './album-card.component.html',
  styleUrl: './album-card.component.scss',
})
export class AlbumCardComponent {
  readonly albumSg = input.required<Album>({ alias: 'album' });
  readonly clickableSg = input<boolean>(true, { alias: 'clickable' });
  readonly linkPrefixSg = input<string>('album', { alias: 'linkPrefix' });
  readonly showPlayControlSg = input<boolean>(false, { alias: 'showPlayControl' });
  readonly isPlayingSg = input<boolean>(false, { alias: 'isPlaying' });

  readonly playToggle = output<Event>();

  artistNames(artists: Artist[] | undefined): string {
    if (!artists || artists.length === 0) {
      return 'Unknown artist';
    }

    return artists.map((artist) => artist.name).join(', ');
  }

  onPlayToggle(event: Event): void {
    this.playToggle.emit(event);
  }
}
