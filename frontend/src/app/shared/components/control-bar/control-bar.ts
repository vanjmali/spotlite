import { CommonModule } from '@angular/common';
import { Component, HostListener, computed, inject, input, signal } from '@angular/core';
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
  readonly isLoadingSg = computed(() => this.playback.isLoadingSg());
  readonly hoveredStarSg = signal(0);
  readonly previewRatingSg = computed(() =>
    this.hoveredStarSg() > 0 ? this.hoveredStarSg() : this.playback.currentRatingSg()
  );

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

  onStarHover(star: number): void {
    this.hoveredStarSg.set(star);
  }

  clearStarHover(): void {
    this.hoveredStarSg.set(0);
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

  @HostListener('window:keydown', ['$event'])
  onWindowKeydown(event: KeyboardEvent): void {
    const action = this.mediaActionFromKey(event);
    if (!action) {
      return;
    }

    if (event.defaultPrevented || this.isInteractiveTarget(event.target)) {
      return;
    }

    if (!this.hasTrackSg()) {
      return;
    }

    event.preventDefault();
    switch (action) {
      case 'toggle':
        this.togglePlayback();
        break;
      case 'play':
        if (!this.playback.isPlayingSg()) {
          this.togglePlayback();
        }
        break;
      case 'pause':
      case 'stop':
        if (this.playback.isPlayingSg()) {
          this.togglePlayback();
        }
        break;
      case 'next':
        this.next();
        break;
      case 'previous':
        this.previous();
        break;
      default:
        break;
    }
  }

  private mediaActionFromKey(
    event: KeyboardEvent
  ): 'toggle' | 'play' | 'pause' | 'next' | 'previous' | 'stop' | null {
    if (event.code === 'Space' || event.key === ' ') {
      return 'toggle';
    }

    const mediaCode = event.code || event.key;
    switch (mediaCode) {
      case 'MediaPlayPause':
        return 'toggle';
      case 'MediaPlay':
        return 'play';
      case 'MediaPause':
        return 'pause';
      case 'MediaTrackNext':
        return 'next';
      case 'MediaTrackPrevious':
        return 'previous';
      case 'MediaStop':
        return 'stop';
      default:
        return null;
    }
  }

  private isInteractiveTarget(target: EventTarget | null): boolean {
    const el = target as HTMLElement | null;
    if (!el) {
      return false;
    }

    if (el.isContentEditable) {
      return true;
    }

    const tag = el.tagName?.toLowerCase();
    if (tag === 'input' || tag === 'textarea' || tag === 'select' || tag === 'button') {
      return true;
    }

    return Boolean(el.closest('input, textarea, select, button, [contenteditable="true"]'));
  }
}
