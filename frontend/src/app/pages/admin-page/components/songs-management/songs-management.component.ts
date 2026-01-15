import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WidgetComponent } from '@app/shared/components/widget';

@Component({
  selector: 'app-songs-management',
  standalone: true,
  imports: [CommonModule, WidgetComponent],
  templateUrl: './songs-management.component.html',
  styleUrl: './songs-management.component.scss',
})
export class SongsManagementComponent {}
