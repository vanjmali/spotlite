import { Routes } from '@angular/router';
import {
  VerificationSuccessPage,
  VerificationFailurePage,
  CheckEmailPage,
  LoginPage,
  RegistrationPage,
} from './pages';

import { HomePage } from './pages/home-page';
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
    path: 'home',
    component: HomePage,
  },
  {
    path: '',
    redirectTo: 'login',
    pathMatch: 'full',
  },
];
