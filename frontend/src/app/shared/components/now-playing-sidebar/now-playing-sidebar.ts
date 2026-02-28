import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '../cover-art/cover-art';

@Component({
  selector: 'app-now-playing-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, CoverArtComponent],
  templateUrl: './now-playing-sidebar.html',
  styleUrl: './now-playing-sidebar.scss',
})
export class NowPlayingSidebarComponent {
  readonly playback = inject(PlaybackService);
}
