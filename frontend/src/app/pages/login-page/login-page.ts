import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';
import { LoginStore } from './store';

@Component({
  selector: 'app-login-page',
  standalone: true,
  imports: [CommonModule, RouterOutlet],
  providers: [LoginStore],
  templateUrl: './login-page.html',
  styleUrls: ['./login-page.scss'],
})
export class LoginPage {}
