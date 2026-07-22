import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWebSocket } from './useWebSocket';

// Mock getToken from client module
vi.mock('../api/client', () => ({
  getToken: vi.fn(),
}));

import { getToken } from '../api/client';
const mockedGetToken = vi.mocked(getToken);

// Mock WebSocket
class MockWebSocket {
  static instances: MockWebSocket[] = [];
  url: string;
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  readyState = 0; // CONNECTING
  closeCalled = false;

  constructor(url: string) {
    this.url = url;
    MockWebSocket.instances.push(this);
  }

  close() {
    this.closeCalled = true;
    this.readyState = 3; // CLOSED
  }

  // Simulate server opening the connection
  simulateOpen() {
    this.readyState = 1; // OPEN
    this.onopen?.();
  }

  // Simulate server sending a message
  simulateMessage(data: string) {
    this.onmessage?.({ data });
  }

  // Simulate connection close
  simulateClose() {
    this.readyState = 3;
    this.onclose?.();
  }
}

describe('useWebSocket', () => {
  beforeEach(() => {
    MockWebSocket.instances = [];
    vi.stubGlobal('WebSocket', MockWebSocket);
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it('does not connect when no token is available', () => {
    mockedGetToken.mockReturnValue(null);

    renderHook(() => useWebSocket());

    expect(MockWebSocket.instances).toHaveLength(0);
  });

  it('connects with token as query param when token is available', () => {
    mockedGetToken.mockReturnValue('my-jwt-token');

    renderHook(() => useWebSocket());

    expect(MockWebSocket.instances).toHaveLength(1);
    expect(MockWebSocket.instances[0].url).toContain('token=my-jwt-token');
  });

  it('sets connected to true when WebSocket opens', async () => {
    mockedGetToken.mockReturnValue('token123');

    const { result } = renderHook(() => useWebSocket());

    expect(result.current.connected).toBe(false);

    act(() => {
      MockWebSocket.instances[0].simulateOpen();
    });

    expect(result.current.connected).toBe(true);
  });

  it('updates lastEvent when a message is received', () => {
    mockedGetToken.mockReturnValue('token123');

    const { result } = renderHook(() => useWebSocket());

    act(() => {
      MockWebSocket.instances[0].simulateOpen();
    });

    act(() => {
      MockWebSocket.instances[0].simulateMessage(
        JSON.stringify({ type: 'order_status', payload: { orderID: 'ord-1', status: 'ready' } })
      );
    });

    expect(result.current.lastEvent).toEqual({
      type: 'order_status',
      payload: { orderID: 'ord-1', status: 'ready' },
    });
  });

  it('ignores malformed messages without crashing', () => {
    mockedGetToken.mockReturnValue('token123');

    const { result } = renderHook(() => useWebSocket());

    act(() => {
      MockWebSocket.instances[0].simulateOpen();
    });

    act(() => {
      MockWebSocket.instances[0].simulateMessage('not valid json');
    });

    expect(result.current.lastEvent).toBeNull();
  });

  it('sets connected to false on close and attempts reconnect', () => {
    mockedGetToken.mockReturnValue('token123');

    const { result } = renderHook(() => useWebSocket());

    act(() => {
      MockWebSocket.instances[0].simulateOpen();
    });
    expect(result.current.connected).toBe(true);

    act(() => {
      MockWebSocket.instances[0].simulateClose();
    });
    expect(result.current.connected).toBe(false);

    // After 1 second (first backoff), should attempt reconnect
    act(() => {
      vi.advanceTimersByTime(1000);
    });

    expect(MockWebSocket.instances).toHaveLength(2);
  });

  it('uses exponential backoff on repeated disconnects', () => {
    mockedGetToken.mockReturnValue('token123');

    renderHook(() => useWebSocket());

    // First disconnect
    act(() => {
      MockWebSocket.instances[0].simulateClose();
    });

    // First reconnect at 1s
    act(() => { vi.advanceTimersByTime(1000); });
    expect(MockWebSocket.instances).toHaveLength(2);

    // Second disconnect
    act(() => {
      MockWebSocket.instances[1].simulateClose();
    });

    // Should wait 2s for second reconnect
    act(() => { vi.advanceTimersByTime(1000); });
    expect(MockWebSocket.instances).toHaveLength(2); // not yet

    act(() => { vi.advanceTimersByTime(1000); });
    expect(MockWebSocket.instances).toHaveLength(3); // now
  });

  it('closes WebSocket on unmount', () => {
    mockedGetToken.mockReturnValue('token123');

    const { unmount } = renderHook(() => useWebSocket());

    act(() => {
      MockWebSocket.instances[0].simulateOpen();
    });

    unmount();

    expect(MockWebSocket.instances[0].closeCalled).toBe(true);
  });

  it('resets reconnect counter after successful connection', () => {
    mockedGetToken.mockReturnValue('token123');

    renderHook(() => useWebSocket());

    // First disconnect + reconnect
    act(() => {
      MockWebSocket.instances[0].simulateClose();
    });
    act(() => { vi.advanceTimersByTime(1000); });

    // Successful reconnect
    act(() => {
      MockWebSocket.instances[1].simulateOpen();
    });

    // Disconnect again — backoff should reset to 1s
    act(() => {
      MockWebSocket.instances[1].simulateClose();
    });
    act(() => { vi.advanceTimersByTime(1000); });
    expect(MockWebSocket.instances).toHaveLength(3);
  });

  it('encodes token in URL', () => {
    mockedGetToken.mockReturnValue('token with spaces & special=chars');

    renderHook(() => useWebSocket());

    expect(MockWebSocket.instances[0].url).toContain(
      'token=' + encodeURIComponent('token with spaces & special=chars')
    );
  });
});
