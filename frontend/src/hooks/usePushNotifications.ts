import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../api/client';

type PushPermission = NotificationPermission | 'unsupported';

interface UsePushNotificationsReturn {
  /** Current notification permission state, or 'unsupported' if Push API is unavailable */
  permission: PushPermission;
  /** Whether the user currently has an active push subscription */
  subscribed: boolean;
  /** Subscribe to push notifications (requests permission if needed) */
  subscribe: () => Promise<void>;
  /** Unsubscribe from push notifications */
  unsubscribe: () => Promise<void>;
  /** Whether a subscribe/unsubscribe operation is in progress */
  loading: boolean;
}

/**
 * Hook for managing Web Push notification subscriptions.
 * Handles service worker registration, VAPID key fetching, and
 * subscription lifecycle with the backend.
 */
export function usePushNotifications(): UsePushNotificationsReturn {
  const [permission, setPermission] = useState<PushPermission>(getInitialPermission);
  const [subscribed, setSubscribed] = useState(false);
  const [loading, setLoading] = useState(false);

  // Check existing subscription on mount
  useEffect(() => {
    if (!isPushSupported()) return;

    navigator.serviceWorker.ready.then((registration) => {
      registration.pushManager.getSubscription().then((sub) => {
        setSubscribed(sub !== null);
      });
    });
  }, []);

  const subscribe = useCallback(async () => {
    if (!isPushSupported()) return;

    setLoading(true);
    try {
      // Request notification permission
      const result = await Notification.requestPermission();
      setPermission(result);
      if (result !== 'granted') return;

      // Register service worker
      const registration = await navigator.serviceWorker.register('/sw.js', { scope: '/' });
      await navigator.serviceWorker.ready;

      // Fetch VAPID public key from backend
      const { publicKey } = await apiFetch<{ publicKey: string }>('/push/vapid-key');

      // Subscribe to push via the browser's Push API
      const pushSubscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(publicKey),
      });

      const subJSON = pushSubscription.toJSON();

      // Send subscription to our backend
      await apiFetch('/user/push/subscribe', {
        method: 'POST',
        body: JSON.stringify({
          endpoint: subJSON.endpoint,
          keys: {
            p256dh: subJSON.keys?.p256dh ?? '',
            auth: subJSON.keys?.auth ?? '',
          },
        }),
      });

      setSubscribed(true);
    } finally {
      setLoading(false);
    }
  }, []);

  const unsubscribe = useCallback(async () => {
    if (!isPushSupported()) return;

    setLoading(true);
    try {
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.getSubscription();

      if (subscription) {
        // Notify backend first
        await apiFetch('/user/push/unsubscribe', {
          method: 'DELETE',
          body: JSON.stringify({ endpoint: subscription.endpoint }),
        });

        // Unsubscribe from browser
        await subscription.unsubscribe();
      }

      setSubscribed(false);
    } finally {
      setLoading(false);
    }
  }, []);

  return { permission, subscribed, subscribe, unsubscribe, loading };
}

/** Get the initial notification permission state */
function getInitialPermission(): PushPermission {
  if (!isPushSupported()) return 'unsupported';
  return Notification.permission;
}

/** Check if the Push API is available in this browser */
function isPushSupported(): boolean {
  return (
    typeof window !== 'undefined' &&
    'serviceWorker' in navigator &&
    navigator.serviceWorker !== undefined &&
    'PushManager' in window &&
    'Notification' in window
  );
}

/**
 * Convert a URL-safe base64 VAPID key to a Uint8Array for the Push API.
 */
function urlBase64ToUint8Array(base64String: string): ArrayBuffer {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; i++) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray.buffer as ArrayBuffer;
}
