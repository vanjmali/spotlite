import { Component, computed, effect, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterOutlet } from '@angular/router';
import { LoginStore } from './store';
import { AuthStepHeader } from '../registration-page/components';
import { BrandLogo } from '@app/shared';

@Component({
  selector: 'app-login-page',
  standalone: true,
  imports: [CommonModule, AuthStepHeader, RouterOutlet, BrandLogo],
  providers: [LoginStore],
  templateUrl: './login-page.html',
  styleUrls: ['./login-page.scss'],
})
export class LoginPage {
  private readonly store = inject(LoginStore);
  private readonly router = inject(Router);

  public readonly stepSg = signal<'credentials' | 'otp'>('credentials');
  private readonly stepTitles: Record<'credentials' | 'otp', string> = {
    credentials: 'Welcome back',
    otp: 'Verify your code',
  };

  public readonly stepTitleSg = computed(() => this.stepTitles[this.stepSg()]);
  public readonly canGoBackSg = computed(() => this.stepSg() === 'otp');

  public goBack(): void {
    this.router.navigate(['/login']);
  }

  constructor() {
    effect(() => {
      const url = this.router.url;
      const step = url.includes('/login/otp') ? 'otp' : 'credentials';
      this.stepSg.set(step);
      this.store.stepSg.set(step);

      if (step === 'credentials') {
        this.store.clearOtpSession();
      }
    });

  }
}
