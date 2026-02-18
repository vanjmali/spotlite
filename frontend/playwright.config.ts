import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E test configuration.
 *
 * The app runs as a Docker Compose stack (Go backends, MongoDB, MailHog, Traefik).
 * Start the stack with `docker compose up` before running tests.
 */
export default defineConfig({
  testDir: './e2e',
  /* Run tests sequentially — all XSS tests share one login email/MailHog */
  fullyParallel: false,
  /* Single worker to avoid OTP race conditions (shared MailHog + email) */
  workers: 1,
  reporter: 'html',
  use: {
    baseURL: 'https://localhost:4443',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    ignoreHTTPSErrors: true,
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
