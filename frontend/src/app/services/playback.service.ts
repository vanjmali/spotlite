import { Injectable, computed, signal } from '@angular/core';
import { Album } from './album.service';
import { Artist } from './artist.service';
import { Song } from './song.service';

export interface PlaybackTrack {
  id: string;
  title: string;
  artists: Artist[];
  lengthSeconds?: number;
  albumId: string;
  albumTitle: string;
}

@Injectable({
  providedIn: 'root',
})
export class PlaybackService {
  private readonly audio = new Audio();

  readonly currentTrackSg = signal<PlaybackTrack | null>(null);
  readonly currentAlbumSg = signal<Album | null>(null);
  readonly queueSg = signal<PlaybackTrack[]>([]);
  readonly currentIndexSg = signal<number>(-1);

  readonly isPlayingSg = signal(false);
  readonly currentTimeSg = signal(0);
  readonly durationSg = signal(0);
  readonly volumeSg = signal(0.85);
  readonly ratingsBySongSg = signal<Record<string, number>>({});

  readonly progressPercentSg = computed(() => {
    const duration = this.durationSg();
    if (!duration) {
      return 0;
    }

    return Math.min((this.currentTimeSg() / duration) * 100, 100);
  });

  readonly currentRatingSg = computed(() => {
    const id = this.currentTrackSg()?.id;
    if (!id) {
      return 0;
    }

    return this.ratingsBySongSg()[id] ?? 0;
  });

  constructor() {
    this.audio.preload = 'metadata';
    this.audio.volume = this.volumeSg();

    this.audio.addEventListener('loadedmetadata', () => {
      const duration = Number.isFinite(this.audio.duration) ? this.audio.duration : 0;
      this.durationSg.set(duration);
    });

    this.audio.addEventListener('timeupdate', () => {
      this.currentTimeSg.set(this.audio.currentTime || 0);
    });

    this.audio.addEventListener('play', () => this.isPlayingSg.set(true));
    this.audio.addEventListener('pause', () => this.isPlayingSg.set(false));
    this.audio.addEventListener('ended', () => this.next());
  }

  playAlbum(album: Album, startSongId?: string): void {
    const queue = (album.songs ?? []).map((song) => this.mapSongToTrack(song, album));
    if (queue.length === 0) {
      return;
    }

    let startIndex = 0;
    if (startSongId) {
      const foundIndex = queue.findIndex((song) => song.id === startSongId);
      startIndex = foundIndex >= 0 ? foundIndex : 0;
    }

    this.currentAlbumSg.set(album);
    this.queueSg.set(queue);
    this.playFromQueue(startIndex);
  }

  playSingleSong(song: Song, album: Album): void {
    this.playAlbum(album, song.id);
  }

  playFromQueue(index: number): void {
    const queue = this.queueSg();
    if (index < 0 || index >= queue.length) {
      return;
    }

    const track = queue[index];
    this.currentIndexSg.set(index);
    this.currentTrackSg.set(track);
    this.durationSg.set(track.lengthSeconds ?? 0);
    this.currentTimeSg.set(0);

    this.audio.src = this.streamUrl(track.id);
    void this.audio.play().catch(() => {
      this.isPlayingSg.set(false);
    });
  }

  togglePlayPause(): void {
    if (!this.currentTrackSg()) {
      return;
    }

    if (this.audio.paused) {
      void this.audio.play().catch(() => {
        this.isPlayingSg.set(false);
      });
      return;
    }

    this.audio.pause();
  }

  previous(): void {
    const queue = this.queueSg();
    if (queue.length === 0) {
      return;
    }

    const current = this.currentIndexSg();
    if (current <= 0) {
      this.seek(0);
      return;
    }

    this.playFromQueue(current - 1);
  }

  next(): void {
    const queue = this.queueSg();
    if (queue.length === 0) {
      return;
    }

    const current = this.currentIndexSg();
    if (current >= queue.length - 1) {
      this.audio.pause();
      this.audio.currentTime = 0;
      this.isPlayingSg.set(false);
      this.currentTimeSg.set(0);
      return;
    }

    this.playFromQueue(current + 1);
  }

  seek(percentage: number): void {
    const duration = this.durationSg();
    if (!duration) {
      return;
    }

    const clamped = Math.max(0, Math.min(100, percentage));
    this.audio.currentTime = (clamped / 100) * duration;
    this.currentTimeSg.set(this.audio.currentTime);
  }

  setVolume(volume: number): void {
    const clamped = Math.max(0, Math.min(1, volume));
    this.audio.volume = clamped;
    this.volumeSg.set(clamped);
  }

  setRating(rating: number): void {
    const songId = this.currentTrackSg()?.id;
    if (!songId) {
      return;
    }

    const normalized = Math.max(0, Math.min(5, Math.round(rating)));
    const currentMap = this.ratingsBySongSg();

    this.ratingsBySongSg.set({
      ...currentMap,
      [songId]: normalized,
    });
  }

  private streamUrl(songId: string): string {
    return `/api/content/songs/${songId}/audio`;
  }

  private mapSongToTrack(song: Song, album: Album): PlaybackTrack {
    return {
      id: song.id,
      title: song.title,
      artists: song.artists ?? [],
      lengthSeconds: song.lengthSeconds,
      albumId: album.id,
      albumTitle: album.title,
    };
  }
}

