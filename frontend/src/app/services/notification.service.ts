import { HttpClient } from '@angular/common/http';
import { inject, Injectable, NgZone } from '@angular/core';
import { EventSourcePolyfill } from 'event-source-polyfill';
import { BehaviorSubject, Subject } from 'rxjs';
import { environment } from 'src/environments/environment';

export interface Notification {
  notification_id: string;
  title: string;
  created_at: string;
  entity_id: string;
  notification_type: string;
  entity_name: string;
  is_new: boolean;
}

interface EventSourcePolyfillError {
  status?: number;
  message?: string;
}

@Injectable({ providedIn: 'root' })
export class NotificationService {
  private readonly apiUrl = `${environment.apiBaseUrl}/notifications`;

  // State management
  private notificationsSubject = new BehaviorSubject<Notification[]>([]);
  public notifications$ = this.notificationsSubject.asObservable();
  private incomingNotificationSubject = new Subject<Notification>();
  public incomingNotification$ = this.incomingNotificationSubject.asObservable();
  private dismissedIds = new Set<string>();

  private http = inject(HttpClient);
  private zone = inject(NgZone);

  private eventSource: EventSourcePolyfill | null = null;

  openSSEConnection(token: string) {
    if (this.eventSource) {
      return;
    }

    this.eventSource = new EventSourcePolyfill(`${this.apiUrl}/stream`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    this.eventSource.onmessage = (event) => {
      const notificationString = event.data;

      const notificationObject = this.transformToNotificationObject(notificationString);

      this.zone.run(() => {
        const current = this.notificationsSubject.value.filter(
          (item) => item.notification_id !== notificationObject.notification_id
        );
        this.notificationsSubject.next([notificationObject, ...current]);
        this.incomingNotificationSubject.next(notificationObject);
      });
    };

    // this.eventSource.addEventListener('ping', () => {
    //   // We received the ping successfully!
    //   // We don't need to do anything, but this proves the connection is healthy.
    //   console.debug('Heartbeat received');
    // });

    this.eventSource.onerror = (err: unknown) => {
      const error = err as EventSourcePolyfillError;

      this.zone.run(() => {
        if (error.status === 401) {
          console.error('SSE Token expired');
          this.closeConnection();
        }
      });
    };
  }

  closeConnection() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
  }

  getAllNotifications() {
    this.http.get<Notification[]>(this.apiUrl).subscribe({
      next: (data) => {
        const normalized = (data || [])
          .filter((item) => !this.dismissedIds.has(item.notification_id))
          .map((item) => ({ ...item, is_new: item.is_new === true }))
          .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
        this.notificationsSubject.next(normalized);
      },
      error: (err) => console.error('Error fetching notifications:', err),
    });
  }

  dismissNotification(id: string): void {
    this.dismissedIds.add(id);
    this.notificationsSubject.next(
      this.notificationsSubject.value.filter((item) => item.notification_id !== id)
    );
  }

  markAllAsSeen(): void {
    this.notificationsSubject.next(
      this.notificationsSubject.value.map((item) => ({ ...item, is_new: false }))
    );
  }

  transformToNotificationObject(data: string): Notification {
    const notificationJSONObject = JSON.parse(data);

    return {
      notification_id: notificationJSONObject.notification_id,
      title: 'Notification',
      entity_id: notificationJSONObject.entity_id,
      entity_name: notificationJSONObject.entity_name,
      notification_type: notificationJSONObject.notification_type,
      created_at: notificationJSONObject.created_at,
      is_new: true,
    };
  }
}
