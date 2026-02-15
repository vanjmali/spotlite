import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { BrandLogo } from '../brand-logo';
import { HeaderComponent } from '../header/header';
import { SiteFooterComponent } from '../site-footer';

@Component({
  selector: 'app-auth-layout',
  standalone: true,
  imports: [CommonModule, BrandLogo, MatIconModule, HeaderComponent, SiteFooterComponent],
  templateUrl: './auth-layout.html',
  styleUrls: ['./auth-layout.scss'],
})
export class AuthLayout {
  @Input() showHeader = false;
  @Input() title?: string;
  @Input() showStepCount = false;
  @Input() stepIndex: number | null = null;
  @Input() totalSteps: number | null = null;
  @Input() back?: () => void;
}
