import { Injectable, computed, effect, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Album } from './album.service';
import { Artist } from './artist.service';
import { Song } from './song.service';
import { AuthService } from './auth.service';
import { RatingService } from './rating.service';

export interface PlayedTrack {
  id: string;
  title: string;
  albumId: string;
  albumTitle: string;
  artists: Artist[];
  playedAt: string;
}

export interface PlaybackTrack {
  id: string;
  title: string;
  artists: Artist[];
  lengthSeconds?: number;
  albumId: string;
  albumTitle: string;
}

type PersistedPlaybackState = {
  queue: PlaybackTrack[];
  currentIndex: number;
  currentAlbum: {
    id: string;
    title: string;
    artists: Artist[];
  } | null;
};

type PersistedAudioState = {
  volume: number;
  muted: boolean;
};

type SignedAudioUrlResponse = {
  url: string;
  expires_at: string;
};

const PLAYBACK_STATE_STORAGE_KEY = 'spotlite.playback.state.v1';
const PLAYBACK_AUDIO_STORAGE_KEY = 'spotlite.playback.audio.v1';
const DEFAULT_VOLUME = 0.85;

@Injectable({
  providedIn: 'root',
})
export class PlaybackService {
  private readonly audio = new Audio();
  private readonly http = inject(HttpClient);
  private readonly authService = inject(AuthService);
  private readonly ratingService = inject(RatingService);
  private streamLoadId = 0;
  private lastVolumeBeforeMute = DEFAULT_VOLUME;

  readonly currentTrackSg = signal<PlaybackTrack | null>(null);
  readonly currentAlbumSg = signal<Album | null>(null);
  readonly queueSg = signal<PlaybackTrack[]>([]);
  readonly currentIndexSg = signal<number>(-1);

  readonly isPlayingSg = signal(false);
  readonly isLoadingSg = signal(false);
  readonly currentTimeSg = signal(0);
  readonly durationSg = signal(0);
  readonly volumeSg = signal(DEFAULT_VOLUME);
  readonly mutedSg = signal(false);
  readonly ratingsBySongSg = signal<Record<string, number>>({});
  readonly ratingIdsBySongSg = signal<Record<string, string>>({});
  readonly playHistorySg = signal<PlayedTrack[]>([]);

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
    this.restoreAudioState();
    this.restorePlaybackState();

    this.audio.preload = 'metadata';
    this.audio.volume = this.volumeSg();
    this.audio.muted = this.mutedSg();

    this.audio.addEventListener('loadedmetadata', () => {
      const duration = Number.isFinite(this.audio.duration) ? this.audio.duration : 0;
      this.durationSg.set(duration);
    });

    this.audio.addEventListener('timeupdate', () => {
      this.currentTimeSg.set(this.audio.currentTime || 0);
    });

    this.audio.addEventListener('play', () => {
      this.isPlayingSg.set(true);
      this.isLoadingSg.set(false);
    });
    this.audio.addEventListener('pause', () => this.isPlayingSg.set(false));
    this.audio.addEventListener('waiting', () => {
      if (this.currentTrackSg() && !this.audio.paused) {
        this.isLoadingSg.set(true);
      }
    });
    this.audio.addEventListener('stalled', () => {
      if (this.currentTrackSg() && !this.audio.paused) {
        this.isLoadingSg.set(true);
      }
    });
    this.audio.addEventListener('canplay', () => {
      if (!this.audio.paused) {
        this.isLoadingSg.set(false);
      }
    });
    this.audio.addEventListener('ended', () => this.next());
    this.setupMediaSessionHandlers();

    effect(() => {
      this.audio.volume = this.volumeSg();
      this.audio.muted = this.mutedSg();
      this.persistAudioState();
    });

    effect(() => {
      this.persistPlaybackState();
    });

    effect(() => {
      this.syncMediaSessionMetadata();
    });

