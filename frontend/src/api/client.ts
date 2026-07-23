export const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

// --- JWT Token Management ---

const TOKEN_KEY = 'auth_token';
const GUEST_MODE_KEY = 'guest_mode';

/** Store the JWT token in localStorage */
export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.removeItem(GUEST_MODE_KEY);
}

/** Retrieve the JWT token from localStorage */
export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

/** Remove the JWT token (logout) */
export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

/** Check if the user currently has a token that is not expired */
export function isAuthenticated(): boolean {
  const token = getToken();
  if (!token) return false;
  const payload = decodeTokenPayload(token);
  if (!payload) return false;
  // If exp claim exists and is a number, check it against current time
  if (typeof payload.exp === 'number') {
    const nowSeconds = Math.floor(Date.now() / 1000);
    if (nowSeconds >= payload.exp) {
      clearToken();
      return false;
    }
  }
  return true;
}

/** Enable guest mode */
export function setGuestMode(): void {
  localStorage.setItem(GUEST_MODE_KEY, 'true');
}

/** Check if the user is in guest mode */
export function isGuestMode(): boolean {
  return localStorage.getItem(GUEST_MODE_KEY) === 'true';
}

/** Clear guest mode */
export function clearGuestMode(): void {
  localStorage.removeItem(GUEST_MODE_KEY);
}

// --- JWT Decode Utility ---

/**
 * Safely decode a base64url-encoded string to a standard string.
 * Translates URL-safe alphabet (-/_ → +//) and adds padding before calling atob().
 */
export function decodeBase64Url(base64url: string): string {
  const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
  const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4);
  return atob(padded);
}

export interface TokenPayload {
  role?: number;
  exp?: number;
  sub?: string;
  [key: string]: unknown;
}

/**
 * Decode a JWT token's payload segment into a typed object.
 * Returns null if the token is malformed or cannot be parsed.
 */
export function decodeTokenPayload(token: string): TokenPayload | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const json = decodeBase64Url(parts[1]);
    const payload = JSON.parse(json);
    if (typeof payload !== 'object' || payload === null) return null;
    return payload as TokenPayload;
  } catch {
    return null;
  }
}

/**
 * Decode the JWT payload and extract the role claim.
 * Returns the numeric role or null if decoding fails.
 */
export function decodeTokenRole(token: string): number | null {
  const payload = decodeTokenPayload(token);
  if (payload === null || typeof payload.role !== 'number') return null;
  return payload.role;
}

// --- API Fetch with Auth ---

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const token = getToken();

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    // In guest mode, don't clear token or dispatch event — public endpoints
    // won't return 401 anyway, but if they do, just ignore it gracefully.
    if (isGuestMode()) {
      throw new ApiError(401, 'UNAUTHORIZED', 'Authentication required. Please sign in.');
    }

    clearToken();
    // Dispatch a custom event so the app can react to auth failures
    window.dispatchEvent(new CustomEvent('auth:unauthorized'));
    throw new ApiError(401, 'UNAUTHORIZED', 'Authentication required. Please log in.');
  }

  if (!response.ok) {
    const body = await response.json().catch(() => ({ code: 'UNKNOWN', message: 'Request failed' }));
    throw new ApiError(response.status, body.code ?? 'UNKNOWN', body.message ?? 'Request failed');
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}
