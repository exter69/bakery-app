import { useEffect, useRef, useState, useCallback } from 'react';
import { getToken } from '../api/client';

export interface WSEvent {
  type: string;
  payload: Record<string, string>;
}

/**
 * useWebSocket connects to the backend WebSocket endpoint for real-time notifications.
 * Automatically reconnects with exponential backoff on disconnect.
 * Auth is passed via the JWT token as a query parameter.
 */
export function useWebSocket() {
  const [lastEvent, setLastEvent] = useState<WSEvent | null>(null);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<number>(0);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const unmountedRef = useRef(false);

  const connect = useCallback(() => {
    if (unmountedRef.current) return;

    const token = getToken();
    if (!token) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const url = `${protocol}//${host}/api/ws?token=${encodeURIComponent(token)}`;

    const socket = new WebSocket(url);
    wsRef.current = socket;

    socket.onopen = () => {
      if (unmountedRef.current) return;
      setConnected(true);
      reconnectRef.current = 0;
    };

    socket.onmessage = (e) => {
      if (unmountedRef.current) return;
      try {
        const event = JSON.parse(e.data) as WSEvent;
        setLastEvent(event);
      } catch {
        // Ignore malformed messages
      }
    };

    socket.onclose = () => {
      if (unmountedRef.current) return;
      setConnected(false);
      wsRef.current = null;

      // Exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s max
      const delay = Math.min(1000 * 2 ** reconnectRef.current, 30000);
      reconnectRef.current++;
      reconnectTimerRef.current = setTimeout(connect, delay);
    };

    socket.onerror = () => {
      // The close event will fire after this, triggering reconnect
      socket.close();
    };
  }, []);

  useEffect(() => {
    unmountedRef.current = false;
    connect();

    return () => {
      unmountedRef.current = true;
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect]);

  return { lastEvent, connected };
}
