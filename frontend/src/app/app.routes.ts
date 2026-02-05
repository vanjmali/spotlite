import { Routes } from '@angular/router';
import {
  VerificationSuccessPage,
  VerificationFailurePage,
  CheckEmailPage,
  LoginPage,
  RegistrationPage,
  ProfilePage,
} from './pages';

import { HomePage } from './pages/home-page';
import { ForgotPasswordPage } from './pages/forgot-password-page/forgot-password-page';
import { ResetPasswordPage } from './pages/reset-password-page/reset-password-page';
import { AdminPage } from './pages/admin-page';
import { CredentialsStep, OtpStep } from './pages/login-page/components';
import {
  PersonalInfoStep,
  RegistrationCredentialsStep,
} from './pages/registration-page/components';
import { InboxPage } from './pages/inbox-page';

export const routes: Routes = [
  {
    path: 'login',
    component: LoginPage,
    children: [
      {
        path: 'credentials',
        component: CredentialsStep,
      },
      {
        path: 'otp',
        component: OtpStep,
      },
      {
        path: '',
        redirectTo: 'credentials',
        pathMatch: 'full',
      },
    ],
  },
  {
    path: 'register',
    component: RegistrationPage,
    children: [
      {
        path: 'personal-info',
        component: PersonalInfoStep,
      },
      {
        path: 'credentials',
        component: RegistrationCredentialsStep,
      },
      {
        path: '',
        redirectTo: 'personal-info',
        pathMatch: 'full',
      },
    ],
  },
  {
    path: 'verification-success',
    component: VerificationSuccessPage,
  },
  {
    path: 'verification-failure',
    component: VerificationFailurePage,
  },
  {
    path: 'check-email',
    component: CheckEmailPage,
  },
  {
    path: 'forgot-password',
    component: ForgotPasswordPage,
  },
  {
    path: 'reset-password',
    component: ResetPasswordPage,
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
];
