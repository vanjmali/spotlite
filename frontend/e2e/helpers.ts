import { Page, expect, request } from '@playwright/test';

const env = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process
  ?.env;
const E2E_BASE_URL = env?.['PLAYWRIGHT_BASE_URL'] || env?.['BASE_URL'] || 'https://localhost:4443';

async function createApiContext() {
  return request.newContext({
    baseURL: E2E_BASE_URL,
    ignoreHTTPSErrors: true,
  });
}

/**
 * Test credentials for different user roles
 */
export const TEST_CREDENTIALS = {
  admin: {
    email: 'admin@example.com',
    password: 'SuperAdmin123!',
  },
};

const MAILHOG_API = 'http://localhost:8025';
let lastLoginSubmitAt = 0;

/**
 * Common XSS payloads for testing
 */
export const XSS_PAYLOADS = {
  scriptTag: '<script>alert("XSS")</script>',
  scriptTagEncoded: '&lt;script&gt;alert("XSS")&lt;/script&gt;',
  imgOnerror: '<img src=x onerror="alert(\'XSS\')"/>',
  svgOnload: '<svg/onload=alert("XSS")>',
  javascriptProtocol: '<a href="javascript:alert(\'XSS\')">Click</a>',
  iframeWithJs: '<iframe src="javascript:alert(\'XSS\')"></iframe>',
  bodyOnload: '<body onload=alert("XSS")>',
  divOnmouseover: '<div onmouseover="alert(\'XSS\')">Hover me</div>',
  inputAutofocus: '<input autofocus onfocus=alert("XSS")>',
  objectData: '<object data="javascript:alert(\'XSS\')">',
};

/**
 * Common NoSQL/SQL Injection payloads for testing
 * These target MongoDB's query operators and regex patterns
 */
export const SQL_INJECTION_PAYLOADS = {
  // NoSQL operator injection - attempts to bypass authentication/filters
  operatorInjection: '{"$ne": null}',
  operatorInjectionAlt: '{"$gt": ""}',

  // MongoDB $where injection - attempts code execution
  whereInjection: "'; return true; var foo = '",
  whereInjectionAlt: '"; return true; var foo = "',

  // Regex injection - attempts to cause ReDoS or bypass filters
  regexInjection: '.*',
  regexInjectionDos: '(a+)+$',
  regexSpecialChars: '$^.*+?()[]{}|\\',

  // JSON injection in query strings
  jsonPayload: '{"$regex": ".*"}',
  jsonArrayPayload: '["admin", {"$gt": ""}]',

  // Traditional SQL injection patterns (for testing sanitization)
  sqlOr: "' OR '1'='1",
  sqlUnion: "' UNION SELECT * FROM users--",
  sqlComment: "admin'--",
  sqlStacked: "'; DROP TABLE users;--",

  // Special MongoDB operators
  existsOperator: '{"$exists": true}',
  typeOperator: '{"$type": "string"}',
  sizeOperator: '{"$size": 0}',
};

/**
 * Fetch the latest OTP code from MailHog for a given email address
 */
