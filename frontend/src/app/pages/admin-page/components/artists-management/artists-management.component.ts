import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { WidgetComponent } from '@app/shared/components/widget';

@Component({
  selector: 'app-artists-management',
  standalone: true,
  imports: [CommonModule, WidgetComponent],
  templateUrl: './artists-management.component.html',
  styleUrl: './artists-management.component.scss',
})
export class ArtistsManagementComponent {}
