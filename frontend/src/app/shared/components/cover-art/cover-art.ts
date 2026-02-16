import { CommonModule } from '@angular/common';
import { Component, computed, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { angleFromName, colorFromName, iconFromName, initialsFromName } from '@app/shared/utils/display';

@Component({
  selector: 'app-cover-art',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './cover-art.html',
  styleUrl: './cover-art.scss',
})
export class CoverArtComponent {
  readonly nameSg = input.required<string>({ alias: 'name' });
  readonly squareSg = input(false, { alias: 'square' });
  readonly patternSg = input(true, { alias: 'pattern' });
  readonly sizeSg = input<number | null>(null, { alias: 'size' });
  readonly fontSizeSg = input<number | null>(null, { alias: 'fontSize' });

  readonly initialsSg = computed(() => initialsFromName(this.nameSg()));
  readonly gradientSg = computed(() => colorFromName(this.nameSg()));
  readonly iconSg = computed(() => iconFromName(this.nameSg()));
  readonly angleSg = computed(() => `${angleFromName(this.nameSg())}deg`);
  readonly sizeStyleSg = computed(() => {
    const size = this.sizeSg();
    return size ? `${size}px` : null;
  });
  readonly fontSizeStyleSg = computed(() => {
    const size = this.fontSizeSg();
    return size ? `${size}px` : null;
  });
}
