export function focusFirstFocusable(root: HTMLElement | null): void {
  if (!root) return;
  const target = root.querySelector<HTMLElement>(
    'input, select, textarea, button, [tabindex]:not([tabindex="-1"])'
  );
  target?.focus();
}
