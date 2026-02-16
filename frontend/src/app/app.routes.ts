import { inject } from '@angular/core';
import { CanMatchFn, Routes } from '@angular/router';
import { CheckEmailPage, RegistrationPage, ProfilePage } from './pages';

import { HomePage } from './pages/home-page';
import { ResetPasswordPage } from './pages/reset-password-page/reset-password-page';
import { AdminPage } from './pages/admin-page';
import { CredentialsPage } from './pages/login/credentials-page';
import { OtpPage } from './pages/login/otp-page';
import { VerifyPage } from './pages/register/verify-page';
import { InboxPage } from './pages/inbox-page';
import { LoginStore } from './pages/login/store';
import { NotFoundPage } from './pages/not-found-page';
import { AuthService } from './services/auth.service';

const adminOnlyMatch: CanMatchFn = async () => {
  const authService = inject(AuthService);

  if (authService.isAdminSg()) {
    return true;
  }

  const refreshed = await authService.refreshAccessToken();
  if (!refreshed) {
    return false;
  }

  return authService.isAdminSg();
};

export const routes: Routes = [
  {
    path: 'login',
    providers: [LoginStore],
    children: [
      {
        path: '',
        component: CredentialsPage,
        data: { authTitle: 'Welcome back' },
      },
      {
        path: 'otp',
        component: OtpPage,
        data: { authTitle: 'Verify your code' },
      },
      {
        path: 'reset',
        component: ResetPasswordPage,
        data: { authTitle: 'Reset your password' },
      },
    ],
  },
  {
    path: 'register/verify',
    component: VerifyPage,
    data: { authTitle: 'Verify your email' },
  },
  {
    path: 'register',
    component: RegistrationPage,
  },
  {
    path: 'check-email',
    component: CheckEmailPage,
  },
  {
    path: 'inbox',
    component: InboxPage,
  },
  {
    path: 'profile',
    component: ProfilePage,
  },
  {
    path: 'admin',
    component: AdminPage,
    canMatch: [adminOnlyMatch],
    children: [
      {
        path: 'genres',
        loadComponent: () =>
          import('./pages/admin-page/components/genres-management/genres-management.component').then(
            (m) => m.GenresManagementComponent
          ),
      },
      {
        path: 'artists',
        loadComponent: () =>
          import('./pages/admin-page/components/artists-management/artists-management.component').then(
            (m) => m.ArtistsManagementComponent
          ),
      },
      {
        path: 'albums',
        loadComponent: () =>
          import('./pages/admin-page/components/albums-management/albums-management.component').then(
            (m) => m.AlbumsManagementComponent
          ),
      },
      {
        path: 'songs',
        loadComponent: () =>
          import('./pages/admin-page/components/songs-management/songs-management.component').then(
            (m) => m.SongsManagementComponent
          ),
      },
      {
        path: '',
        redirectTo: 'artists',
        pathMatch: 'full',
      },
    ],
  },
  {
    path: '',
    component: HomePage,
    children: [
      {
        path: '',
        loadComponent: () =>
          import('./pages/home-page/components/artists-list/artists-list.component').then(
            (m) => m.ArtistsListComponent
          ),
      },
      {
        path: 'artist/:id',
        loadComponent: () =>
          import('./pages/home-page/components/artist-details/artist-details.component').then(
            (m) => m.ArtistDetailsComponent
          ),
      },
      {
        path: 'album/:id',
        loadComponent: () =>
          import('./pages/home-page/components/album-details/album-details.component').then(
            (m) => m.AlbumDetailsComponent
          ),
      },
    ],
  },
  {
    path: '**',
    component: NotFoundPage,
  },
];
