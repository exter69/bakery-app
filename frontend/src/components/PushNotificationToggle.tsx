import { usePushNotifications } from '../hooks/usePushNotifications';
import './PushNotificationToggle.css';

/**
 * PushNotificationToggle provides a UI for opting in/out of Web Push notifications.
 * Hidden when push is unsupported or when the user has permanently denied permission.
 */
export function PushNotificationToggle() {
  const { permission, subscribed, subscribe, unsubscribe, loading } = usePushNotifications();

  // Don't render if push is unsupported or permanently denied
  if (permission === 'unsupported' || permission === 'denied') {
    return null;
  }

  return (
    <div className="push-toggle" role="group" aria-label="Push notifications">
      {!subscribed ? (
        <button
          className="push-toggle__button"
          onClick={subscribe}
          disabled={loading}
          aria-label="Enable push notifications"
        >
          🔔 {loading ? 'Enabling…' : 'Enable notifications'}
        </button>
      ) : (
        <button
          className="push-toggle__button push-toggle__button--active"
          onClick={unsubscribe}
          disabled={loading}
          aria-label="Disable push notifications"
        >
          🔕 {loading ? 'Disabling…' : 'Notifications on'}
        </button>
      )}
    </div>
  );
}
