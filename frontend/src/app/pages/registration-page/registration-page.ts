import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, RouterOutlet } from '@angular/router';
import { RegistrationStore } from './store';

@Component({
  selector: 'app-registration-page',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterOutlet],
  providers: [RegistrationStore],
  templateUrl: './registration-page.html',
  styleUrls: ['./registration-page.scss'],
})
export class RegistrationPage {}
