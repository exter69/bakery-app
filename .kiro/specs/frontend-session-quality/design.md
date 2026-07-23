# Design: Frontend Session & Dashboard Quality

## Overview

This design addresses six interrelated quality issues in the frontend:

1. **JWT base64url decode bug** — The `decodeTokenRole` function in `client.ts` calls `atob()` directly on base64url payloads without translating `-`/`_` to `+`/`/` or adding padding. Tokens whose payloads happen to contain these characters throw, causing role to be `null` and `RoleRoute` to redirect legitimate users.

2. **No client-side expiry check** — `isAuthenticated()` only verifies token presence in localStorage. Expired tokens are discovered only on the first 401 response.

3. **TypeScript strict mode off** — `tsconfig.app.json` lacks `strict`, `strictNullChecks`, and `noImplicitAny`, weakening the type safety guarantee.

4. **Role magic numbers** — Routes use `[0, 1]` and `[3]` literals; the mapping lives only in developer memory.

5. **Dashboard i18n bypass** — Seven dashboard pages have hardcoded French/English strings instead of using the `t()` function; the translation keys already exist at 372-key parity across EN/FR/NL.

6. **DashboardOverview "today" stats** — Orders are fetched with `status=confirmed` but no date filter, so "today's orders" may include past confirmed orders.

## Architecture

All changes are confined to the frontend React/TypeScript codebase. No backend changes are required.

```
frontend/src/
  api/
    client.ts          ← Fix JWT decode, add expiry check, export auth hook
  auth/
    roles.ts           ← NEW: Role enum
    useAuth.ts         ← NEW: Reactive auth context hook
    AuthProvider.tsx   ← NEW: Context provider wrapping the app
  components/
    RoleRoute.tsx      ← Use Role enum + auth context
  pages/dashboard/
    *.tsx              ← Add useI18n, replace hardcoded strings
  i18n/
    translations.ts   ← Add missing dashboard keys
```

## Components and Interfaces

### 1. JWT Decode Utility (refactored)

**File:** `frontend/src/api/client.ts`

The existing `decodeTokenRole` function is replaced with a safer `decodeTokenPayload` that:
- Translates base64url to standard base64 (replace `-` with `+`, `_` with `/`)
- Pads to a multiple of 4 with `=`
- Calls `atob()` on the safe string
- Parses JSON and returns the full payload object (or `null`)

```typescript
/**
 * Safely decode a base64url-encoded JWT payload.
 * Handles the URL-safe alphabet and missing padding.
 */
function decodeBase64Url(base64url: string): string {
  const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
  const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4);
  return atob(padded);
}

interface TokenPayload {
  role?: number;
  exp?: number;
  sub?: string;
  [key: string]: unknown;
}

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

export function decodeTokenRole(token: string): number | null {
  const payload = decodeTokenPayload(token);
  if (payload === null || typeof payload.role !== 'number') return null;
  return payload.role;
}
```

### 2. Token Expiry Check

**File:** `frontend/src/api/client.ts`

```typescript
/** Check if token exists and is not expired */
export function isAuthenticated(): boolean {
  const token = getToken();
  if (!token) return false;
  const payload = decodeTokenPayload(token);
  if (!payload) return false;
  // If exp claim exists and is a number, check it
  if (typeof payload.exp === 'number') {
    const nowSeconds = Math.floor(Date.now() / 1000);
    if (nowSeconds >= payload.exp) {
      clearToken();
      return false;
    }
  }
  return true;
}
```

### 3. Role Enum

**File:** `frontend/src/auth/roles.ts`

```typescript
export enum UserRole {
  Admin = 0,
  Seller = 1,
  Customer = 2,
  B2B = 3,
}
```

### 4. Reactive Auth Context

**File:** `frontend/src/auth/useAuth.ts` and `AuthProvider.tsx`

A lightweight context that exposes:
- `isLoggedIn: boolean`
- `role: UserRole | null`
- `login(token: string): void`
- `logout(): void`

This replaces direct localStorage reads in `RoleRoute` with a context subscription that re-renders on auth state change.

### 5. RoleRoute Update

```typescript
import { UserRole } from '../auth/roles';
import { useAuth } from '../auth/useAuth';

interface RoleRouteProps {
  children: React.ReactNode;
  allowedRoles: UserRole[];
  fallback?: string;
}

export default function RoleRoute({ children, allowedRoles, fallback = '/' }: RoleRouteProps) {
  const { isLoggedIn, role } = useAuth();
  if (!isLoggedIn) return <Navigate to="/login" replace />;
  if (role === null || !allowedRoles.includes(role)) return <Navigate to={fallback} replace />;
  return <>{children}</>;
}
```

