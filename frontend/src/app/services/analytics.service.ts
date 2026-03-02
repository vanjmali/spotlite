import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface TopArtistAnalytics {
  artist_id: string;
  play_count: number;
}

export interface UserAnalyticsResponse {
  total_songs_played: number;
  average_rating: number;
  songs_by_genre: Record<string, number>;
  top_artists: TopArtistAnalytics[];
  subscribed_artists_count: number;
}

@Injectable({
  providedIn: 'root',
})
export class AnalyticsService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = '/api/analytics';

  getUserAnalytics(): Observable<UserAnalyticsResponse> {
    return this.http.get<UserAnalyticsResponse>(`${this.apiUrl}/`);
  }
}
