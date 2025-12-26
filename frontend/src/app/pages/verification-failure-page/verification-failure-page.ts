import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-verification-failure-page',
  standalone: true,
  imports: [CommonModule, MatIconModule, RouterLink],
  templateUrl: './verification-failure-page.html',
  styleUrls: ['./verification-failure-page.scss'],
})
export class VerificationFailurePage {}
