import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';

type DialogSize = 'small' | 'normal' | 'large';

@Component({
  selector: 'app-dialog',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './dialog.html',
  styleUrls: ['./dialog.scss'],
})
export class DialogComponent {
  readonly sizeSg = input<DialogSize>('normal', { alias: 'size' });
  readonly titleSg = input<string>('', { alias: 'title' });
  readonly closed = output<void>();

  close(): void {
    this.closed.emit();
  }
}
