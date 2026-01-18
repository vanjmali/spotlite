import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Artist } from './artist.service';

// Song interfaces
export interface Song {
  id: string;
  title: string;
  genre: string;
  lengthSeconds: number;
  artists: Artist[];
}

export interface CreateSongDto {
  title: string;
  genre: string;
  length_seconds: number;
  artist_ids: string[];
}

export interface UpdateSongDto {
  title?: string;
  genre?: string;
  length_seconds?: number;
  artist_ids?: string[];
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
export class SongService {
  private http = inject(HttpClient);
  private apiUrl = '/api/content/songs';

  /**
   * Get all songs with pagination and filtering
   */
  getSongs(
    page: number = 1,
    size: number = 10,
    filters?: {
      title?: string;
      genre?: string;
      artist_id?: string;
    }
  ): Observable<PaginatedResponse<Song>> {
    let params = new HttpParams().set('page', page.toString()).set('size', size.toString());

    if (filters?.title) {
      params = params.set('title', filters.title);
    }
    if (filters?.genre) {
      params = params.set('genre', filters.genre);
    }
    if (filters?.artist_id) {
      params = params.set('artist_id', filters.artist_id);
    }

    return this.http.get<PaginatedResponse<Song>>(this.apiUrl, { params });
  }

  /**
   * Get single song by ID
   */
  getSongById(id: string): Observable<Song> {
    return this.http.get<Song>(`${this.apiUrl}/${id}`);
  }

  /**
   * Create new song
   */
  createSong(song: CreateSongDto): Observable<void> {
    return this.http.post<void>(this.apiUrl, song);
  }

  /**
   * Update song
   */
  updateSong(id: string, song: UpdateSongDto): Observable<Song> {
    return this.http.patch<Song>(`${this.apiUrl}/${id}`, song);
  }

  /**
   * Delete song
   */
  deleteSong(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}
