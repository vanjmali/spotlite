import { test, expect } from '@playwright/test';
import {
  loginAsAdmin,
  navigateToGenresManagement,
  createGenre,
  setupAlertListener,
  findGenreRow,
  generateTestName,
  XSS_PAYLOADS,
  assertNoDangerousElements,
  shortWait,
  verifyGenreExistsViaApi,
} from './helpers';

/**
 * Simplified XSS Test Suite for Stored XSS Prevention
 *
 * This test suite uses helper functions for cleaner and more maintainable tests.
 */
test.describe('XSS Prevention - Stored XSS (Simplified)', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin before each test
    await loginAsAdmin(page);
  });

  test('should block <script> tag injection', async ({ page }) => {
    const testName = generateTestName('ScriptTest');
    const maliciousName = `${testName}${XSS_PAYLOADS.scriptTag}`;
    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousName);
      await shortWait(page, 1000);

      // PRIMARY SECURITY ASSERTION: no JavaScript executed
      expect(alertListener.wasAlertTriggered()).toBe(false);
      await assertNoDangerousElements(page);

      if (result.success) {
        // Verify genre was persisted via API (not limited by UI pagination)
        const exists = await verifyGenreExistsViaApi(testName);
        expect(exists, 'Genre should exist in database after creation').toBe(true);
        console.log('Script payload stored but rendered safely');
      } else {
        console.log('Script payload rejected by server:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should block img onerror handler', async ({ page }) => {
    const testName = generateTestName('ImgTest');
    const maliciousName = `${testName}${XSS_PAYLOADS.imgOnerror}`;
    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousName);
      await shortWait(page, 1000);

      expect(alertListener.wasAlertTriggered()).toBe(false);
      await assertNoDangerousElements(page);

      if (result.success) {
        const exists = await verifyGenreExistsViaApi(testName);
        expect(exists, 'Genre should exist in database after creation').toBe(true);
        console.log('img onerror payload stored but rendered safely');
      } else {
        console.log('img onerror payload rejected by server:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should block SVG onload handler', async ({ page }) => {
    const testName = generateTestName('SvgTest');
    const maliciousName = `${testName}${XSS_PAYLOADS.svgOnload}`;
    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousName);
      await shortWait(page, 1000);

      expect(alertListener.wasAlertTriggered()).toBe(false);
      await assertNoDangerousElements(page);

      if (result.success) {
        const exists = await verifyGenreExistsViaApi(testName);
        expect(exists, 'Genre should exist in database after creation').toBe(true);
        console.log('SVG payload was stored but rendered safely');
      } else {
        console.log('SVG payload was rejected by server:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should block javascript: protocol in links', async ({ page }) => {
    const testName = generateTestName('LinkTest');
    const maliciousName = `${testName}${XSS_PAYLOADS.javascriptProtocol}`;
    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousName);
      await shortWait(page, 1000);

      expect(alertListener.wasAlertTriggered()).toBe(false);
      await assertNoDangerousElements(page);

      if (result.success) {
        const exists = await verifyGenreExistsViaApi(testName);
        expect(exists, 'Genre should exist in database after creation').toBe(true);
        console.log('javascript: protocol payload was stored but rendered safely');
      } else {
        console.log('javascript: protocol payload was rejected by server:', result.errorMessage);
      }
    } finally {
      alertListener.cleanup();
    }
  });

  test('should safely display XSS payload when viewing/editing genre', async ({ page }) => {
    const testName = generateTestName('EditTest');
    const maliciousName = `${testName}${XSS_PAYLOADS.scriptTag}`;
    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);
      const result = await createGenre(page, maliciousName);
      await shortWait(page, 1000);

      if (result.rejected) {
        console.log('Script payload was rejected by server — XSS prevented at input');
        expect(alertListener.wasAlertTriggered()).toBe(false);
        return;
      }

      // Genre was created — find and edit it
      const row = await findGenreRow(page, testName);
      if (!row) {
        // Genre may be on a different page — verify via API and skip DOM edit test
        const exists = await verifyGenreExistsViaApi(testName);
        expect(exists, 'Genre should exist in database').toBe(true);
        console.log('Genre exists in DB but not visible on current page — skipping edit test');
        expect(alertListener.wasAlertTriggered()).toBe(false);
        return;
      }

      // Click edit button in the genre row
      const editButton = row.locator('.item-table__btn-edit');
      await editButton.click();

      // Wait for dialog
      await page.waitForSelector('app-genre-editor-dialog app-dialog .dialog__overlay', {
        state: 'visible',
        timeout: 5000,
      });
      await shortWait(page, 1000);

      // Verify no alert triggered when viewing the XSS payload in the edit dialog
      expect(alertListener.wasAlertTriggered()).toBe(false);

      // Check input value exists
      const inputValue = await page.inputValue(
        'app-genre-editor-dialog app-text-input input.text-input__input'
      );
      expect(inputValue).toBeTruthy();
      console.log('Genre name in edit dialog:', inputValue);

      // Close dialog
      await page.click('app-genre-editor-dialog button.btn-secondary');
    } finally {
      alertListener.cleanup();
    }
  });

  test('comprehensive XSS prevention check with multiple payloads', async ({ page }) => {
    const testName = generateTestName('MultiTest');
    const alertListener = setupAlertListener(page);

    try {
      await navigateToGenresManagement(page);

      const payloadsToTest = [
        { payload: XSS_PAYLOADS.scriptTag, name: 'script tag' },
        { payload: XSS_PAYLOADS.imgOnerror, name: 'img onerror' },
        { payload: XSS_PAYLOADS.svgOnload, name: 'SVG onload' },
      ];

      for (let i = 0; i < payloadsToTest.length; i++) {
        const { payload, name } = payloadsToTest[i];
        const uniqueName = `${testName}_${i}`;
        const maliciousName = `${uniqueName}${payload}`;

        const result = await createGenre(page, maliciousName);
        await shortWait(page, 500);

        // No alert should fire regardless of outcome
        expect(alertListener.wasAlertTriggered(), `Alert triggered for payload: ${name}`).toBe(
          false
        );

        if (result.success) {
          const exists = await verifyGenreExistsViaApi(uniqueName);
          expect(exists, `Genre should exist for: ${uniqueName}`).toBe(true);
          console.log(`Payload "${name}": stored but rendered safely`);
        } else {
          console.log(`Payload "${name}": rejected by server`);
        }
      }

      // Final comprehensive check
      await assertNoDangerousElements(page);

      const messages = alertListener.getAlertMessages();
      if (messages.length > 0) {
        console.error('Unexpected alert messages:', messages);
      }

      expect(messages.length).toBe(0);
    } finally {
      alertListener.cleanup();
    }
  });
});
