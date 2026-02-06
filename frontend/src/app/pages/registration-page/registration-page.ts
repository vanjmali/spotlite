import { CommonModule } from '@angular/common';
import { Component, DestroyRef, computed, effect, inject, signal } from '@angular/core';
import { animate, style, transition, trigger } from '@angular/animations';
import { BrandLogo } from '@app/shared';
import {
  PersonalInfoStep,
  RegistrationCredentialsStep,
  RegistrationCheckEmailStep,
  AuthStepHeader,
} from './components';
import { RegistrationStore } from './store';

export type RegistrationStep = 'personal' | 'credentials' | 'verify';

@Component({
  selector: 'app-registration-page',
  standalone: true,
  imports: [
    CommonModule,
    AuthStepHeader,
    PersonalInfoStep,
    RegistrationCredentialsStep,
    RegistrationCheckEmailStep,
    BrandLogo,
  ],
  providers: [RegistrationStore],
  templateUrl: './registration-page.html',
  styleUrls: ['./registration-page.scss'],
  animations: [
    trigger('stepSlide', [
      transition(':enter', [
        style({ opacity: 0, transform: 'translateX({{ enterFrom }})' }),
        animate('260ms ease-out', style({ opacity: 1, transform: 'translateX(0)' })),
      ]),
      transition(':leave', [
        animate('220ms ease-in', style({ opacity: 0, transform: 'translateX({{ leaveTo }})' })),
      ]),
    ]),
  ],
})
export class RegistrationPage {
  private readonly store = inject(RegistrationStore);
  private readonly destroyRef = inject(DestroyRef);
  private handlingPop = false;
  private initializing = true;
  private previousStep: RegistrationStep = 'personal';

  constructor() {
    this.initHistorySync();
  }

  public step = this.store.stepSg;
  public readonly directionSg = signal<'forward' | 'back'>('forward');

  private readonly stepOrder: RegistrationStep[] = ['personal', 'credentials', 'verify'];
  private readonly stepTitles: Record<RegistrationStep, string> = {
    personal: "Let's you signed up",
    credentials: 'Protect your account',
    verify: 'Almost there!',
  };

  public readonly stepIndexSg = computed(
    () => this.stepOrder.indexOf(this.step()) + 1
  );
  public readonly stepTitleSg = computed(() => this.stepTitles[this.step()]);
  public readonly canGoBackSg = computed(
    () => this.stepOrder.indexOf(this.step()) > 0 && this.step() !== 'verify'
  );
  public readonly totalSteps = this.stepOrder.length;

  public goBack(): void {
    const index = this.stepOrder.indexOf(this.step());
    const prevStep = this.stepOrder[Math.max(0, index - 1)];
    this.store.setStep(prevStep);
  }

  public animationParams() {
    const direction = this.directionSg();
    return {
      value: this.step(),
      params: {
        enterFrom: direction === 'forward' ? '24px' : '-24px',
        leaveTo: direction === 'forward' ? '-24px' : '24px',
      },
    };
  }

  private initHistorySync(): void {
    const popHandler = (event: PopStateEvent) => {
      const step = (event.state?.registrationStep as RegistrationStep) || 'personal';
      this.handlingPop = true;
      this.store.setStep(step);
      queueMicrotask(() => {
        this.handlingPop = false;
      });
    };

    window.addEventListener('popstate', popHandler);
    this.destroyRef.onDestroy(() => window.removeEventListener('popstate', popHandler));

    const initialStep: RegistrationStep = 'personal';
    this.previousStep = initialStep;
    this.store.setStep(initialStep);

    history.replaceState({ ...history.state, registrationStep: initialStep }, '');

    effect(() => {
      const step = this.store.stepSg();
      const prevIndex = this.stepOrder.indexOf(this.previousStep);
      const nextIndex = this.stepOrder.indexOf(step);

      if (step !== this.previousStep) {
        this.directionSg.set(nextIndex >= prevIndex ? 'forward' : 'back');
        this.previousStep = step;
      }

      if (this.handlingPop || this.initializing) return;

      const current = history.state?.registrationStep as RegistrationStep | undefined;
      if (current === step) return;

      history.pushState({ ...history.state, registrationStep: step }, '');
    });

    queueMicrotask(() => {
      this.initializing = false;
    });
  }
}
