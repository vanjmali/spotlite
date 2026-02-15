import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Artist } from './artist.service';
import { Genre } from './genre.service';

// Song interfaces
export interface Song {
  id: string;
  title: string;
  genres: Genre[];
  lengthSeconds: number;
  artists: Artist[];
}

export interface CreateSongDto {
  title: string;
  album_id: string;
  genre_ids: string[];
  artist_ids: string[];
}

export interface UpdateSongDto {
  title?: string;
  genre_ids?: string[];
  artist_ids?: string[];
}

// Paginated response
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  size: number;
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
      genre_id?: string;
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
    if (filters?.genre_id) {
      params = params.set('genre_id', filters.genre_id);
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
   * Create new song with uploaded audio file
   */
  createSongWithAudio(song: CreateSongDto, file: File): Observable<Song> {
    const formData = new FormData();
    formData.append('meta', JSON.stringify(song));
    formData.append('file', file);
    return this.http.post<Song>(this.apiUrl, formData);
  }

  /**
   * Update song
   */
  updateSong(id: string, song: UpdateSongDto): Observable<Song> {
    return this.http.patch<Song>(`${this.apiUrl}/${id}`, song);
  }

  /**
   * Upload/replace song audio file
   */
  uploadSongAudio(id: string, file: File): Observable<Song> {
    const formData = new FormData();
    formData.append('file', file);
    return this.http.put<Song>(`${this.apiUrl}/${id}/audio`, formData);
  }

  /**
   * Delete song
   */
  deleteSong(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}
