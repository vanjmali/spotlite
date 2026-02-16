import { Component, input, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HeaderComponent } from '../header/header';
import { SiteFooterComponent } from '../site-footer';
import { AuthService } from '@app/services/auth.service';
import { PlaybackBarComponent } from '../playback-bar/playback-bar';
import { NowPlayingSidebarComponent } from '../now-playing-sidebar/now-playing-sidebar';

@Component({
  selector: 'app-page',
  standalone: true,
  imports: [
    CommonModule,
    HeaderComponent,
    SiteFooterComponent,
    PlaybackBarComponent,
    NowPlayingSidebarComponent,
  ],
  templateUrl: './page.html',
  styleUrls: ['./page.scss'],
})
export class PageComponent {
  titleSg = input<string | undefined>();
  readonly authService = inject(AuthService);
}
