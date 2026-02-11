import { Injectable, inject } from '@angular/core';
import { map, Observable, shareReplay } from 'rxjs';
import { ArtistService } from '@app/services/artist.service';
import { GenreService } from '@app/services/genre.service';
import { AlbumService } from '@app/services/album.service';
import type { SelectOption } from '@app/shared/components/input';

@Injectable({
  providedIn: 'root',
})
export class OptionsService {
  private readonly artistService = inject(ArtistService);
  private readonly genreService = inject(GenreService);
  private readonly albumService = inject(AlbumService);

  private readonly artistsCache = new Map<number, Observable<SelectOption[]>>();
  private readonly genresCache = new Map<number, Observable<SelectOption[]>>();
  private readonly albumsCache = new Map<number, Observable<SelectOption[]>>();

  loadArtists(limit: number = 100, refresh: boolean = false): Observable<SelectOption[]> {
    if (!refresh && this.artistsCache.has(limit)) {
      return this.artistsCache.get(limit)!;
    }

    const request$ = this.artistService.getArtists(1, limit).pipe(
      map((response) =>
        response.items.map((artist) => ({
          label: artist.name,
          value: artist.id,
        }))
      ),
      shareReplay({ bufferSize: 1, refCount: false })
    );

    this.artistsCache.set(limit, request$);
    return request$;
  }

  loadGenres(limit: number = 200, refresh: boolean = false): Observable<SelectOption[]> {
    if (!refresh && this.genresCache.has(limit)) {
      return this.genresCache.get(limit)!;
    }

    const request$ = this.genreService.getGenres(1, limit).pipe(
      map((response) =>
        response.items.map((genre) => ({
          label: genre.name,
          value: genre.id,
        }))
      ),
      shareReplay({ bufferSize: 1, refCount: false })
    );

    this.genresCache.set(limit, request$);
    return request$;
  }

  loadAlbums(limit: number = 200, refresh: boolean = false): Observable<SelectOption[]> {
    if (!refresh && this.albumsCache.has(limit)) {
      return this.albumsCache.get(limit)!;
    }

    const request$ = this.albumService.getAlbums(1, limit).pipe(
      map((response) =>
        response.items.map((album) => ({
          label: album.title,
          value: album.id,
        }))
      ),
      shareReplay({ bufferSize: 1, refCount: false })
    );

    this.albumsCache.set(limit, request$);
    return request$;
  }

  invalidateArtists(): void {
    this.artistsCache.clear();
  }

  invalidateGenres(): void {
    this.genresCache.clear();
  }

  invalidateAlbums(): void {
    this.albumsCache.clear();
  }

  invalidateAll(): void {
    this.invalidateArtists();
    this.invalidateGenres();
    this.invalidateAlbums();
  }
}
