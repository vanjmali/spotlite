import { Component, Input } from '@angular/core';
import { Notification } from '@app/services/notification.service';

@Component({
  selector: 'app-notification-card',
  imports: [],
  templateUrl: './notification-card.html',
  styleUrl: './notification-card.scss',
})
export class NotificationCard {
  @Input({ required: true }) notification!: Notification;

  // Helper logic moved inside the card
  get timeAgo(): string {
    if (!this.notification?.created_at) return '';

    const date = new Date(this.notification.created_at);
    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    let interval = seconds / 31536000;
    if (interval > 1) return Math.floor(interval) + ' years ago';
    interval = seconds / 2592000;
    if (interval > 1) return Math.floor(interval) + ' months ago';
    interval = seconds / 86400;
    if (interval > 1) return Math.floor(interval) + ' days ago';
    interval = seconds / 3600;
    if (interval > 1) return Math.floor(interval) + ' hours ago';
    interval = seconds / 60;
    if (interval > 1) return Math.floor(interval) + ' mins ago';

    return 'Just now';
  }
}
