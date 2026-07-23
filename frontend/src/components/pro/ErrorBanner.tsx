import './ErrorBanner.css';

interface ErrorBannerProps {
  /** Error message to display (in French) */
  message: string;
  /** Callback when the user clicks "Réessayer" */
  onRetry?: () => void;
}

/**
 * Shared inline error display with French messaging and retry button.
 * Used across all dashboard pages to satisfy Requirement 7.1.
 */
export function ErrorBanner({ message, onRetry }: ErrorBannerProps) {
  return (
    <div className="error-banner" role="alert">
      <span className="error-banner__message">{message}</span>
      {onRetry && (
        <button
          type="button"
          className="error-banner__retry"
          onClick={onRetry}
        >
          Réessayer
        </button>
      )}
    </div>
  );
}
