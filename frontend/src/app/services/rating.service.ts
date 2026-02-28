import { inject, Injectable, signal } from '@angular/core';
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
  next_cursor?: string;
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
  readonly summariesBySongSg = signal<Record<string, SongRatingSummary>>({});
  private readonly loadingSummaryIds = new Set<string>();
  private readonly queuedSummaryIds = new Set<string>();

  getRatingsBySong(songId: string): Observable<RatingsResponse> {
    return this.http.get<RatingsResponse>(`${this.apiUrl}/songs/${songId}`);
  }

  getAverageBySong(songId: string): Observable<SongRatingSummary> {
    return this.http.get<SongRatingSummary>(`${this.apiUrl}/songs/${songId}/average`);
  }

  getLiveSummary(songId: string): SongRatingSummary | null {
    return this.summariesBySongSg()[songId] ?? null;
  }

  refreshSongSummary(songId: string): void {
    if (!songId) {
      return;
    }

    if (this.loadingSummaryIds.has(songId)) {
      console.debug('[rating] summary refresh queued', { songId });
      this.queuedSummaryIds.add(songId);
      return;
    }

    console.debug('[rating] summary refresh start', { songId });
    this.loadingSummaryIds.add(songId);
    const finish = () => {
      this.loadingSummaryIds.delete(songId);
      if (this.queuedSummaryIds.has(songId)) {
        this.queuedSummaryIds.delete(songId);
        this.refreshSongSummary(songId);
      }
    };

    this.getAverageBySong(songId).subscribe({
      next: (summary) => {
        console.debug('[rating] summary refresh success', { songId, summary });
        this.summariesBySongSg.set({
          ...this.summariesBySongSg(),
          [songId]: summary,
        });
      },
      error: (error) => {
        console.debug('[rating] summary refresh failed', { songId, error });
        finish();
      },
      complete: () => finish(),
    });
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
