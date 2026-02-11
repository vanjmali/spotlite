import { Injectable } from '@angular/core';
import { map, Observable } from 'rxjs';
import { ArtistService } from '@app/services/artist.service';
import { GenreService } from '@app/services/genre.service';
import { AlbumService } from '@app/services/album.service';
import type { SelectOption } from '@app/shared/components/input';

@Injectable({
  providedIn: 'root',
})
export class OptionsService {
  constructor(
    private readonly artistService: ArtistService,
    private readonly genreService: GenreService,
    private readonly albumService: AlbumService
  ) {}

  loadArtists(limit: number = 100): Observable<SelectOption[]> {
    return this.artistService.getArtists(1, limit).pipe(
      map((response) =>
        response.items.map((artist) => ({
          label: artist.name,
          value: artist.id,
        }))
      )
    );
  }

  loadGenres(limit: number = 200): Observable<SelectOption[]> {
    return this.genreService.getGenres(1, limit).pipe(
      map((response) =>
        response.items.map((genre) => ({
          label: genre.name,
          value: genre.id,
        }))
      )
    );
  }

  loadAlbums(limit: number = 200): Observable<SelectOption[]> {
    return this.albumService.getAlbums(1, limit).pipe(
      map((response) =>
        response.items.map((album) => ({
          label: album.title,
          value: album.id,
        }))
      )
    );
  }
}
