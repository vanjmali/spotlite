import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

export interface Notification {
  notification_id: string;
  title: string;
  message: string;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class NotificationService {
  private apiUrl = 'http://localhost:3000/api/notifications';

  // State management
  private notificationsSubject = new BehaviorSubject<Notification[]>([]);
  public notifications$ = this.notificationsSubject.asObservable();
  private http = inject(HttpClient);

  getAllNotifications() {
    this.http.get<Notification[]>(this.apiUrl).subscribe({
      next: (data) => {
        this.notificationsSubject.next(data);
      },
      error: (err) => console.error('Error fetching notifications:', err),
    });
  }
}