### 6. Dashboard i18n Integration

Each of the 7 non-i18n dashboard pages will:
1. Import `useI18n` from `../../i18n`
2. Replace hardcoded strings with `t('dashboard.section.key')` calls
3. Add corresponding keys to all three locale dictionaries

Pages affected: `DashboardSchedule`, `DashboardPayouts`, `DashboardBakery`, `DashboardProducts`, `DashboardOrders`, `DashboardBundles`, and remaining hardcoded strings in `DashboardOverview`.

### 7. TypeScript Strict Mode

**File:** `frontend/tsconfig.app.json`

Add to `compilerOptions`:
```json
"strict": true
```

This enables `strictNullChecks`, `noImplicitAny`, `strictFunctionTypes`, `strictBindCallApply`, `strictPropertyInitialization`, `noImplicitThis`, `useUnknownInCatchVariables`, and `alwaysStrict`. If full strict produces too many errors (>50), fall back to adding just `strictNullChecks` and `noImplicitAny` and document the remainder.

## Data Models

### TokenPayload Interface

```typescript
interface TokenPayload {
  role?: number;
  exp?: number;
  sub?: string;
  [key: string]: unknown;
}
```

### UserRole Enum

```typescript
enum UserRole {
  Admin = 0,
  Seller = 1,
  Customer = 2,
  B2B = 3,
}
```

### AuthState (context value)

```typescript
interface AuthState {
  isLoggedIn: boolean;
  role: UserRole | null;
  token: string | null;
  login: (token: string) => void;
  logout: () => void;
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: JWT Base64url Round-Trip Decode

*For any* valid JSON object, encoding it as base64url (using `-`, `_`, no padding) and passing the resulting string through `decodeBase64Url` then `JSON.parse` SHALL produce an object equal to the original.

**Validates: Requirements 1.1, 1.2**

### Property 2: Malformed Tokens Return Null

*For any* string that is not a valid 3-dot-separated JWT with a JSON-parseable base64url middle segment, `decodeTokenPayload` SHALL return `null` without throwing an exception.

**Validates: Requirements 1.4**

### Property 3: Expiry Determines Authentication

*For any* well-formed JWT token with a numeric `exp` claim, `isAuthenticated()` SHALL return `false` when `Date.now() / 1000 >= exp`, and `true` when `Date.now() / 1000 < exp`.

**Validates: Requirements 2.1, 2.2**

### Property 4: Translation Key Coverage

*For any* translation key referenced via `t()` in a dashboard page source file, that key SHALL exist in all three locale dictionaries (en, fr, nl) in `translations.ts`.

**Validates: Requirements 5.2, 5.3**

## Error Handling

- **JWT decode failure**: Returns `null` — never throws. Callers (RoleRoute, isAuthenticated) treat `null` as "not authenticated" and redirect to login.
- **Expired token on API call**: The existing `auth:unauthorized` 401 handler remains as a fallback. The client-side expiry check prevents most 401s from happening.
- **Strict mode type errors**: If enabling `strict: true` produces errors, they are fixed at the source (type narrowing, explicit types). No `@ts-ignore` suppressions allowed without a ticket reference.

## Testing Strategy

### Unit Tests (Vitest)

- `decodeBase64Url` — specific examples: standard payload, payload with `-`/`_`, payload needing 1/2 padding chars
- `decodeTokenPayload` — malformed inputs: missing dots, non-JSON, empty string, non-object JSON
- `isAuthenticated` — expired token, future token, missing exp, malformed token
- `UserRole` enum — values match expected numbers
- Dashboard pages — render tests verifying `t()` is called (no hardcoded visible text)

### Property-Based Tests (fast-check)

- **Property 1**: Generate random JSON payloads, encode as base64url, decode via `decodeBase64Url`, verify round-trip equality. Minimum 100 iterations.
- **Property 2**: Generate random malformed strings (missing dots, bad base64, non-JSON), verify `decodeTokenPayload` returns null. Minimum 100 iterations.
- **Property 3**: Generate tokens with random `exp` values and a mocked `Date.now`, verify `isAuthenticated` returns the correct boolean. Minimum 100 iterations.
- **Property 4**: Extract all `t('...')` calls from dashboard source files, verify each key exists in all three locales. Run once (deterministic but exhaustive).

### Smoke Test

- `tsc --noEmit` with strict mode enabled exits 0.

### Grep Check

- No hardcoded French/English user-facing strings remain in dashboard page files (CI lint step).
