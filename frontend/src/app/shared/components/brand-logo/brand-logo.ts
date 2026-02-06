import { CommonModule } from '@angular/common';
import { Component, input } from '@angular/core';

@Component({
  selector: 'app-brand-logo',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './brand-logo.html',
  styleUrls: ['./brand-logo.scss'],
})
export class BrandLogo {
  public readonly sizeSg = input<number>(50, { alias: 'size' });
}
