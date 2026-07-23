import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import OAuthCallbackPage from './OAuthCallbackPage';

// Mock the API client
const mockApiFetch = vi.fn();
const mockSetToken = vi.fn();
const mockDecodeTokenRole = vi.fn();

vi.mock('../api/client', () => ({
  apiFetch: (...args: unknown[]) => mockApiFetch(...args),
  setToken: (...args: unknown[]) => mockSetToken(...args),
  decodeTokenRole: (...args: unknown[]) => mockDecodeTokenRole(...args),
}));

// Mock the i18n hook
vi.mock('../i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}));

// Mock useNavigate
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('OAuthCallbackPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('showsLoadingStateWhileExchangingCode', () => {
    mockApiFetch.mockReturnValue(new Promise(() => {})); // never resolves

    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc&provider=google']}>
        <OAuthCallbackPage />
      </MemoryRouter>
    );

    expect(screen.getByText('Signing you in...')).toBeInTheDocument();
  });

  it('showsErrorWhenNoCodeInURL', async () => {
    render(
      <MemoryRouter initialEntries={['/auth/callback']}>
        <OAuthCallbackPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('No authorization code received.')).toBeInTheDocument();
    });
  });

  it('showsErrorWhenAPICallFails', async () => {
    mockApiFetch.mockRejectedValue(new Error('OAuth exchange failed'));

    render(
      <MemoryRouter initialEntries={['/auth/callback?code=bad-code&provider=google']}>
        <OAuthCallbackPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('OAuth exchange failed')).toBeInTheDocument();
    });
    expect(screen.getByText('Back to login')).toBeInTheDocument();
  });

  it('redirectsToHomeAfterSuccessfulLogin', async () => {
    mockApiFetch.mockResolvedValue({
      token: 'jwt-token-123',
      user: { id: '1', username: 'user@test.com', role: 2 },
    });
    mockDecodeTokenRole.mockReturnValue(2);

    render(
      <MemoryRouter initialEntries={['/auth/callback?code=valid&provider=google']}>
        <OAuthCallbackPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(mockSetToken).toHaveBeenCalledWith('jwt-token-123');
      expect(mockNavigate).toHaveBeenCalledWith('/', { replace: true });
    });
  });

  it('redirectsToDashboardForSellerRole', async () => {
    mockApiFetch.mockResolvedValue({
      token: 'jwt-seller-token',
      user: { id: '2', username: 'baker@test.com', role: 1 },
    });
    mockDecodeTokenRole.mockReturnValue(1);

    render(
      <MemoryRouter initialEntries={['/auth/callback?code=valid&provider=google']}>
        <OAuthCallbackPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(mockSetToken).toHaveBeenCalledWith('jwt-seller-token');
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true });
    });
  });

  it('usesStateParamAsRedirectTarget', async () => {
    mockApiFetch.mockResolvedValue({
      token: 'jwt-token',
      user: { id: '3', username: 'user@test.com', role: 2 },
    });
    mockDecodeTokenRole.mockReturnValue(2);

    render(
      <MemoryRouter initialEntries={['/auth/callback?code=valid&provider=google&state=/bakeries']}>
        <OAuthCallbackPage />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/bakeries', { replace: true });
    });
  });
});