    effect(() => {
      const userId = this.authService.currentUserIdSg();
      if (!userId) {
        this.ratingsBySongSg.set({});
        this.ratingIdsBySongSg.set({});
        return;
      }
      this.loadUserRatings(userId);
    });
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
    this.isPlayingSg.set(false);
    this.isLoadingSg.set(true);

    const loadId = ++this.streamLoadId;
    this.clearAudioSource();

    this.http.get<SignedAudioUrlResponse>(this.signedStreamUrl(track.id)).subscribe({
      next: (res) => {
        if (loadId !== this.streamLoadId) {
          return;
        }

        if (!res?.url) {
          this.isLoadingSg.set(false);
          return;
        }

        this.setAudioSource(res.url);
        void this.audio.play().catch(() => {
          this.isPlayingSg.set(false);
          this.isLoadingSg.set(false);
        });
      },
      error: () => {
        if (loadId !== this.streamLoadId) {
          return;
        }
        this.clearAudioSource();
        this.isPlayingSg.set(false);
        this.isLoadingSg.set(false);
      },
    });

    this.recordPlay(track);
  }

  togglePlayPause(): void {
    const currentTrack = this.currentTrackSg();
    if (!currentTrack) {
      return;
    }

    if (this.audio.paused) {
      if (!this.audio.src) {
        const queue = this.queueSg();
        const currentIndex = this.currentIndexSg();
        const playIndex =
          currentIndex >= 0 && currentIndex < queue.length
            ? currentIndex
            : queue.findIndex((track) => track.id === currentTrack.id);
        if (playIndex >= 0) {
          this.playFromQueue(playIndex);
        } else {
          // Recover gracefully from stale/restored state where queue/index is missing.
          this.queueSg.set([currentTrack]);
          this.currentIndexSg.set(0);
          this.playFromQueue(0);
        }
        return;
      }

      void this.audio.play().catch(() => {
        this.isPlayingSg.set(false);
        this.isLoadingSg.set(false);
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
    this.volumeSg.set(clamped);
    this.mutedSg.set(clamped === 0);
    if (clamped > 0) {
      this.lastVolumeBeforeMute = clamped;
    }
  }

  toggleMute(): void {
    if (this.mutedSg()) {
      const restored = this.lastVolumeBeforeMute > 0 ? this.lastVolumeBeforeMute : DEFAULT_VOLUME;
      this.volumeSg.set(restored);
      this.mutedSg.set(false);
      return;
    }

    if (this.volumeSg() > 0) {
      this.lastVolumeBeforeMute = this.volumeSg();
    }
    this.mutedSg.set(true);
  }

  setRating(rating: number): void {
    const songId = this.currentTrackSg()?.id;
    if (!songId) {
      return;
    }

    const normalized = Math.max(0, Math.min(5, Math.round(rating)));
    const currentValue = this.ratingsBySongSg()[songId] ?? 0;
    const nextValue = currentValue === normalized ? 0 : normalized;
    const existingRatingId = this.ratingIdsBySongSg()[songId];

    if (nextValue === 0) {
      if (!existingRatingId) {
        this.removeLocalRating(songId);
        return;
      }

      this.ratingService.deleteRating(existingRatingId).subscribe({
        next: () => {
          this.removeLocalRating(songId);
        },
      });
      return;
    }

    if (existingRatingId) {
      this.ratingService.updateRating(existingRatingId, nextValue).subscribe({
        next: () => {
          this.upsertLocalRating(songId, nextValue, existingRatingId);
        },
      });
      return;
    }

    this.ratingService.createRating(songId, nextValue).subscribe({
      next: () => {
        this.syncSongRatingFromApi(songId, nextValue);
      },
    });
  }

  private streamUrl(songId: string): string {
    return `/api/content/songs/${songId}/audio`;
  }

  private signedStreamUrl(songId: string): string {
    return `${this.streamUrl(songId)}/signed-url`;
  }

  private setAudioSource(url: string): void {
    this.clearAudioSource();
    this.audio.src = url;
  }

  private clearAudioSource(): void {
    this.audio.pause();
    this.audio.removeAttribute('src');
    this.audio.load();
  }

  private recordPlay(track: PlaybackTrack): void {
    const next: PlayedTrack = {
      id: track.id,
      title: track.title,
      albumId: track.albumId,
      albumTitle: track.albumTitle,
      artists: track.artists,
      playedAt: new Date().toISOString(),
    };

    const history = this.playHistorySg().filter((entry) => entry.id !== track.id);
    this.playHistorySg.set([next, ...history].slice(0, 20));
  }

  private loadUserRatings(userId: string): void {
    this.ratingService.getRatingsByUser(userId, 1, 200).subscribe({
      next: (response) => {
        const ratingsMap: Record<string, number> = {};
        const idsMap: Record<string, string> = {};

        for (const rating of response.items ?? []) {
          ratingsMap[rating.song_id] = rating.value;
          idsMap[rating.song_id] = rating.id;
        }

        this.ratingsBySongSg.set(ratingsMap);
        this.ratingIdsBySongSg.set(idsMap);
      },
      error: () => {
        this.ratingsBySongSg.set({});
        this.ratingIdsBySongSg.set({});
      },
    });
  }

  private syncSongRatingFromApi(songId: string, fallbackValue: number): void {
    const userId = this.authService.currentUserIdSg();
    if (!userId) {
      this.upsertLocalRating(songId, fallbackValue);
      return;
    }

    this.ratingService.getRatingsBySong(songId).subscribe({
      next: (response) => {
        const ownRating = (response.items ?? []).find((entry) => entry.user_id === userId);
        if (!ownRating) {
          this.upsertLocalRating(songId, fallbackValue);
          return;
        }
        this.upsertLocalRating(songId, ownRating.value, ownRating.id);
      },
      error: () => {
        this.upsertLocalRating(songId, fallbackValue);
      },
    });
  }

  private upsertLocalRating(songId: string, value: number, ratingId?: string): void {
    this.ratingsBySongSg.set({
      ...this.ratingsBySongSg(),
      [songId]: value,
    });

    if (!ratingId) {
      return;
    }

    this.ratingIdsBySongSg.set({
      ...this.ratingIdsBySongSg(),
      [songId]: ratingId,
    });
  }

  private removeLocalRating(songId: string): void {
    const nextRatings = { ...this.ratingsBySongSg() };
    const nextRatingIds = { ...this.ratingIdsBySongSg() };
    delete nextRatings[songId];
    delete nextRatingIds[songId];
    this.ratingsBySongSg.set(nextRatings);
    this.ratingIdsBySongSg.set(nextRatingIds);
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

  private setupMediaSessionHandlers(): void {
    const mediaSession = this.getMediaSession();
    if (!mediaSession) {
      return;
    }

    const safeSet = (
      action: 'play' | 'pause' | 'previoustrack' | 'nexttrack' | 'stop',
      handler: (() => void) | null
    ) => {
      try {
        mediaSession.setActionHandler(action, handler);
      } catch {
        // Some browsers/platforms may not support all actions.
      }
    };

    safeSet('play', () => {
      if (!this.isPlayingSg()) {
        this.togglePlayPause();
      }
    });
    safeSet('pause', () => {
      if (this.isPlayingSg()) {
        this.togglePlayPause();
      }
    });
    safeSet('previoustrack', () => this.previous());
    safeSet('nexttrack', () => this.next());
    safeSet('stop', () => {
      if (this.isPlayingSg()) {
        this.togglePlayPause();
      }
    });
  }

  private syncMediaSessionMetadata(): void {
    const mediaSession = this.getMediaSession();
    if (!mediaSession) {
      return;
    }

    const currentTrack = this.currentTrackSg();
    if (!currentTrack) {
      mediaSession.playbackState = 'none';
      return;
    }

    mediaSession.playbackState = this.isPlayingSg() ? 'playing' : 'paused';

    if (typeof MediaMetadata !== 'undefined') {
      mediaSession.metadata = new MediaMetadata({
        title: currentTrack.title,
        artist: currentTrack.artists.map((artist) => artist.name).join(', '),
        album: currentTrack.albumTitle,
      });
    }
  }

  private getMediaSession(): MediaSession | null {
    if (typeof navigator === 'undefined' || !('mediaSession' in navigator)) {
      return null;
    }

    return navigator.mediaSession;
  }

  private restoreAudioState(): void {
    const saved = this.safeReadStorage<PersistedAudioState>(PLAYBACK_AUDIO_STORAGE_KEY);
    if (!saved) {
      return;
    }

    const clampedVolume = Math.max(0, Math.min(1, Number(saved.volume)));
    this.volumeSg.set(Number.isFinite(clampedVolume) ? clampedVolume : DEFAULT_VOLUME);
    this.mutedSg.set(Boolean(saved.muted));
    if (this.volumeSg() > 0) {
      this.lastVolumeBeforeMute = this.volumeSg();
    }
  }

  private persistAudioState(): void {
    this.safeWriteStorage<PersistedAudioState>(PLAYBACK_AUDIO_STORAGE_KEY, {
      volume: this.volumeSg(),
      muted: this.mutedSg(),
    });
  }

  private restorePlaybackState(): void {
    const saved = this.safeReadStorage<PersistedPlaybackState>(PLAYBACK_STATE_STORAGE_KEY);
    if (!saved || !Array.isArray(saved.queue) || saved.queue.length === 0) {
      return;
    }

    const queue = saved.queue.filter((track): track is PlaybackTrack =>
      Boolean(track?.id && track?.title && track?.albumId && track?.albumTitle)
    );
    if (queue.length === 0) {
      return;
    }

    const index = Number.isInteger(saved.currentIndex) ? saved.currentIndex : 0;
    const safeIndex = Math.max(0, Math.min(queue.length - 1, index));
    const currentTrack = queue[safeIndex];

    this.queueSg.set(queue);
    this.currentIndexSg.set(safeIndex);
    this.currentTrackSg.set(currentTrack);
    this.currentTimeSg.set(0);
    this.durationSg.set(currentTrack.lengthSeconds ?? 0);
    this.isPlayingSg.set(false);

    if (saved.currentAlbum?.id && saved.currentAlbum?.title) {
      this.currentAlbumSg.set({
        id: saved.currentAlbum.id,
        title: saved.currentAlbum.title,
        artists: saved.currentAlbum.artists ?? [],
        genres: [],
        songs: [],
        releaseDate: '',
      });
      return;
    }

    this.currentAlbumSg.set({
      id: currentTrack.albumId,
      title: currentTrack.albumTitle,
      artists: currentTrack.artists,
      genres: [],
      songs: [],
      releaseDate: '',
    });
  }

  private persistPlaybackState(): void {
    this.safeWriteStorage<PersistedPlaybackState>(PLAYBACK_STATE_STORAGE_KEY, {
      queue: this.queueSg(),
      currentIndex: this.currentIndexSg(),
      currentAlbum: this.currentAlbumSg()
        ? {
            id: this.currentAlbumSg()!.id,
            title: this.currentAlbumSg()!.title,
            artists: this.currentAlbumSg()!.artists ?? [],
          }
        : null,
    });
  }

  private safeReadStorage<T>(key: string): T | null {
    try {
      const raw = localStorage.getItem(key);
      if (!raw) {
        return null;
      }
      return JSON.parse(raw) as T;
    } catch {
      return null;
    }
  }

  private safeWriteStorage<T>(key: string, value: T): void {
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch {
      // Ignore storage write errors in private mode/quota edge cases.
    }
  }
}
