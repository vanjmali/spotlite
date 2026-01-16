import { Component, OnInit, inject } from '@angular/core';
import { Notification, NotificationService } from '@app/services/notification.service';
import { Observable } from 'rxjs';
import { NotificationCard } from './components/notification-card/notification-card';
import { AsyncPipe } from '@angular/common';
import { AuthService } from '@app/services/auth.service';

@Component({
  selector: 'app-inbox-page',
  imports: [NotificationCard, AsyncPipe],
  templateUrl: './inbox-page.html',
  styleUrl: './inbox-page.scss',
})
export class InboxPage implements OnInit {
  notifications$: Observable<Notification[]>;
  notificationService = inject(NotificationService);
  authService = inject(AuthService);

  constructor() {
    this.notifications$ = this.notificationService.notifications$;
  }

  ngOnInit() {
    this.notificationService.getAllNotifications();
  }
}
