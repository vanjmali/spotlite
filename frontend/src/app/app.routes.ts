import { Routes } from '@angular/router';
import { LoginPage } from './pages/login-page';
import { CredentialsStep, OtpStep } from './pages/login-page/components';

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
];
