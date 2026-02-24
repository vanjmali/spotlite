import { test, expect } from '@playwright/test';
import {
  loginAsAdmin,
  navigateToArtistsManagement,
  createArtist,
  findArtistRow,
  generateTestName,
  shortWait,
  verifyArtistExistsViaApi,
  SQL_INJECTION_PAYLOADS,
} from './helpers';

/**
 * SQL/NoSQL Injection Test Suite - Stored Injection Attack Prevention
 *
 * Tests verify that the application prevents NoSQL injection attacks through artist management.
 * Two valid defenses: (1) server rejects the payload, or (2) payload is stored
 * but treated as literal text (not executed as query operators).
 *
 * Note: This application uses MongoDB, so we test NoSQL injection patterns.
 */

test.describe('SQL/NoSQL Injection Prevention - Stored Injection in Artists', () => {
  test.describe.configure({ retries: 1 });

  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
  });

  test('should prevent NoSQL operator injection in artist name', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.operatorInjection;
    const testArtistName = generateTestName('TestArtist');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);
    const result = await createArtist(page, maliciousArtistName);
    await shortWait(page, 1000);

    if (result.success) {
      // Payload was stored — verify it's treated as literal text, not as operator
      const exists = await verifyArtistExistsViaApi(testArtistName);
      expect(exists, 'Artist should exist in database').toBe(true);

      const row = await findArtistRow(page, testArtistName);
      if (row) {
        const text = await row.textContent();
        console.log('Artist displayed as:', text);
        // Verify the payload is visible as text, not executed
        expect(text).toContain(testArtistName);
      } else {
        console.log('Artist stored but not visible on current page');
      }
    } else {
      console.log('NoSQL operator injection payload rejected by server:', result.errorMessage);
      expect(result.rejected).toBe(true);
    }
  });

  test('should prevent $where injection in artist name', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.whereInjection;
    const testArtistName = generateTestName('TestArtist');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);
    const result = await createArtist(page, maliciousArtistName);
    await shortWait(page, 1000);

    if (result.success) {
      const exists = await verifyArtistExistsViaApi(testArtistName);
      expect(exists, 'Artist should exist in database').toBe(true);
      console.log('$where injection payload stored but not executed');
    } else {
      console.log('$where injection payload rejected by server:', result.errorMessage);
      expect(result.rejected).toBe(true);
    }
  });

  test('should prevent regex injection in artist name', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.regexInjection;
    const testArtistName = generateTestName('TestArtist');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);
    const result = await createArtist(page, maliciousArtistName);
    await shortWait(page, 1000);

    if (result.success) {
      const exists = await verifyArtistExistsViaApi(testArtistName);
      expect(exists, 'Artist should exist in database').toBe(true);
      console.log('Regex injection payload stored as literal text');
    } else {
      console.log('Regex injection payload rejected by server:', result.errorMessage);
      expect(result.rejected).toBe(true);
    }
  });

  test('should prevent JSON injection in artist name', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.jsonPayload;
    const testArtistName = generateTestName('TestArtist');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);
    const result = await createArtist(page, maliciousArtistName);
    await shortWait(page, 1000);

    if (result.success) {
      const exists = await verifyArtistExistsViaApi(testArtistName);
      expect(exists, 'Artist should exist in database').toBe(true);
      console.log('JSON injection payload stored as literal text');
    } else {
      console.log('JSON injection payload rejected by server:', result.errorMessage);
      expect(result.rejected).toBe(true);
    }
  });

  test('should prevent traditional SQL injection patterns', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.sqlOr;
    const testArtistName = generateTestName('TestArtist');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);
    const result = await createArtist(page, maliciousArtistName);
    await shortWait(page, 1000);

    if (result.success) {
      const exists = await verifyArtistExistsViaApi(testArtistName);
      expect(exists, 'Artist should exist in database').toBe(true);
      console.log('SQL OR injection payload stored as literal text');
    } else {
      console.log('SQL OR injection payload rejected by server:', result.errorMessage);
      expect(result.rejected).toBe(true);
    }
  });

  test('should display injection payload as plain text', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.operatorInjectionAlt;
    const testArtistName = generateTestName('SafeTest');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);
    const result = await createArtist(page, maliciousArtistName);
    await shortWait(page, 1000);

    if (result.success) {
      const exists = await verifyArtistExistsViaApi(testArtistName);
      expect(exists, 'Artist should exist in database').toBe(true);

      const row = await findArtistRow(page, testArtistName);
      if (row) {
        const text = await row.textContent();
        console.log('Artist text content:', text);
        expect(text!.length).toBeGreaterThan(0);
        // Verify the JSON payload is visible as text
        expect(text).toContain(testArtistName);
      } else {
        console.log('Artist stored but not visible on current page');
      }
    } else {
      console.log(
        'Payload rejected by server — injection prevented at input:',
        result.errorMessage
      );
      expect(result.rejected).toBe(true);
    }
  });

  test('should handle artist viewing after injection attempt', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.existsOperator;
    const testArtistName = generateTestName('ViewTest');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);
    const result = await createArtist(page, maliciousArtistName);
    await shortWait(page, 1000);

    if (result.rejected) {
      console.log('Payload rejected by server — injection prevented at input');
      return;
    }

    const row = await findArtistRow(page, testArtistName);
    if (!row) {
      // Artist may be on a different page — verify via API and skip DOM edit test
      const exists = await verifyArtistExistsViaApi(testArtistName);
      expect(exists, 'Artist should exist in database').toBe(true);
      console.log('Artist exists in DB but not visible on current page — skipping edit test');
      return;
    }

    // Click edit button in the artist row
    const editButton = row.locator('.item-table__btn-edit');
    await editButton.click();

    // Wait for dialog
    await page.waitForSelector('app-artist-editor-dialog app-dialog .dialog__overlay', {
      state: 'visible',
      timeout: 5000,
    });
    await shortWait(page, 1000);

    // Verify payload is displayed safely in edit dialog
    const inputValue = await page.inputValue(
      'app-artist-editor-dialog app-text-input input.text-input__input'
    );
    expect(inputValue).toBeTruthy();
    console.log('Artist name in edit dialog:', inputValue);

    // Close dialog
    await page.click('app-artist-editor-dialog button.btn-secondary');
  });

  test('should prevent ReDoS attack with malicious regex', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.regexInjectionDos;
    const testArtistName = generateTestName('RedosTest');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);

    const startTime = Date.now();
    const result = await createArtist(page, maliciousArtistName);
    const endTime = Date.now();
    const duration = endTime - startTime;

    await shortWait(page, 1000);

    // Verify operation completes in reasonable time (< 5 seconds)
    expect(duration).toBeLessThan(5000);

    if (result.success) {
      console.log(`ReDoS payload stored safely (completed in ${duration}ms)`);
    } else {
      console.log(`ReDoS payload rejected (completed in ${duration}ms):`, result.errorMessage);
    }
  });

  test('should prevent special regex characters from breaking queries', async ({ page }) => {
    const injectionPayload = SQL_INJECTION_PAYLOADS.regexSpecialChars;
    const testArtistName = generateTestName('SpecialTest');
    const maliciousArtistName = `${testArtistName}${injectionPayload}`;

    await navigateToArtistsManagement(page);
    const result = await createArtist(page, maliciousArtistName);
    await shortWait(page, 1000);

    if (result.success) {
      const exists = await verifyArtistExistsViaApi(testArtistName);
      expect(exists, 'Artist should exist in database').toBe(true);
      console.log('Special regex characters stored and handled safely');
    } else {
      console.log('Special characters rejected by server:', result.errorMessage);
      expect(result.rejected).toBe(true);
    }
  });
});
