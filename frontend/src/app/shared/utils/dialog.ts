import { effect, type Signal } from '@angular/core';

export function runOnOpen(isOpen: Signal<boolean>, onOpen: () => void): void {
  effect(() => {
    if (isOpen()) {
      onOpen();
    }
  });
}
