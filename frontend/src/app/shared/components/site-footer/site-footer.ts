import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { BUILD_YEAR } from '@app/shared/build-info';

@Component({
  selector: 'app-site-footer',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './site-footer.html',
  styleUrls: ['./site-footer.scss'],
})
export class SiteFooterComponent {
  readonly year = BUILD_YEAR;
}
