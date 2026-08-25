/**
 * Token Manager - إدارة موحدة للـ token في جميع أنحاء التطبيق
 * يحل مشكلة التضارب بين مفاتيح localStorage المختلفة
 */

export class TokenManager {
  private static readonly AUTH_TOKEN_KEY = 'auth_token';
  private static readonly TOKEN_KEY = 'token'; // For backwards compatibility

  /**
   * الحصول على الـ token من أي من المفاتيح
   */
  static getToken(): string | null {
    // Check both keys for backwards compatibility
    return (
      localStorage.getItem(this.AUTH_TOKEN_KEY) || 
      localStorage.getItem(this.TOKEN_KEY)
    );
  }

  /**
   * حفظ الـ token في كلا المفتاحين للتوافقية
   */
  static setToken(token: string): void {
    localStorage.setItem(this.AUTH_TOKEN_KEY, token);
    localStorage.setItem(this.TOKEN_KEY, token);
  }

  /**
   * حذف الـ token من كلا المفتاحين
   */
  static clearToken(): void {
    localStorage.removeItem(this.AUTH_TOKEN_KEY);
    localStorage.removeItem(this.TOKEN_KEY);
  }

  /**
   * التحقق من وجود token صالح
   */
  static hasValidToken(): boolean {
    const token = this.getToken();
    if (!token) return false;
    
    // Basic JWT validation - check if it has 3 parts
    const parts = token.split('.');
    return parts.length === 3;
  }

  /**
   * الحصول على header الـ Authorization
   */
  static getAuthHeader(): { Authorization: string } | {} {
    const token = this.getToken();
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  /**
   * تحديث جميع المكونات عند تغيير الـ token
   */
  static onTokenChange(callback: (token: string | null) => void): void {
    // Event listener for storage changes
    window.addEventListener('storage', (e) => {
      if (e.key === this.AUTH_TOKEN_KEY || e.key === this.TOKEN_KEY) {
        callback(this.getToken());
      }
    });
  }
}

// Export singleton instance
export const tokenManager = TokenManager;
