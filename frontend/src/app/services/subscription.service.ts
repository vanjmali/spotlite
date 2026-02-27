import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Subject, tap } from 'rxjs';

export type SubscriptionType = 'ARTIST' | 'GENRE';

export interface Subscription {
  id: string;
  subscriber_id: string;
  entity_id: string;
  sub_type: SubscriptionType;
  subscribed_at: string;
  entity_name: string;
}

export interface PaginatedSubscriptionsResponse {
  items: Subscription[];
  total: number;
  page: number;
  size: number;
}

export interface SubscriberCountResponse {
  sub_count: number;
}

@Injectable({
  providedIn: 'root',
})
export class SubscriptionService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = '/api/subscriptions';
  private readonly changedSubject = new Subject<void>();
  readonly changes$ = this.changedSubject.asObservable();

  getMySubscriptions(
    page: number = 1,
    size: number = 100
  ): Observable<PaginatedSubscriptionsResponse> {
    return this.http.get<PaginatedSubscriptionsResponse>(`${this.apiUrl}/`, {
      params: {
        page,
        size,
      },
    });
  }

  isSubscribed(entityId: string): Observable<unknown> {
    return this.http.get(`${this.apiUrl}/${entityId}`);
  }

  subscribe(entityId: string, subType: SubscriptionType): Observable<void> {
    return this.http
      .post<void>(`${this.apiUrl}/`, {
        entity_id: entityId,
        sub_type: subType,
      })
      .pipe(tap(() => this.notifyChanged()));
  }

  unsubscribe(entityId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${entityId}`).pipe(tap(() => this.notifyChanged()));
  }

  getSubscriberCount(entityId: string): Observable<SubscriberCountResponse> {
    return this.http.get<SubscriberCountResponse>(`${this.apiUrl}/count/${entityId}`);
  }

  notifyChanged(): void {
    this.changedSubject.next();
  }
}
