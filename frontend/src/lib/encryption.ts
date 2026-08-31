/**
 * Simple encryption/decryption for sensitive data using SubtleCrypto
 * This provides basic protection against casual inspection, not cryptographic strength protection
 */

const ENCRYPTION_ALGORITHM = 'AES-GCM';
const KEY_DERIVATION_ALGORITHM = 'PBKDF2';
const ITERATIONS = 100000;

/**
 * Generate a stable encryption key from a device identifier
 */
export async function getEncryptionKey(): Promise<CryptoKey> {
  // Get or create a device identifier
  let deviceId = localStorage.getItem('partflow-device-id');
  if (!deviceId) {
    deviceId = generateDeviceId();
    localStorage.setItem('partflow-device-id', deviceId);
  }

  // Use the device ID to derive a stable encryption key
  const encoder = new TextEncoder();
  const data = encoder.encode(deviceId);
  
  const baseKey = await window.crypto.subtle.importKey(
    'raw',
    data,
    { name: KEY_DERIVATION_ALGORITHM },
    false,
    ['deriveKey']
  );

  return window.crypto.subtle.deriveKey(
    {
      name: KEY_DERIVATION_ALGORITHM,
      salt: encoder.encode('partflow-cloud-connection'),
      iterations: ITERATIONS,
      hash: 'SHA-256',
    },
    baseKey,
    { name: ENCRYPTION_ALGORITHM, length: 256 },
    false,
    ['encrypt', 'decrypt']
  );
}

/**
 * Generate a unique device identifier
 */
function generateDeviceId(): string {
  return `device-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Encrypt a string (typically a PostgreSQL connection URL)
 */
export async function encrypt(text: string): Promise<string> {
  try {
    const key = await getEncryptionKey();
    const encoder = new TextEncoder();
    const data = encoder.encode(text);
    
    // Generate a random IV for each encryption
    const iv = window.crypto.getRandomValues(new Uint8Array(12));
    
    const encrypted = await window.crypto.subtle.encrypt(
      { name: ENCRYPTION_ALGORITHM, iv },
      key,
      data
    );

    // Combine IV and encrypted data
    const combined = new Uint8Array(iv.length + encrypted.byteLength);
    combined.set(iv);
    combined.set(new Uint8Array(encrypted), iv.length);

    // Convert to base64
    return btoa(String.fromCharCode(...combined));
  } catch (error) {
    console.error('Encryption failed:', error);
    throw new Error('Failed to encrypt data');
  }
}

/**
 * Decrypt a previously encrypted string
 */
export async function decrypt(encryptedText: string): Promise<string> {
  try {
    const key = await getEncryptionKey();
    
    // Decode from base64
    const combined = new Uint8Array(atob(encryptedText).split('').map(c => c.charCodeAt(0)));
    
    // Extract IV (first 12 bytes) and encrypted data (rest)
    const iv = combined.slice(0, 12);
    const encrypted = combined.slice(12);

    const decrypted = await window.crypto.subtle.decrypt(
      { name: ENCRYPTION_ALGORITHM, iv },
      key,
      encrypted
    );

    const decoder = new TextDecoder();
    return decoder.decode(decrypted);
  } catch (error) {
    console.error('Decryption failed:', error);
    throw new Error('Failed to decrypt data');
  }
}

/**
 * Store encrypted data in localStorage
 */
export async function storeEncrypted(key: string, value: string): Promise<void> {
  const encrypted = await encrypt(value);
  localStorage.setItem(key, encrypted);
}

/**
 * Retrieve encrypted data from localStorage
 */
export async function retrieveEncrypted(key: string): Promise<string | null> {
  const encrypted = localStorage.getItem(key);
  if (!encrypted) return null;
  
  try {
    return await decrypt(encrypted);
  } catch {
    // If decryption fails, clear the corrupted data
    localStorage.removeItem(key);
    return null;
  }
}

/**
 * Validate PostgreSQL connection string format
 */
export function isValidPostgresConnectionString(url: string): boolean {
  if (!url) return false;
  
  // Basic validation: should start with postgresql:// or postgres://
  if (!url.startsWith('postgresql://') && !url.startsWith('postgres://')) {
    return false;
  }

  // Should have @ separator between credentials and host
  if (!url.includes('@')) {
    return false;
  }

  // Should have host and port
  if (!url.match(/@[^/]+:\d+/)) {
    return false;
  }

  return true;
}

/**
 * Extract connection details from PostgreSQL URL
 */
export function parsePostgresUrl(url: string): {
  host?: string;
  port?: string;
  database?: string;
  user?: string;
} | null {
  try {
    const urlObj = new URL(url.replace(/^postgres(ql)?:\/\//, 'https://'));
    return {
      host: urlObj.hostname,
      port: urlObj.port,
      database: urlObj.pathname.replace(/^\//, '') || 'postgres',
      user: urlObj.username,
    };
  } catch {
    return null;
  }
}
