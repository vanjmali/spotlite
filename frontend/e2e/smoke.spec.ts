import { test, expect } from '@playwright/test';

/**
 * Smoke Test - Verify test infrastructure is working
 *
 * This is a simple test to verify that:
 * 1. The application is accessible
 * 2. Playwright can navigate and interact with pages
 * 3. The test environment is properly configured
 */

test.describe('Smoke Test - XSS Testing Infrastructure', () => {
  test('should load the application homepage', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Verify the page loaded
    expect(page.url()).toContain('localhost');

    // Check that the page has a title
    const title = await page.title();
    expect(title).toBeTruthy();

    console.log('✓ Application loaded successfully');
    console.log('  URL:', page.url());
    console.log('  Title:', title);
  });

  test('should be able to navigate to login page', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // Verify we're on the login page
    expect(page.url()).toContain('login');

    // Check for email and password inputs
    const emailInput = page.locator('input[type="email"]');
    const passwordInput = page.locator('input[type="password"]');

    await expect(emailInput).toBeVisible();
    await expect(passwordInput).toBeVisible();

    console.log('✓ Login page loaded successfully');
    console.log('  Email input found:', await emailInput.count());
    console.log('  Password input found:', await passwordInput.count());
  });

  test('should not have any XSS vulnerabilities in page title', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Get the page title
    const title = await page.title();

    // Verify title doesn't contain script tags or malicious content
    expect(title).not.toMatch(/<script/i);
    expect(title).not.toMatch(/onerror=/i);
    expect(title).not.toMatch(/javascript:/i);

    console.log('✓ Page title is safe from XSS');
  });

  test('environment check - verify test configuration', async ({ page }) => {
    console.log('Test Configuration:');
    console.log('  Base URL:', 'https://localhost:4443');
    console.log('  Browser:', await page.context().browser()?.version());
    console.log('  Viewport:', await page.viewportSize());

    // This test always passes - it's just for logging configuration
    expect(true).toBe(true);
  });
});
