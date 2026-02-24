import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { BUILD_YEAR } from '@app/shared/build-info';

@Component({
  selector: 'app-control-bar',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './control-bar.html',
  styleUrls: ['./control-bar.scss'],
})
export class ControlBarComponent {
  readonly year = BUILD_YEAR;
}
