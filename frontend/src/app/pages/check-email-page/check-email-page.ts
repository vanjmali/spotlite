import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink } from '@angular/router';
import { AuthLayout } from '@app/shared/components/auth-layout';

@Component({
  selector: 'app-check-email-page',
  standalone: true,
  imports: [CommonModule, MatIconModule, RouterLink, AuthLayout],
  templateUrl: './check-email-page.html',
  styleUrls: ['./check-email-page.scss'],
})
export class CheckEmailPage {}
