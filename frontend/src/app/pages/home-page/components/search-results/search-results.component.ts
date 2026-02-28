import { CommonModule } from '@angular/common';
import { Component, DestroyRef, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { firstValueFrom } from 'rxjs';
import { Album, AlbumService } from '@app/services/album.service';
import { ContentSearchService, GlobalSearchResponse } from '@app/services/content-search.service';
import { Artist } from '@app/services/artist.service';
import { debounceTime, distinctUntilChanged, map, of, switchMap, catchError } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { SongTrailingMetaComponent } from '@app/shared/components/song-trailing-meta/song-trailing-meta';
import { Song } from '@app/services/song.service';
import { AlbumCardComponent } from '@app/shared/components/album-card/album-card.component';
import { PlaybackService } from '@app/services/playback.service';

@Component({
  selector: 'app-search-results',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatIconModule,
    CoverArtComponent,
    SongTrailingMetaComponent,
    AlbumCardComponent,
  ],
  templateUrl: './search-results.component.html',
  styleUrl: './search-results.component.scss',
})
export class SearchResultsComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly searchService = inject(ContentSearchService);
  private readonly albumService = inject(AlbumService);
  readonly playback = inject(PlaybackService);

  readonly querySg = signal('');
  readonly isLoadingSg = signal(false);
  readonly resultsSg = signal<GlobalSearchResponse>({
    songs: [],
    artists: [],
    albums: [],
    genres: [],
  });

  constructor() {
    this.route.queryParamMap
      .pipe(
        map((params) => params.get('q')?.trim() ?? ''),
        debounceTime(150),
        distinctUntilChanged(),
        switchMap((query) => {
          this.querySg.set(query);
          if (query.length < 2) {
            this.resultsSg.set({ songs: [], artists: [], albums: [], genres: [] });
            return of(null);
          }

          this.isLoadingSg.set(true);
          return this.searchService.search(query).pipe(
            catchError(() =>
              of<GlobalSearchResponse>({
                songs: [],
                artists: [],
                albums: [],
                genres: [],
              })
            )
          );
        }),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe((result) => {
        if (result) {
          this.resultsSg.set(result);
        }
        this.isLoadingSg.set(false);
      });
  }

  artistNames(artists: Artist[] | undefined): string {
    if (!artists || artists.length === 0) {
      return 'Unknown artist';
    }

    return artists.map((artist) => artist.name).join(', ');
  }

  shortDescription(value: string, maxLength: number = 120): string {
    const normalized = (value || '').trim();
    if (normalized.length <= maxLength) {
      return normalized;
    }

    return `${normalized.slice(0, maxLength).trimEnd()}...`;
  }

  async openSong(song: Song): Promise<void> {
    const albumId = await this.resolveSongAlbumId(song);
    if (!albumId) {
      return;
    }

    await this.router.navigate(['/album', albumId]);
  }

  toggleAlbumPlayback(album: Album, event: Event): void {
    event.preventDefault();
    event.stopPropagation();

    const currentTrack = this.playback.currentTrackSg();
    if (currentTrack?.albumId === album.id) {
      this.playback.togglePlayPause();
      return;
    }

    this.playback.playAlbum(album);
  }

  isAlbumPlaying(album: Album): boolean {
    const currentTrack = this.playback.currentTrackSg();
    return currentTrack?.albumId === album.id && this.playback.isPlayingSg();
  }

  async toggleSongPlayback(song: Song, event: Event): Promise<void> {
    event.preventDefault();
    event.stopPropagation();

    const currentTrack = this.playback.currentTrackSg();
    if (currentTrack?.id === song.id) {
      this.playback.togglePlayPause();
      return;
    }

    const album = await this.resolveSongAlbum(song);
    if (!album) {
      return;
    }

    this.playback.playSingleSong(song, album);
  }

  isSongPlaying(songId: string): boolean {
    return this.playback.currentTrackSg()?.id === songId && this.playback.isPlayingSg();
  }

  private async resolveSongAlbumId(song: Song): Promise<string | null> {
    const album = await this.resolveSongAlbum(song);
    return album?.id ?? null;
  }

  private async resolveSongAlbum(song: Song): Promise<Album | null> {
    const fromSearch = this.resultsSg().albums.find((album) =>
      (album.songs ?? []).some((entry) => entry.id === song.id)
    );
    if (fromSearch) {
      return fromSearch;
    }

    const candidate = this.tryReadAlbumId(song);
    if (candidate && candidate.trim().length > 0) {
      try {
        return await firstValueFrom(this.albumService.getAlbumById(candidate));
      } catch {
        return null;
      }
    }

    const firstArtistId = song.artists?.[0]?.id;
    if (!firstArtistId) {
      return null;
    }

    try {
      const albums = await firstValueFrom(
        this.albumService.getAlbums(1, 100, { artist_id: firstArtistId })
      );
      return (
        (albums.items ?? []).find((album) =>
          (album.songs ?? []).some(
            (entry) =>
              entry.id === song.id || entry.title.toLowerCase() === song.title.toLowerCase()
          )
        ) ?? null
      );
    } catch {
      return null;
    }
  }

  private tryReadAlbumId(song: Song): string | null {
    const maybe = song as Song & {
      albumId?: string;
      album_id?: string;
      album?: { id?: string };
    };

    return maybe.albumId ?? maybe.album_id ?? maybe.album?.id ?? null;
  }
}
