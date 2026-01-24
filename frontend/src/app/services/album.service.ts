import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Artist } from './artist.service';
import { Song } from './song.service';
import { Genre } from './genre.service';

// Album interfaces
export interface Album {
  id: string;
  title: string;
  releaseDate: string;
  genres: Genre[];
  songs: Song[];
  artists: Artist[];
}

export interface CreateAlbumDto {
  title: string;
  release_date: string;
  genre_ids: string[];
  song_ids: string[];
  artist_ids: string[];
}

export interface UpdateAlbumDto {
  title?: string;
  release_date?: string;
  genre_ids?: string[];
  artist_ids?: string[];
}

export interface AddAlbumSongsDto {
  ids: string[];
}

// Paginated response
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

@Injectable({
  providedIn: 'root',
})
export class AlbumService {
  private http = inject(HttpClient);
  private apiUrl = '/api/content/albums';

  /**
   * Get all albums with pagination and filtering
   */
  getAlbums(
    page: number = 1,
    size: number = 10,
    filters?: {
      title?: string;
      genres?: string;
      genre_id?: string;
      artist_id?: string;
    }
  ): Observable<PaginatedResponse<Album>> {
    let params = new HttpParams().set('page', page.toString()).set('size', size.toString());

    if (filters?.title) {
      params = params.set('title', filters.title);
    }
    if (filters?.genres) {
      params = params.set('genres', filters.genres);
    }
    if (filters?.genre_id) {
      params = params.set('genre_id', filters.genre_id);
    }
    if (filters?.artist_id) {
      params = params.set('artist_id', filters.artist_id);
    }

    return this.http.get<PaginatedResponse<Album>>(this.apiUrl, { params });
  }

  /**
   * Get single album by ID
   */
  getAlbumById(id: string): Observable<Album> {
    return this.http.get<Album>(`${this.apiUrl}/${id}`);
  }

  /**
   * Create new album
   */
  createAlbum(dto: CreateAlbumDto): Observable<void> {
    return this.http.post<void>(this.apiUrl, dto);
  }

  /**
   * Update existing album
   */
  updateAlbum(id: string, dto: UpdateAlbumDto): Observable<void> {
    return this.http.patch<void>(`${this.apiUrl}/${id}`, dto);
  }

  /**
   * Delete album
   */
  deleteAlbum(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }

  /**
   * Get songs in album
   */
  getAlbumSongs(id: string): Observable<Song[]> {
    return this.http.get<Song[]>(`${this.apiUrl}/${id}/songs`);
  }

  /**
   * Add songs to album
   */
  addAlbumSongs(id: string, dto: AddAlbumSongsDto): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/${id}/songs`, dto);
  }

  /**
   * Remove song from album
   */
  removeAlbumSong(albumId: string, songId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${albumId}/songs/${songId}`);
  }
}
