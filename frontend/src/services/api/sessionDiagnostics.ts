type DiagnosticFields = Record<string, string | number | boolean | undefined>;

/**
 * Temporary, opt-in auth tracing. It intentionally accepts only sanitized
 * metadata; tokens, cookies, request bodies, and server response bodies must
 * never be passed here.
 */
export function logSessionDiagnostic(event: string, fields: DiagnosticFields = {}): void {
  if (typeof window === 'undefined') return;

  let enabled = Boolean(import.meta.env.DEV);
  try {
    enabled ||= window.localStorage.getItem('partflow-session-diagnostics') === 'true';
  } catch {
    // Diagnostics are optional and must not interrupt session handling.
  }
  if (!enabled) return;

  const safeFields = Object.fromEntries(
    Object.entries(fields).filter(([, value]) => value !== undefined),
  );
  console.info('[PartFlow session]', {
    at: new Date().toISOString(),
    event,
    ...safeFields,
  });
}

export function getLogoutCaller(): string {
  const stack = new Error().stack?.split('\n').slice(2, 6).map((line) => line.trim());
  return stack?.join(' <- ') || 'unknown';
}
