import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthLayout, LoaderComponent, MessageComponent, VALIDATION_MESSAGES } from '@app/shared';
import { AuthFormFooter } from '@app/pages/registration-page/components';
import { MatIconModule } from '@angular/material/icon';
import { AuthService } from '@app/services/auth.service';

@Component({
  selector: 'app-register-verify-page',
  standalone: true,
  imports: [
    CommonModule,
    AuthLayout,
    LoaderComponent,
    MessageComponent,
    MatIconModule,
    AuthFormFooter,
  ],
  templateUrl: './verify-page.html',
  styleUrls: ['./verify-page.scss'],
})
export class VerifyPage implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);

  public readonly title = this.route.snapshot.data?.['authTitle'] ?? 'Verify your email';
  public readonly statusSg = signal<'loading' | 'success' | 'error'>('loading');
  public readonly errorSg = signal<string>('');

  public ngOnInit(): void {
    this.route.queryParamMap.subscribe((params) => {
      const token = params.get('token');
      if (!token) {
        this.statusSg.set('error');
        this.errorSg.set('Missing or invalid verification token.');
        return;
      }
      void this.verify(token);
    });
  }

  public goToLogin(): void {
    this.router.navigate(['/login']);
  }

  public goToRegister(): void {
    this.router.navigate(['/register']);
  }

  private async verify(token: string): Promise<void> {
    this.statusSg.set('loading');
    this.errorSg.set('');

    const result = await this.auth.verifyAccount(token);
    if (result.success) {
      this.statusSg.set('success');
      return;
    }

    this.statusSg.set('error');
    this.errorSg.set(result.error || VALIDATION_MESSAGES.VERIFICATION_FAILED);
  }
}
