import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';
import { AuthLayout, FormFooter, StatusCard } from '@app/shared';

@Component({
  selector: 'app-not-found-page',
  standalone: true,
  imports: [CommonModule, AuthLayout, StatusCard, FormFooter],
  templateUrl: './not-found-page.html',
  styleUrls: ['./not-found-page.scss'],
})
export class NotFoundPage {
  private readonly router = inject(Router);

  public goHome(): void {
    this.router.navigate(['/']);
  }
}
