import { test, expect } from '@playwright/test';
import {
  loginAsAdmin,
  navigateToGenresManagement,
  createGenre,
  setupAlertListener,
  findGenreRow,
  generateTestName,
  assertNoDangerousElements,
  shortWait,
  verifyGenreExistsViaApi,
} from './helpers';

/**
 * XSS Test Suite - Stored XSS Attack Prevention
 *
 * Tests verify that the application prevents stored XSS attacks through genres.
 * Two valid defenses: (1) server rejects the payload, or (2) payload is stored
 * but rendered safely (escaped/sanitized).
 */

test.describe('XSS Attack Prevention - Stored XSS in Genres', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
  });

  test('should prevent stored XSS with simple script tag in genre name', async ({ page }) => {
    const xssPayload = '<script>alert("XSS Attack!")</script>';
    const testGenreName = generateTestName('TestGenre');
    const maliciousGenreName = `${testGenreName}${xssPayload}`;

    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousGenreName);
      await shortWait(page, 1000);

      expect(alertListener.wasAlertTriggered()).toBe(false);

      if (result.success) {
        const exists = await verifyGenreExistsViaApi(testGenreName);
        expect(exists, 'Genre should exist in database').toBe(true);

        const row = await findGenreRow(page, testGenreName);
        if (row) {
          const innerHTML = await row.innerHTML();
          expect(innerHTML).not.toMatch(/<script\b[^>]*>[\s\S]*?<\/script>/i);
          const text = await row.textContent();
          console.log('Genre displayed as:', text);
        } else {
          console.log('Genre stored but not visible on current page');
        }
      } else {
        console.log('Script tag payload rejected by server:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should prevent stored XSS with event handler attributes', async ({ page }) => {
    const xssPayload = '<img src=x onerror="alert(\'XSS\')"/>';
    const testGenreName = generateTestName('TestGenre');
    const maliciousGenreName = `${testGenreName}${xssPayload}`;

    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousGenreName);
      await shortWait(page, 1000);

      expect(alertListener.wasAlertTriggered()).toBe(false);

      if (result.success) {
        const dangerousImages = await page.locator('img[onerror]').count();
        expect(dangerousImages).toBe(0);
        console.log('img onerror payload stored but rendered safely');
      } else {
        console.log('img onerror payload rejected by server:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should prevent stored XSS with javascript: protocol', async ({ page }) => {
    const xssPayload = '<a href="javascript:alert(\'XSS\')">Click me</a>';
    const testGenreName = generateTestName('TestGenre');
    const maliciousGenreName = `${testGenreName}${xssPayload}`;

    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousGenreName);
      await shortWait(page, 1000);

      expect(alertListener.wasAlertTriggered()).toBe(false);

      if (result.success) {
        const dangerousLinks = await page.locator('a[href^="javascript:"]').count();
        expect(dangerousLinks).toBe(0);
        console.log('javascript: protocol payload stored but rendered safely');
      } else {
        console.log('javascript: protocol payload rejected by server:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should prevent stored XSS with SVG-based attack', async ({ page }) => {
    const xssPayload = '<svg/onload=alert("XSS")>';
    const testGenreName = generateTestName('TestGenre');
    const maliciousGenreName = `${testGenreName}${xssPayload}`;

    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousGenreName);
      await shortWait(page, 1000);

      expect(alertListener.wasAlertTriggered()).toBe(false);
      await assertNoDangerousElements(page);

      if (result.success) {
        console.log('SVG onload payload stored but rendered safely');
      } else {
        console.log('SVG onload payload rejected by server:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should display XSS payload as plain text', async ({ page }) => {
    const xssPayload = '<script>alert("XSS")</script>';
    const testGenreName = generateTestName('SafeTest');
    const maliciousGenreName = `${testGenreName}${xssPayload}`;

    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousGenreName);
      await shortWait(page, 1000);

      expect(alertListener.wasAlertTriggered()).toBe(false);

      if (result.success) {
        const exists = await verifyGenreExistsViaApi(testGenreName);
        expect(exists, 'Genre should exist in database').toBe(true);

        const row = await findGenreRow(page, testGenreName);
        if (row) {
          const text = await row.textContent();
          console.log('Genre text content:', text);
          expect(text!.length).toBeGreaterThan(0);
        } else {
          console.log('Genre stored but not visible on current page');
        }
      } else {
        console.log('Payload rejected by server — XSS prevented at input:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should handle genre viewing after XSS attempt', async ({ page }) => {
    const xssPayload = '<script>alert("Stored XSS")</script>';
    const testGenreName = generateTestName('ViewTest');
    const maliciousGenreName = `${testGenreName}${xssPayload}`;

    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousGenreName);
      await shortWait(page, 1000);

      if (result.rejected) {
        console.log('Payload rejected by server — XSS prevented at input');
        expect(alertListener.wasAlertTriggered()).toBe(false);
        return;
      }

      const row = await findGenreRow(page, testGenreName);
      if (!row) {
        // Genre may be on a different page — verify via API and skip DOM edit test
        const exists = await verifyGenreExistsViaApi(testGenreName);
        expect(exists, 'Genre should exist in database').toBe(true);
        console.log('Genre exists in DB but not visible on current page — skipping edit test');
        expect(alertListener.wasAlertTriggered()).toBe(false);
        return;
      }

      if (row) {
        const editButton = row.locator('.item-table__btn-edit');
        await editButton.click();

        await page.waitForSelector('app-genre-editor-dialog app-dialog .dialog__overlay', {
          state: 'visible',
          timeout: 5000,
        });
        await shortWait(page, 1000);

        expect(alertListener.wasAlertTriggered()).toBe(false);

        const inputValue = await page.inputValue(
          'app-genre-editor-dialog app-text-input input.text-input__input'
        );
        console.log('Input value in edit dialog:', inputValue);

        await page.click('app-genre-editor-dialog button.btn-secondary');
      }
    } finally {
      alertListener.cleanup();
    }
  });
});
