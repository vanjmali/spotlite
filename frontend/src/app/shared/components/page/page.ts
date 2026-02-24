import { Component, input, inject, signal, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HeaderComponent } from '../header/header';
import { SiteFooterComponent } from '../site-footer';
import { AuthService } from '@app/services/auth.service';
import { PlaybackBarComponent } from '../playback-bar/playback-bar';
import { NowPlayingSidebarComponent } from '../now-playing-sidebar/now-playing-sidebar';

type ResizeSide = 'left' | 'right';

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
export class PageComponent implements OnDestroy {
  private static readonly LEFT_KEY = 'spotlite.page.sidebar.left.width';
  private static readonly RIGHT_KEY = 'spotlite.page.sidebar.right.width';
  private static readonly LEFT_DEFAULT = 280;
  private static readonly RIGHT_DEFAULT = 280;
  private static readonly LEFT_MIN = 220;
  private static readonly RIGHT_MIN = 220;
  private static readonly MAX_WIDTH = 520;

  private activeResize: ResizeSide | null = null;
  private startX = 0;
  private startWidth = 0;

  titleSg = input<string | undefined>();
  showSidebarsSg = input<boolean>(false, { alias: 'showSidebars' });
  readonly authService = inject(AuthService);
  readonly leftSidebarWidthSg = signal(PageComponent.LEFT_DEFAULT);
  readonly rightSidebarWidthSg = signal(PageComponent.RIGHT_DEFAULT);
  readonly isResizingSg = signal(false);

  constructor() {
    this.leftSidebarWidthSg.set(this.readWidth(PageComponent.LEFT_KEY, PageComponent.LEFT_DEFAULT));
    this.rightSidebarWidthSg.set(this.readWidth(PageComponent.RIGHT_KEY, PageComponent.RIGHT_DEFAULT));
  }

  startResize(side: ResizeSide, event: MouseEvent): void {
    event.preventDefault();
    this.activeResize = side;
    this.startX = event.clientX;
    this.startWidth = side === 'left' ? this.leftSidebarWidthSg() : this.rightSidebarWidthSg();
    this.isResizingSg.set(true);

    window.addEventListener('mousemove', this.onMouseMove);
    window.addEventListener('mouseup', this.onMouseUp);
  }

  resetSidebarWidth(side: ResizeSide): void {
    if (side === 'left') {
      this.leftSidebarWidthSg.set(PageComponent.LEFT_DEFAULT);
      this.persistWidth(PageComponent.LEFT_KEY, PageComponent.LEFT_DEFAULT);
      return;
    }

    this.rightSidebarWidthSg.set(PageComponent.RIGHT_DEFAULT);
    this.persistWidth(PageComponent.RIGHT_KEY, PageComponent.RIGHT_DEFAULT);
  }

  private readonly onMouseMove = (event: MouseEvent): void => {
    if (!this.activeResize) {
      return;
    }

    const delta = event.clientX - this.startX;
    if (this.activeResize === 'left') {
      this.leftSidebarWidthSg.set(this.clampWidth(this.startWidth + delta, 'left'));
      return;
    }

    this.rightSidebarWidthSg.set(this.clampWidth(this.startWidth - delta, 'right'));
  };

  private readonly onMouseUp = (): void => {
    if (!this.activeResize) {
      return;
    }

    this.persistWidth(PageComponent.LEFT_KEY, this.leftSidebarWidthSg());
    this.persistWidth(PageComponent.RIGHT_KEY, this.rightSidebarWidthSg());

    this.activeResize = null;
    this.isResizingSg.set(false);
    window.removeEventListener('mousemove', this.onMouseMove);
    window.removeEventListener('mouseup', this.onMouseUp);
  };

  ngOnDestroy(): void {
    window.removeEventListener('mousemove', this.onMouseMove);
    window.removeEventListener('mouseup', this.onMouseUp);
  }

  private clampWidth(value: number, side: ResizeSide): number {
    const min = side === 'left' ? PageComponent.LEFT_MIN : PageComponent.RIGHT_MIN;
    return Math.max(min, Math.min(PageComponent.MAX_WIDTH, Math.round(value)));
  }

  private readWidth(key: string, fallback: number): number {
    if (typeof window === 'undefined') {
      return fallback;
    }

    const parsed = Number(window.localStorage.getItem(key));
    if (!Number.isFinite(parsed)) {
      return fallback;
    }

    return this.clampWidth(parsed, key === PageComponent.LEFT_KEY ? 'left' : 'right');
  }

  private persistWidth(key: string, value: number): void {
    if (typeof window === 'undefined') {
      return;
    }
    window.localStorage.setItem(key, String(value));
  }
}
