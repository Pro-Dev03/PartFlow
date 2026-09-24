import type { Page } from '@playwright/test';

/**
 * PartFlow keeps access tokens in memory. Read the app's in-memory token
 * manager from the Vite test page instead of depending on legacy localStorage
 * keys that are deliberately cleared after login.
 */
export async function getBrowserAuthHeaders(page: Page): Promise<Record<string, string>> {
  return page.evaluate(async () => {
    const tokenManagerUrl = new URL('/src/lib/token-manager.ts', window.location.origin).href;
    const { TokenManager } = await import(tokenManagerUrl);
    const token = TokenManager.getToken();
    const cloudToken = TokenManager.getCloudToken();
    return {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(cloudToken ? { 'X-PartFlow-Cloud-Token': cloudToken } : {}),
    };
  });
}
