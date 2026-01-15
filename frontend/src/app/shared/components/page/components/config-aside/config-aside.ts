import { Component, input, output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-config-aside',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './config-aside.html',
  styleUrls: ['./config-aside.scss'],
})
export class ConfigAsideComponent {
  readonly isExpandedSg = input<boolean>(false);
  readonly configOptionClicked = output<string>();

  readonly expandedSg = signal(false);

  readonly configOptions = [
    { key: 'artists', icon: 'person', label: 'Manage Artists' },
    { key: 'albums', icon: 'album', label: 'Manage Albums' },
    { key: 'songs', icon: 'music_note', label: 'Manage Songs' },
  ];

  toggleExpand(): void {
    this.expandedSg.update((isExpanded) => !isExpanded);
  }

  selectOption(optionKey: string): void {
    this.configOptionClicked.emit(optionKey);
  }
}
