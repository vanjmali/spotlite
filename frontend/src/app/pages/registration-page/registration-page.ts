import { Component, DestroyRef, effect, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { RegistrationStore } from './store';
import {
  PersonalInfoStep,
  RegistrationCredentialsStep,
} from './components';

@Component({
  selector: 'app-registration-page',
  standalone: true,
  imports: [CommonModule, RouterLink, PersonalInfoStep, RegistrationCredentialsStep],
  providers: [RegistrationStore],
  templateUrl: './registration-page.html',
  styleUrls: ['./registration-page.scss'],
})
export class RegistrationPage {
  private readonly store = inject(RegistrationStore);
  private readonly destroyRef = inject(DestroyRef);
  private handlingPop = false;
  private initializing = true;

  constructor() {
    this.initHistorySync();
  }

  public step = this.store.stepSg;

  private initHistorySync(): void {
    const popHandler = (event: PopStateEvent) => {
      const step = (event.state?.registrationStep as 'personal' | 'credentials') || 'personal';
      this.handlingPop = true;
      this.store.setStep(step);
      queueMicrotask(() => {
        this.handlingPop = false;
      });
    };

    window.addEventListener('popstate', popHandler);
    this.destroyRef.onDestroy(() => window.removeEventListener('popstate', popHandler));

    const initialStep =
      (history.state?.registrationStep as 'personal' | 'credentials' | undefined) || 'personal';
    this.store.setStep(initialStep);
    if (!history.state?.registrationStep) {
      history.replaceState({ ...history.state, registrationStep: initialStep }, '');
    }

    effect(() => {
      const step = this.store.stepSg();
      if (this.handlingPop || this.initializing) return;

      const current = history.state?.registrationStep as 'personal' | 'credentials' | undefined;
      if (current === step) return;

      history.pushState({ ...history.state, registrationStep: step }, '');
    });

    queueMicrotask(() => {
      this.initializing = false;
    });
  }
}
