import { Component, input, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HeaderComponent } from '../header/header';
import { SiteFooterComponent } from '../site-footer';
import { AuthService } from '@app/services/auth.service';

@Component({
  selector: 'app-page',
  standalone: true,
  imports: [CommonModule, HeaderComponent, SiteFooterComponent],
  templateUrl: './page.html',
  styleUrls: ['./page.scss'],
})
export class PageComponent {
  titleSg = input<string | undefined>();
  readonly authService = inject(AuthService);
}
