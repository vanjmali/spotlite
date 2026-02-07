import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-auth-check-email',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  templateUrl: './auth-check-email.html',
  styleUrls: ['./auth-check-email.scss'],
})
export class AuthCheckEmail {
  public readonly titleSg = input<string>('Check your email', { alias: 'title' });
  public readonly messageSg = input<string>(
    "We've sent you a confirmation email to verify your account.",
    { alias: 'message' }
  );
  public readonly hintSg = input<string>("Didn't receive it? Check your spam folder.", {
    alias: 'hint',
  });
}
