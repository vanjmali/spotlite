import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Rating {
  id: string;
  song_id: string;
  user_id: string;
  username: string;
  value: number;
  created_at: string;
  is_edited: boolean;
}

export interface SongRatingSummary {
  avg: number;
  count: number;
}

export interface RatingsResponse {
  items: Rating[];
  nextCursor?: string;
}

export interface PaginatedRatingsResponse {
  items: Rating[];
  total: number;
  page: number;
  size: number;
}

@Injectable({
  providedIn: 'root',
})
export class RatingService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = '/api/ratings';

  getRatingsBySong(songId: string): Observable<RatingsResponse> {
    return this.http.get<RatingsResponse>(`${this.apiUrl}/songs/${songId}`);
  }

  getAverageBySong(songId: string): Observable<SongRatingSummary> {
    return this.http.get<SongRatingSummary>(`${this.apiUrl}/songs/${songId}/average`);
  }

  getRatingsByUser(
    userId: string,
    page: number = 1,
    size: number = 50
  ): Observable<PaginatedRatingsResponse> {
    return this.http.get<PaginatedRatingsResponse>(`${this.apiUrl}/users/${userId}`, {
      params: {
        page,
        size,
      },
    });
  }

  createRating(songId: string, value: number): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/`, {
      song_id: songId,
      value,
    });
  }

  updateRating(ratingId: string, value: number): Observable<Rating> {
    return this.http.patch<Rating>(`${this.apiUrl}/${ratingId}`, { value });
  }

  deleteRating(ratingId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${ratingId}`);
  }
}
