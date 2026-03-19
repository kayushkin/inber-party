# E2E Testing Guide

This directory contains comprehensive end-to-end tests for the Inber Party application using Playwright.

## Test Structure

### Core Test Files

- **`core-flows.spec.ts`** - Original comprehensive tests covering navigation, interactions, and basic functionality
- **`improved-core-flows.spec.ts`** - Enhanced version with better error handling and reliability
- **`visual-regression.spec.ts`** - Screenshot comparison tests for UI consistency
- **`accessibility.spec.ts`** - Accessibility compliance tests using axe-core

### Test Categories

1. **Functional Tests**
   - Page loading and navigation
   - User interactions (clicks, forms, etc.)
   - Responsive design behavior
   - Error handling scenarios

2. **Visual Regression Tests**
   - Screenshot comparisons across different pages
   - Responsive layout verification
   - Theme consistency checks

3. **Accessibility Tests**
   - Keyboard navigation
   - ARIA labels and roles
   - Color contrast ratios
   - Screen reader compatibility

## Running Tests

### Prerequisites

1. Ensure the frontend development server is running:
   ```bash
   npm run dev
   ```

2. Install Playwright browsers (first time only):
   ```bash
   npx playwright install
   ```

### Test Commands

```bash
# Run all tests
npm run test:e2e

# Run tests with UI for debugging
npm run test:e2e:ui

# Run tests in headed mode (see browser)
npm run test:e2e:headed

# Run specific test file
npx playwright test tests/e2e/improved-core-flows.spec.ts

# Run tests for specific browser
npx playwright test --project=chromium

# Run in debug mode
npx playwright test --debug
```

### Configuration

The test configuration is in `playwright.config.ts`:

- **Base URL**: `http://localhost:5173`
- **Browsers**: Chromium, Firefox (WebKit disabled for WSL compatibility)
- **Timeout**: 30 seconds per test
- **Screenshots**: On failure only
- **Visual regression threshold**: 30% (tolerant for viewport differences)

## Test Strategy

### Handling Missing Backend

Many tests are designed to work even when the backend (inber server) is not running:

- Tests focus on frontend functionality and UI behavior
- Network errors from missing APIs are filtered out of error checking
- Tests verify graceful degradation when data is unavailable

### Visual Regression Test Considerations

Visual tests may fail if:
- Screen resolution differs from baseline
- Fonts are different (different OS/environment)
- Content loads at different speeds

To update visual baselines:
```bash
npx playwright test --update-snapshots
```

### Writing New Tests

Follow these patterns for reliable tests:

1. **Use robust selectors**:
   ```typescript
   // Good - multiple fallback options
   page.locator('nav, header, [role="navigation"]')
   
   // Avoid - too specific
   page.locator('.navbar-component .nav-item:nth-child(2)')
   ```

2. **Wait for content, not time**:
   ```typescript
   // Good - wait for actual content
   await page.waitForFunction(() => document.body.textContent.length > 50);
   
   // Avoid - arbitrary timeouts
   await page.waitForTimeout(5000);
   ```

3. **Handle expected errors gracefully**:
   ```typescript
   // Filter out network errors from missing backend
   const criticalErrors = errors.filter(error => 
     !error.includes('404') && 
     !error.includes('ECONNREFUSED')
   );
   ```

## Debugging Failed Tests

1. **Check the HTML report** generated after test runs:
   ```bash
   npx playwright show-report
   ```

2. **Use debug mode** to step through tests:
   ```bash
   npx playwright test --debug test-name
   ```

3. **Check screenshots** in `test-results/` directory for visual failures

4. **Review console logs** in test output for JavaScript errors

## Continuous Integration

Tests are configured to run with:
- Retries enabled (2 retries in CI)
- Single worker in CI for stability
- HTML reporter for easy debugging

## Common Issues

### "Connection Refused" Errors

If tests fail with connection errors:
1. Ensure `npm run dev` is running
2. Check that port 5173 is available
3. Wait for the dev server to fully start before running tests

### Visual Regression Failures

If screenshots don't match:
1. Check if viewport size changed
2. Update snapshots if intentional changes were made
3. Consider if test environment differs (fonts, OS, etc.)

### Flaky Tests

If tests are inconsistent:
1. Add better waiting conditions
2. Use more robust element selectors
3. Increase timeouts for slow environments
4. Consider if the test depends on external factors

## Maintenance

- **Update baselines** when UI intentionally changes
- **Review test coverage** as new features are added
- **Monitor test performance** and optimize slow tests
- **Keep dependencies updated** (Playwright, browsers)

Regular review of test failures helps maintain a reliable test suite and catch regressions early.