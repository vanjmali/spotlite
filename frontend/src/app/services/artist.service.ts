import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

// Artist interfaces
export interface Artist {
  id: string;
  name: string;
  genres: string[];
  description: string;
}

export interface CreateArtistDto {
  name: string;
  genres: string[];
  description: string;
}

export interface UpdateArtistDto {
  name?: string;
  genres?: string[];
  description?: string;
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
export class ArtistService {
  private http = inject(HttpClient);
  private apiUrl = '/api/content/artists';

  /**
   * Get all artists with pagination
   */
  getArtists(page: number = 1, size: number = 10): Observable<PaginatedResponse<Artist>> {
    const params = new HttpParams().set('page', page.toString()).set('size', size.toString());
    return this.http.get<PaginatedResponse<Artist>>(this.apiUrl, { params });
  }

  /**
   * Get single artist by ID
   */
  getArtistById(id: string): Observable<Artist> {
    return this.http.get<Artist>(`${this.apiUrl}/${id}`);
  }

  /**
   * Create new artist
   */
  createArtist(artist: CreateArtistDto): Observable<void> {
    return this.http.post<void>(this.apiUrl, artist);
  }

  /**
   * Update artist
   */
  updateArtist(id: string, artist: UpdateArtistDto): Observable<Artist> {
    return this.http.patch<Artist>(`${this.apiUrl}/${id}`, artist);
  }

  /**
   * Delete artist
   */
  deleteArtist(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}
