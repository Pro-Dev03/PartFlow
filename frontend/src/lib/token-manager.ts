/**
 * Token Manager - manages session tokens in memory only.
 * Access and refresh tokens are never written to browser storage because the
 * backend uses HttpOnly refresh cookies for credential protection.
 */

export class TokenManager {
  private static readonly AUTH_TOKEN_KEY = 'auth_token';
  private static readonly TOKEN_KEY = 'token';
  private static readonly REFRESH_TOKEN_KEY = 'refresh_token';
  private static accessToken: string | null = null;
  private static refreshToken: string | null = null;
  private static cloudAccessToken: string | null = null;

  private static clearLegacyBrowserTokenStorage(): void {
    if (typeof window === 'undefined') return;

    [
      this.AUTH_TOKEN_KEY,
      this.TOKEN_KEY,
      this.REFRESH_TOKEN_KEY,
      'cloud_token',
      'cloud_refresh_token',
      'auth-storage',
    ].forEach((key) => {
      localStorage.removeItem(key);
    });
  }

  static getToken(): string | null {
    return this.accessToken;
  }

  static getRefreshToken(): string | null {
    return this.refreshToken;
  }

  static getCloudToken(): string | null {
    return this.cloudAccessToken;
  }

  static setToken(token: string): void {
    this.accessToken = token;
    this.clearLegacyBrowserTokenStorage();
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('partflow:token-change', { detail: { token } }));
    }
  }

  static setRefreshToken(refreshToken: string): void {
    this.refreshToken = refreshToken;
    this.clearLegacyBrowserTokenStorage();
  }

  static setCloudToken(cloudToken: string | null): void {
    this.cloudAccessToken = cloudToken ?? null;
    this.clearLegacyBrowserTokenStorage();
  }

  static clearToken(): void {
    this.accessToken = null;
    this.clearLegacyBrowserTokenStorage();
  }

  static clearRefreshToken(): void {
    this.refreshToken = null;
    this.clearLegacyBrowserTokenStorage();
  }

  static clearCloudToken(): void {
    this.cloudAccessToken = null;
    this.clearLegacyBrowserTokenStorage();
  }

  static hasValidToken(): boolean {
    const token = this.getToken();
    if (!token) return false;

    const parts = token.split('.');
    return parts.length === 3;
  }

  static getAuthHeader(): { Authorization: string } | {} {
    const token = this.getToken();
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  static onTokenChange(callback: (token: string | null) => void): void {
    if (typeof window === 'undefined') return;

    const handler = (event: Event) => {
      const detail = (event as CustomEvent<{ token?: string | null }>).detail;
      callback(detail?.token ?? this.getToken());
    };

    window.addEventListener('partflow:token-change', handler);
  }
}

export const tokenManager = TokenManager;
