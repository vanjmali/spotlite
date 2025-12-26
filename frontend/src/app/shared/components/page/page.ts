import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-page',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './page.html',
  styleUrls: ['./page.scss'],
})
export class PageComponent {
  titleSg = input<string | undefined>();
}
