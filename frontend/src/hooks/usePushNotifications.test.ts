import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { usePushNotifications } from './usePushNotifications';

// Mock the apiFetch module
vi.mock('../api/client', () => ({
  apiFetch: vi.fn(),
  getToken: vi.fn(() => 'test-token'),
}));

import { apiFetch } from '../api/client';

const mockApiFetch = apiFetch as ReturnType<typeof vi.fn>;

function createMockSubscription() {
  return {
    endpoint: 'https://push.example.com/sub1',
    toJSON: () => ({
      endpoint: 'https://push.example.com/sub1',
      keys: { p256dh: 'test-p256dh', auth: 'test-auth' },
    }),
    unsubscribe: vi.fn().mockResolvedValue(true),
  };
}

function setupPushEnvironment(opts: {
  permission?: NotificationPermission;
  existingSubscription?: ReturnType<typeof createMockSubscription> | null;
} = {}) {
  const { permission = 'default', existingSubscription = null } = opts;
  const mockSub = createMockSubscription();

  const mockPushManager = {
    getSubscription: vi.fn().mockResolvedValue(existingSubscription),
    subscribe: vi.fn().mockResolvedValue(mockSub),
  };

  const mockRegistration = { pushManager: mockPushManager };

  Object.defineProperty(navigator, 'serviceWorker', {
    value: {
      register: vi.fn().mockResolvedValue(mockRegistration),
      ready: Promise.resolve(mockRegistration),
    },
    writable: true,
    configurable: true,
  });

  Object.defineProperty(window, 'Notification', {
    value: {
      permission,
      requestPermission: vi.fn().mockResolvedValue('granted'),
    },
    writable: true,
    configurable: true,
  });

  Object.defineProperty(window, 'PushManager', {
    value: class {},
    writable: true,
    configurable: true,
  });

  return { mockPushManager, mockRegistration, mockSub };
}

describe('usePushNotifications', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('reports unsupported when serviceWorker is not in navigator', () => {
    // Simulate a browser without service worker support
    Object.defineProperty(navigator, 'serviceWorker', {
      value: undefined,
      writable: true,
      configurable: true,
    });
    Object.defineProperty(window, 'PushManager', {
      value: undefined,
      writable: true,
      configurable: true,
    });
    Object.defineProperty(window, 'Notification', {
      value: undefined,
      writable: true,
      configurable: true,
    });

    const { result } = renderHook(() => usePushNotifications());
    expect(result.current.permission).toBe('unsupported');
    expect(result.current.subscribed).toBe(false);
  });

  it('reports initial permission state from Notification API', async () => {
    setupPushEnvironment({ permission: 'denied' });

    const { result } = renderHook(() => usePushNotifications());

    await act(async () => {
      await new Promise((r) => setTimeout(r, 10));
    });

    expect(result.current.permission).toBe('denied');
  });

  it('detects existing subscription on mount', async () => {
    const sub = createMockSubscription();
    setupPushEnvironment({ existingSubscription: sub });

    const { result } = renderHook(() => usePushNotifications());

    await act(async () => {
      await new Promise((r) => setTimeout(r, 10));
    });

    expect(result.current.subscribed).toBe(true);
  });

  it('subscribes successfully', async () => {
    const { mockSub } = setupPushEnvironment();
    mockApiFetch.mockResolvedValueOnce({ publicKey: 'test-vapid-key' });
    mockApiFetch.mockResolvedValueOnce({ id: 'sub-123' });

    const { result } = renderHook(() => usePushNotifications());

    await act(async () => {
      await new Promise((r) => setTimeout(r, 10));
    });

    await act(async () => {
      await result.current.subscribe();
    });

    expect(result.current.subscribed).toBe(true);
    expect(result.current.permission).toBe('granted');
    expect(mockApiFetch).toHaveBeenCalledWith('/push/vapid-key');
    expect(mockApiFetch).toHaveBeenCalledWith('/user/push/subscribe', {
      method: 'POST',
      body: JSON.stringify({
        endpoint: mockSub.toJSON().endpoint,
        keys: { p256dh: 'test-p256dh', auth: 'test-auth' },
      }),
    });
  });

  it('does not subscribe when permission is denied', async () => {
    setupPushEnvironment();
    (window.Notification.requestPermission as ReturnType<typeof vi.fn>).mockResolvedValue('denied');

    const { result } = renderHook(() => usePushNotifications());

    await act(async () => {
      await new Promise((r) => setTimeout(r, 10));
    });

    await act(async () => {
      await result.current.subscribe();
    });

    expect(result.current.subscribed).toBe(false);
    expect(result.current.permission).toBe('denied');
    expect(mockApiFetch).not.toHaveBeenCalled();
  });

  it('unsubscribes successfully', async () => {
    const sub = createMockSubscription();
    setupPushEnvironment({ existingSubscription: sub });
    mockApiFetch.mockResolvedValueOnce(undefined);

    const { result } = renderHook(() => usePushNotifications());

    await act(async () => {
      await new Promise((r) => setTimeout(r, 10));
    });

    expect(result.current.subscribed).toBe(true);

    await act(async () => {
      await result.current.unsubscribe();
    });

    expect(result.current.subscribed).toBe(false);
    expect(mockApiFetch).toHaveBeenCalledWith('/user/push/unsubscribe', {
      method: 'DELETE',
      body: JSON.stringify({ endpoint: 'https://push.example.com/sub1' }),
    });
    expect(sub.unsubscribe).toHaveBeenCalled();
  });
});
