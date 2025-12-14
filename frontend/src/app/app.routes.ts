import { Routes } from '@angular/router';
import { LoginPage } from './pages/login-page';
import { EmailStep, OtpStep, PasswordStep } from './pages/login-page/components';

export const routes: Routes = [
  {
    path: 'login',
    component: LoginPage,
    children: [
      {
        path: 'email',
        component: EmailStep,
      },
      {
        path: 'otp',
        component: OtpStep,
      },
      {
        path: 'password',
        component: PasswordStep,
      },
      {
        path: '',
        redirectTo: 'email',
        pathMatch: 'full',
      },
    ],
  },
];
