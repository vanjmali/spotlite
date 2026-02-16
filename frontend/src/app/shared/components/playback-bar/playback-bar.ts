import { CommonModule } from '@angular/common';
import { Component, computed, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { PlaybackService } from '@app/services/playback.service';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-playback-bar',
  standalone: true,
  imports: [CommonModule, RouterLink, MatIconModule],
  templateUrl: './playback-bar.html',
  styleUrl: './playback-bar.scss',
})
export class PlaybackBarComponent {
  readonly playback = inject(PlaybackService);
  readonly stars = [1, 2, 3, 4, 5];

  readonly currentTrackArtistsSg = computed(
    () => this.playback.currentTrackSg()?.artists?.map((artist) => artist.name).join(', ') || 'Unknown'
  );

  togglePlayPause(): void {
    this.playback.togglePlayPause();
  }

  onSeek(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    if (!input) {
      return;
    }

    this.playback.seek(Number(input.value));
  }

  onVolume(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    if (!input) {
      return;
    }

    this.playback.setVolume(Number(input.value) / 100);
  }

  rate(value: number): void {
    this.playback.setRating(value);
  }

  formatTime(totalSeconds: number): string {
    const seconds = Math.max(0, Math.floor(totalSeconds));
    const minutes = Math.floor(seconds / 60);
    const remainder = seconds % 60;
    return `${minutes}:${remainder.toString().padStart(2, '0')}`;
  }
}

