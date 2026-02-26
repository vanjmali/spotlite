import { CommonModule } from '@angular/common';
import { Component, computed, inject, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink } from '@angular/router';
import { PlaybackService } from '@app/services/playback.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { BUILD_YEAR } from '@app/shared/build-info';

@Component({
  selector: 'app-control-bar',
  standalone: true,
  imports: [CommonModule, RouterLink, CoverArtComponent, MatIconModule],
  templateUrl: './control-bar.html',
  styleUrls: ['./control-bar.scss'],
})
export class ControlBarComponent {
  readonly homepageModeSg = input<boolean>(false, { alias: 'homepageMode' });
  readonly year = BUILD_YEAR;
  readonly playback = inject(PlaybackService);

  readonly stars = [1, 2, 3, 4, 5];
  readonly hasTrackSg = computed(() => !!this.playback.currentTrackSg());
  readonly currentTrackSg = computed(() => this.playback.currentTrackSg());
  readonly currentAlbumTitleSg = computed(
    () =>
      this.playback.currentAlbumSg()?.title ??
      this.playback.currentTrackSg()?.albumTitle ??
      'No album'
  );
  readonly currentVolumePercentSg = computed(() => Math.round(this.playback.volumeSg() * 100));
  readonly isMutedSg = computed(() => this.playback.mutedSg());

  previous(): void {
    this.playback.previous();
  }

  next(): void {
    this.playback.next();
  }

  togglePlayback(): void {
    this.playback.togglePlayPause();
  }

  rate(star: number): void {
    this.playback.setRating(star);
  }

  onVolumeInput(rawValue: string): void {
    const volume = Number(rawValue);
    this.playback.setVolume(volume / 100);
  }

  toggleMute(): void {
    this.playback.toggleMute();
  }

  volumeIcon(): string {
    if (this.isMutedSg()) {
      return 'volume_off';
    }
    const volume = this.currentVolumePercentSg();
    if (volume <= 50) {
      return 'volume_down';
    }
    return 'volume_up';
  }
}
