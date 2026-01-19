import { HttpClient } from '@angular/common/http';
import { inject, Injectable, NgZone } from '@angular/core';
import { EventSourcePolyfill } from 'event-source-polyfill';
import { BehaviorSubject } from 'rxjs';
import { environment } from 'src/environments/environment';

export interface Notification {
  notification_id: string;
  title: string;
  message: string;
  created_at: string;
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
        const current = this.notificationsSubject.value;
        this.notificationsSubject.next([notificationObject, ...current]);
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
        this.notificationsSubject.next(data);
      },
      error: (err) => console.error('Error fetching notifications:', err),
    });
  }

  sendTestNotification() {
    return this.http.post(`${this.apiUrl}`, {});
  }

  transformToNotificationObject(data: string): Notification {
    const notificationJSONObject = JSON.parse(data);

    return {
      notification_id: notificationJSONObject.notification_id,
      title: 'Notification',
      message: notificationJSONObject.message,
      created_at: notificationJSONObject.created_at,
      is_new: true,
    };
  }
}
