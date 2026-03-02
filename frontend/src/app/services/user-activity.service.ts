import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export type ActivityType =
  | 'LISTEN'
  | 'SUBSCRIPTION_CREATED'
  | 'SUBSCRIPTION_DELETED'
  | 'RATING_CREATED'
  | 'RATING_UPDATED';

export interface UserActivity {
  event_id: string;
  user_id: string;
  type: ActivityType;
  entity_id?: string;
  entity_name?: string;
  entity_type?: string;
  rating_value?: number;
  occurred_at: string;
  created_at: string;
}

export interface UserActivityResponse {
  items: UserActivity[];
  total: number;
  page: number;
  size: number;
}

@Injectable({
  providedIn: 'root',
})
export class UserActivityService {
  private readonly http = inject(HttpClient);
  private readonly apiBase = `${environment.apiBaseUrl}/users`;

  getMyActivities(page: number = 1, size: number = 20): Observable<UserActivityResponse> {
    return this.http.get<UserActivityResponse>(`${this.apiBase}/activities`, {
      params: { page, size },
    });
  }
}