async function fetchOtpFromMailHog(email: string, excludedCodes: Set<string> = new Set()): Promise<string> {
  const api = await request.newContext();

  // Poll MailHog for the latest message to the given email
  for (let attempt = 0; attempt < 45; attempt++) {
    const res = await api.get(
      `${MAILHOG_API}/api/v2/search?kind=to&query=${encodeURIComponent(email)}`
    );
    const data = await res.json();

    if (data.items && data.items.length > 0) {
      // Get the most recent message
      const latest = data.items[0];
      const body = latest.Content?.Body || latest.Raw?.Data || '';

      // Extract the OTP from the .otp-code element in the HTML email
      // The email body uses quoted-printable encoding, so look near the "otp-code" class
      const otpCodeMatch = body.match(/otp-code[^>]*>(\d{6})</);
      if (otpCodeMatch && !excludedCodes.has(otpCodeMatch[1])) {
        await api.dispose();
        return otpCodeMatch[1];
      }

      // Fallback: look for a standalone 6-digit number that's NOT a CSS hex color
      // (CSS colors are preceded by # and may repeat patterns like 121212)
      const allDigits = [...body.matchAll(/(?<!#|\w)(\d{6})(?!\d)/g)];
      for (const m of allDigits) {
        const code = m[1];
        // Skip common CSS color-like patterns (e.g., 121212, 000000, 333333, etc.)
        if (
          !excludedCodes.has(code) &&
          !/^(.)\1{5}$/.test(code) &&
          !/^(.{2})\1{2}$/.test(code)
        ) {
          await api.dispose();
          return code;
        }
      }
    }

    // Wait before retrying
    await new Promise((r) => setTimeout(r, 1000));
  }

  await api.dispose();
  throw new Error(`Could not fetch OTP from MailHog for ${email}`);
}

/**
 * Delete all MailHog messages for a given email address
 */
async function clearMailHogMessages(): Promise<void> {
  const api = await request.newContext();
  await api.delete(`${MAILHOG_API}/api/v1/messages`);
  await api.dispose();
}

/**
 * Login helper function — handles email/password + OTP verification
 */
export async function login(page: Page, email: string, password: string) {
  const otpBoxes = page.locator('.otp-input__box');
  const loginError = page.locator('text=/Invalid credentials|Too many requests|Login failed/i');
  const otpError = page.locator(
    'text=/Verification code has expired|Invalid code|OTP verification failed/i'
  );

  for (let authAttempt = 0; authAttempt < 4; authAttempt++) {
    // Keep mailbox clean per attempt so OTP retrieval is deterministic.
    await clearMailHogMessages();

    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[type="email"]', email);
    await page.fill('input[type="password"]', password);

    // Throttle login submissions to reduce backend rate limiting.
    const elapsedSinceLastSubmit = Date.now() - lastLoginSubmitAt;
    if (elapsedSinceLastSubmit < 2000) {
      await page.waitForTimeout(2000 - elapsedSinceLastSubmit);
    }
    await page.click('button[type="submit"]');
    lastLoginSubmitAt = Date.now();

    const stepResult = await Promise.race([
      page.waitForURL('**/login/otp', { timeout: 15000 }).then(() => 'otp' as const),
      otpBoxes.first().waitFor({ state: 'visible', timeout: 15000 }).then(() => 'otp' as const),
      loginError.first().waitFor({ state: 'visible', timeout: 15000 }).then(() => 'login-error' as const),
    ]);

    if (stepResult === 'login-error') {
      const text = ((await loginError.first().textContent()) || '').toLowerCase();
      if (text.includes('too many requests') && authAttempt < 3) {
        await page.waitForTimeout(3000 * (authAttempt + 1));
        continue;
      }
      throw new Error(`Login blocked for ${email}. URL=${page.url()} Error=${text}`);
    }

    await page.waitForLoadState('networkidle');

    // OTP flow with multiple resend/fetch retries.
    const usedOtps = new Set<string>();
    let otpVerified = false;
    for (let otpAttempt = 0; otpAttempt < 3; otpAttempt++) {
      let otp: string;
      try {
        otp = await fetchOtpFromMailHog(email, usedOtps);
      } catch {
        if (otpAttempt < 2) {
          await page.getByRole('button', { name: /Resend Code/i }).click();
          continue;
        }
        break;
      }

      usedOtps.add(otp);
      for (let i = 0; i < 6; i++) {
        await otpBoxes.nth(i).fill(otp[i]);
      }

      const otpResult = await Promise.race([
        page
          .waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 15000 })
          .then(() => 'navigated' as const),
        otpError
          .first()
          .waitFor({ state: 'visible', timeout: 15000 })
          .then(() => 'otp-error' as const),
      ]);

      if (otpResult === 'navigated') {
        otpVerified = true;
        break;
      }

      if (otpAttempt < 2) {
        await page.getByRole('button', { name: /Resend Code/i }).click();
      }
    }

    if (otpVerified) {
      await page.waitForLoadState('networkidle');
      return;
    }
  }

  const bodyText = (await page.locator('body').textContent()) || '';
  throw new Error(`Login failed after retries for ${email}. URL=${page.url()} Body=${bodyText.slice(0, 300)}`);
}

/**
 * Login as admin helper
 */
export async function loginAsAdmin(page: Page) {
  await login(page, TEST_CREDENTIALS.admin.email, TEST_CREDENTIALS.admin.password);
}

/**
 * Setup alert dialog listener
 * Returns an object with methods to check if alert was triggered and cleanup
 */
export function setupAlertListener(page: Page): {
  wasAlertTriggered: () => boolean;
  getAlertMessages: () => string[];
  cleanup: () => void;
} {
  let alertTriggered = false;
  const alertMessages: string[] = [];

  const handler = async (dialog: any) => {
    alertTriggered = true;
    alertMessages.push(dialog.message());
    await dialog.dismiss();
  };

  page.on('dialog', handler);

  return {
    wasAlertTriggered: () => alertTriggered,
    getAlertMessages: () => [...alertMessages],
    cleanup: () => page.off('dialog', handler),
  };
}

/**
 * Check if any dangerous HTML elements exist in the page
 */
export async function checkForDangerousElements(page: Page): Promise<{
  scriptsFound: number;
  imgsWithOnerror: number;
  svgsWithOnload: number;
  linksWithJavascript: number;
  iframesWithJavascript: number;
}> {
  return {
    scriptsFound: await page.locator('script:not([src])').count(),
    imgsWithOnerror: await page.locator('img[onerror]').count(),
    svgsWithOnload: await page.locator('svg[onload]').count(),
    linksWithJavascript: await page.locator('a[href^="javascript:"]').count(),
    iframesWithJavascript: await page.locator('iframe[src^="javascript:"]').count(),
  };
}

/**
 * Verify that no dangerous elements were injected
 */
export async function assertNoDangerousElements(page: Page) {
  const dangerous = await checkForDangerousElements(page);

  expect(dangerous.scriptsFound, 'Unexpected inline scripts found').toBe(0);
  expect(dangerous.imgsWithOnerror, 'Images with onerror handlers found').toBe(0);
  expect(dangerous.svgsWithOnload, 'SVG elements with onload handlers found').toBe(0);
  expect(dangerous.linksWithJavascript, 'Links with javascript: protocol found').toBe(0);
  expect(dangerous.iframesWithJavascript, 'Iframes with javascript: protocol found').toBe(0);
}

/**
 * Navigate to admin genres page using SPA click navigation (preserves in-memory auth token)
 */
export async function navigateToGenresManagement(page: Page) {
  // Step 1: Click the Admin button in the header (navigates to /admin via Angular router)
  await page.click('button[aria-label="Admin"]', { timeout: 10000 });
  await page.waitForURL('**/admin/**', { timeout: 10000 });
  await page.waitForLoadState('networkidle');

  // Step 2: Click the "Manage Genres" link in the sidebar
  await page.click('a[href="/admin/genres"]', { timeout: 10000 });
  await page.waitForURL('**/admin/genres', { timeout: 10000 });
  await page.waitForLoadState('networkidle');

  // Step 3: Wait for genres page content to render
  await page.waitForSelector(
    'app-genres-management, .genres-management__btn-add-circle, .item-table__table, .item-table__empty',
    { timeout: 15000 }
  );

  // Step 4: If there's a rows-per-page dropdown, set it to 50 so all genres are visible
  const rowsDropdown = page.locator('select, [role="combobox"]').last();
  if ((await rowsDropdown.count()) > 0) {
    try {
      await rowsDropdown.selectOption('50');
      await page.waitForLoadState('networkidle');
    } catch {
      // Dropdown might not be present or have different options — continue
    }
  }
}

/**
 * Open genre creation dialog
 */
export async function openCreateGenreDialog(page: Page) {
  await page.click('button.genres-management__btn-add-circle', { timeout: 10000 });
  // Wait for the dialog overlay to appear
  await page.waitForSelector('app-genre-editor-dialog app-dialog .dialog__overlay', {
    state: 'visible',
    timeout: 5000,
  });
}

/**
 * Result of a genre form submission
 */
export type GenreSubmitResult = {
  /** Whether the genre was successfully created (dialog closed) */
  success: boolean;
  /** Whether the server rejected the payload (dialog stayed open with error) */
  rejected: boolean;
  /** Error message if rejected */
  errorMessage?: string;
};

/**
 * Fill and submit genre form.
 * Returns whether submission succeeded or was rejected.
 */
export async function submitGenreForm(page: Page, genreName: string): Promise<GenreSubmitResult> {
  // Fill the name input inside the custom TextInputComponent
  const nameInput = page.locator('app-genre-editor-dialog app-text-input input.text-input__input');
  await nameInput.fill(genreName);

  // Click the "Create" (or "Update") button
  await page.click('app-genre-editor-dialog button.btn-primary');

  // Wait for either: dialog closes (success) or error appears (rejection)
  const dialogOverlay = page.locator('app-genre-editor-dialog app-dialog .dialog__overlay');

  try {
    // First check if dialog closes quickly (successful creation)
    await dialogOverlay.waitFor({ state: 'hidden', timeout: 5000 });
    await page.waitForLoadState('networkidle');
    // Brief wait for backend to commit and table to refresh
    await page.waitForTimeout(1000);
    return { success: true, rejected: false };
  } catch {
    // Dialog is still open — check for error message
    const errorMessage = await page.locator('app-genre-editor-dialog').textContent();
    const hasError = errorMessage?.includes('Failed') || errorMessage?.includes('error');

    if (hasError) {
      // Close the dialog manually
      await page.click('app-genre-editor-dialog button.btn-secondary');
      await dialogOverlay.waitFor({ state: 'hidden', timeout: 5000 });
      return { success: false, rejected: true, errorMessage: errorMessage?.trim() };
    }

    // Unknown state — still a failure
    throw new Error('Genre form submission timed out without success or clear error');
  }
}

/**
 * Create a genre with given name (opens dialog, fills, submits).
 * Returns the submission result.
 */
export async function createGenre(page: Page, genreName: string): Promise<GenreSubmitResult> {
  await openCreateGenreDialog(page);
  return await submitGenreForm(page, genreName);
}

/**
 * Verify a genre exists via the API (not limited by UI pagination).
 * Uses the GET /api/content/genres?name= filter.
 */
export async function verifyGenreExistsViaApi(partialName: string): Promise<boolean> {
  const api = await createApiContext();

  for (let attempt = 0; attempt < 5; attempt++) {
    try {
      const res = await api.get(
        `/api/content/genres?name=${encodeURIComponent(partialName)}&page_size=50`
      );
      if (!res.ok()) {
        throw new Error(`Genre API returned ${res.status()}`);
      }
      const data = await res.json();

      if (data.items && data.items.length > 0) {
        await api.dispose();
        return true;
      }
    } catch {
      // Retry on failure
    }

    // Wait before retry to handle backend write latency
    await new Promise((r) => setTimeout(r, 2000));
  }

  await api.dispose();
  return false;
}

/**
 * Find a genre row in the UI table by partial name match.
 * If not found initially, reloads the genre list by clicking the sidebar link.
 */
export async function findGenreRow(page: Page, partialName: string, retries = 3) {
  for (let attempt = 0; attempt < retries; attempt++) {
    const rows = await page.locator('.item-table__row').all();

    for (const row of rows) {
      const text = await row.textContent();
      if (text && text.includes(partialName)) {
        return row;
      }
    }

    // Not found yet — force a fresh data load by re-clicking the genres sidebar link
    if (attempt < retries - 1) {
      await page.waitForTimeout(2000);
      // Click the "Manage Genres" link to trigger a fresh loadGenres()
      await page.click('a[href="/admin/genres"]');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(500);

      // Re-select 50 rows per page if pagination dropdown exists
      const rowsDropdown = page.locator('select, [role="combobox"]').last();
      if ((await rowsDropdown.count()) > 0) {
        try {
          await rowsDropdown.selectOption('50');
          await page.waitForLoadState('networkidle');
        } catch {
          // Ignore
        }
      }
    }
  }

  return null;
}

/**
 * Generate a unique test name with timestamp
 */
export function generateTestName(prefix: string = 'Test'): string {
  return `${prefix}_${Date.now()}`;
}

/**
 * Wait for a short period (use sparingly)
 */
export async function shortWait(page: Page, ms: number = 500) {
  await page.waitForTimeout(ms);
}

/**
 * Navigate to admin artists page using SPA click navigation (preserves in-memory auth token)
 */
export async function navigateToArtistsManagement(page: Page) {
  // Step 1: Click the Admin button in the header (navigates to /admin via Angular router)
  await page.click('button[aria-label="Admin"]', { timeout: 10000 });
  await page.waitForURL('**/admin/**', { timeout: 10000 });
  await page.waitForLoadState('networkidle');

  // Step 2: Click the "Manage Artists" link in the sidebar
  await page.click('a[href="/admin/artists"]', { timeout: 10000 });
  await page.waitForURL('**/admin/artists', { timeout: 10000 });
  await page.waitForLoadState('networkidle');

  // Step 3: Wait for artists page content to render
  await page.waitForSelector(
    'app-artists-management, .artists-management__btn-add-circle, .item-table__table, .item-table__empty',
    { timeout: 15000 }
  );

  // Step 4: If there's a rows-per-page dropdown, set it to 50 so all artists are visible
  const rowsDropdown = page.locator('select.pagination__size-select');
  if ((await rowsDropdown.count()) > 0) {
    try {
      await rowsDropdown.selectOption('50');
      await page.waitForLoadState('networkidle');
    } catch {
      // Dropdown might not be present or have different options — continue
    }
  }
}

/**
 * Open artist creation dialog
 */
export async function openCreateArtistDialog(page: Page) {
  await page.click('button.artists-management__btn-add-circle', { timeout: 10000 });
  // Wait for the dialog overlay to appear
  await page.waitForSelector('app-artist-editor-dialog app-dialog .dialog__overlay', {
    state: 'visible',
    timeout: 5000,
  });
}

/**
 * Result of an artist form submission
 */
export type ArtistSubmitResult = {
  /** Whether the artist was successfully created (dialog closed) */
  success: boolean;
  /** Whether the server rejected the payload (dialog stayed open with error) */
  rejected: boolean;
  /** Error message if rejected */
  errorMessage?: string;
};

/**
 * Fill and submit artist form.
 * The artist form requires: name (≥2 chars), description (≥2 chars), and at least one genre.
 * Returns whether submission succeeded or was rejected.
 */
export async function submitArtistForm(
  page: Page,
  artistName: string,
  description: string = 'Test artist description'
): Promise<ArtistSubmitResult> {
  // Wait for genre options to load (they load async when dialog opens)
  await page.waitForTimeout(1500);

  // Fill the name input inside the custom TextInputComponent
  const nameInput = page.locator('app-artist-editor-dialog app-text-input input.text-input__input');
  await nameInput.fill(artistName);

  // Fill description (required, min 2 chars) — class is textarea-input__input
  const descInput = page.locator(
    'app-artist-editor-dialog app-textarea-input textarea.textarea-input__input'
  );
  await descInput.fill(description);

  // Select at least one genre (required):
  // 1. Open and wait for options with retries.
  const genreCombobox = page.locator(
    'app-artist-editor-dialog app-select-input [role="combobox"]'
  );
  const firstOption = page.locator('app-artist-editor-dialog app-select-input .select-input__option').first();
  const noOptionsLabel = page.locator('app-artist-editor-dialog app-select-input .select-input__no-options');
  const searchInput = page.locator('app-artist-editor-dialog app-select-input .select-input__search');

  let hasVisibleOption = false;
  for (let attempt = 0; attempt < 4; attempt++) {
    await genreCombobox.click();
    await page.waitForTimeout(400);

    const optionCount = await page
      .locator('app-artist-editor-dialog app-select-input .select-input__option')
      .count();
    if (optionCount > 0 && (await firstOption.isVisible())) {
      hasVisibleOption = true;
      break;
    }

    if ((await noOptionsLabel.count()) > 0 && (await noOptionsLabel.first().isVisible())) {
      if ((await searchInput.count()) > 0) {
        await searchInput.fill('');
      }
      await page.waitForTimeout(1500);
    }

    await page.keyboard.press('Escape');
    await page.waitForTimeout(600);
  }

  if (!hasVisibleOption) {
    const dialogOverlay = page.locator('app-artist-editor-dialog app-dialog .dialog__overlay');
    await page.click('app-artist-editor-dialog button.btn-secondary');
    await dialogOverlay.waitFor({ state: 'hidden', timeout: 5000 });
    return {
      success: false,
      rejected: true,
      errorMessage: 'No selectable genres available in artist dialog',
    };
  }

  // 3. Click the first genre checkbox/option
  await firstOption.click();

  // 4. Close the dropdown by pressing Escape (the component handles keydown.escape)
  await page.keyboard.press('Escape');
  await page.waitForTimeout(300);

  // Click the "Create" (or "Update") button
  await page.click('app-artist-editor-dialog button.btn-primary');

  // Wait for either: dialog closes (success) or error appears (rejection)
  const dialogOverlay = page.locator('app-artist-editor-dialog app-dialog .dialog__overlay');

  try {
    // First check if dialog closes quickly (successful creation)
    await dialogOverlay.waitFor({ state: 'hidden', timeout: 5000 });
    await page.waitForLoadState('networkidle');
    // Brief wait for backend to commit and table to refresh
    await page.waitForTimeout(1000);
    return { success: true, rejected: false };
  } catch {
    // Dialog is still open — check for error message
    const errorMessage = await page.locator('app-artist-editor-dialog').textContent();
    const hasError = errorMessage?.includes('Failed') || errorMessage?.includes('error');

    if (hasError) {
      // Close the dialog manually
      await page.click('app-artist-editor-dialog button.btn-secondary');
      await dialogOverlay.waitFor({ state: 'hidden', timeout: 5000 });
      return { success: false, rejected: true, errorMessage: errorMessage?.trim() };
    }

    // Unknown state — still a failure
    throw new Error('Artist form submission timed out without success or clear error');
  }
}

/**
 * Create an artist with given name (opens dialog, fills, submits).
 * Returns the submission result.
 */
export async function createArtist(
  page: Page,
  artistName: string,
  description?: string
): Promise<ArtistSubmitResult> {
  await openCreateArtistDialog(page);
  return await submitArtistForm(page, artistName, description);
}

/**
 * Verify an artist exists via the API (not limited by UI pagination).
 * Uses the GET /api/content/artists?name= filter.
 */
export async function verifyArtistExistsViaApi(partialName: string): Promise<boolean> {
  const api = await createApiContext();

  for (let attempt = 0; attempt < 5; attempt++) {
    try {
      const res = await api.get(
        `/api/content/artists?name=${encodeURIComponent(partialName)}&page_size=50`
      );
      if (!res.ok()) {
        throw new Error(`Artist API returned ${res.status()}`);
      }
      const data = await res.json();

      if (data.items && data.items.length > 0) {
        await api.dispose();
        return true;
      }
    } catch {
      // Retry on failure
    }

    // Wait before retry to handle backend write latency
    await new Promise((r) => setTimeout(r, 2000));
  }

  await api.dispose();
  return false;
}

/**
 * Find an artist row in the UI table by partial name match.
 * If not found initially, reloads the artist list by clicking the sidebar link.
 */
export async function findArtistRow(page: Page, partialName: string, retries = 3) {
  for (let attempt = 0; attempt < retries; attempt++) {
    const rows = await page.locator('.item-table__row').all();

    for (const row of rows) {
      const text = await row.textContent();
      if (text && text.includes(partialName)) {
        return row;
      }
    }

    // Not found yet — force a fresh data load by re-clicking the artists sidebar link
    if (attempt < retries - 1) {
      await page.waitForTimeout(2000);
      // Click the "Manage Artists" link to trigger a fresh loadArtists()
      await page.click('a[href="/admin/artists"]');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(500);

      // Re-select 50 rows per page if pagination dropdown exists
      const rowsDropdown = page.locator('select, [role="combobox"]').last();
      if ((await rowsDropdown.count()) > 0) {
        try {
          await rowsDropdown.selectOption('50');
          await page.waitForLoadState('networkidle');
        } catch {
          // Ignore
        }
      }
    }
  }

  return null;
}

/**
 * Test search functionality with a given query parameter
 * Returns the API response for validation
 */
export async function testSearchQuery(
  endpoint: string,
  queryParam: string,
  queryValue: string
): Promise<any> {
  const api = await createApiContext();

  try {
    const res = await api.get(`${endpoint}?${queryParam}=${encodeURIComponent(queryValue)}`);
    const data = await res.json();
    await api.dispose();
    return { status: res.status(), data };
  } catch (error) {
    await api.dispose();
    return { status: 500, error };
  }
}
