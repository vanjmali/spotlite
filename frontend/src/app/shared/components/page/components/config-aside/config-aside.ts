import { Component, input, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-config-aside',
  standalone: true,
  imports: [CommonModule, MatIconModule, RouterModule],
  templateUrl: './config-aside.html',
  styleUrls: ['./config-aside.scss'],
})
export class ConfigAsideComponent {
  readonly isExpandedSg = input<boolean>(false);

  readonly expandedSg = signal(false);

  readonly configOptions = [
    { key: 'artists', icon: 'person', label: 'Manage Artists' },
    { key: 'albums', icon: 'album', label: 'Manage Albums' },
    { key: 'songs', icon: 'music_note', label: 'Manage Songs' },
  ];

  toggleExpand(): void {
    this.expandedSg.update((isExpanded) => !isExpanded);
  }
}
