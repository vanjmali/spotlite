import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-error',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './error.html',
  styleUrls: ['./error.scss'],
})
export class ErrorComponent {
  messageSg = input<string>('', { alias: 'message' });
}
