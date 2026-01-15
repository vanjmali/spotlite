import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WidgetComponent } from '@app/shared/components/widget';

@Component({
  selector: 'app-albums-management',
  standalone: true,
  imports: [CommonModule, WidgetComponent],
  templateUrl: './albums-management.component.html',
  styleUrl: './albums-management.component.scss',
})
export class AlbumsManagementComponent {}
