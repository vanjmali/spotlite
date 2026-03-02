import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface SongRecommendation {
  song_id: string;
  title: string;
  duration: number;
  rating: number;
  artists: string[];
}

@Injectable({
  providedIn: 'root',
})
export class RecommendationService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = '/api/recommendations';

  getSubscriptionRecommendations(): Observable<SongRecommendation[]> {
    return this.http.get<SongRecommendation[]>(`${this.apiUrl}/subscriptions`);
  }

  getLikeBasedRecommendations(): Observable<SongRecommendation[]> {
    return this.http.get<SongRecommendation[]>(`${this.apiUrl}/likes`);
  }
}
