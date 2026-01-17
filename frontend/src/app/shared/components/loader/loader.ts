import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-loader',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './loader.html',
  styleUrls: ['./loader.scss'],
})
export class LoaderComponent {
  /**
   * Optional title to display above the loader
   */
  public readonly titleSg = input<string | null>(null, { alias: 'title' });

  /**
   * Optional message to display below the loader
   */
  public readonly messageSg = input<string | null>(null, { alias: 'message' });
}
