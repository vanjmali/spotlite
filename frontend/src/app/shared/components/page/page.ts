import { Component, input, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HeaderComponent } from '../header/header';
import { AuthService } from '@app/services/auth.service';

@Component({
  selector: 'app-page',
  standalone: true,
  imports: [CommonModule, HeaderComponent],
  templateUrl: './page.html',
  styleUrls: ['./page.scss'],
})
export class PageComponent {
  titleSg = input<string | undefined>();
  readonly authService = inject(AuthService);
}
