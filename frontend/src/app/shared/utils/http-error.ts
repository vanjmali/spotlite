export function getHttpErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'string') return error;
  if (error && typeof error === 'object') {
    const err = error as { error?: { message?: string } };
    return err.error?.message ?? fallback;
  }
  return fallback;
}
