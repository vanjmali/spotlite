import { Routes } from '@angular/router';
import { LoginPage } from './pages/login-page';
import { RegistrationPage } from './pages/registration-page';
import { CredentialsStep, OtpStep } from './pages/login-page/components';
import {
  PersonalInfoStep,
  RegistrationCredentialsStep,
} from './pages/registration-page/components';

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
    path: '',
    redirectTo: 'login',
    pathMatch: 'full',
  },
];
