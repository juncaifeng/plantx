export const TOKEN_KEY = 'plantx_token';

export interface TokenClaims {
  sub: string;
  tenant_id: string;
  roles?: string[];
  permissions?: string[];
  username?: string;
  preferred_username?: string;
  name?: string;
  email?: string;
}

function base64UrlDecode(input: string): string {
  const base64 = input.replace(/-/g, '+').replace(/_/g, '/');
  const pad = base64.length % 4;
  const padded = pad ? base64 + '='.repeat(4 - pad) : base64;
  const raw = atob(padded);
  return decodeURIComponent(
    raw
      .split('')
      .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
      .join('')
  );
}

export function decodeJwt(token: string): TokenClaims | null {
  try {
    const payload = token.split('.')[1];
    if (!payload) return null;
    return JSON.parse(base64UrlDecode(payload)) as TokenClaims;
  } catch {
    return null;
  }
}
