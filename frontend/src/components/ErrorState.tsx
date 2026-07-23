import './ErrorState.css';

interface ErrorStateProps {
  message: string;
  onRetry: () => void;
  retryLabel?: string;
}

export function ErrorState({ message, onRetry, retryLabel = 'Retry' }: ErrorStateProps) {
  return (
    <div className="error-state" role="alert">
      <svg
        className="error-state__icon"
        width="48"
        height="48"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
      <p className="error-state__message">{message}</p>
      <button type="button" className="error-state__retry-btn" onClick={onRetry}>
        {retryLabel}
      </button>
    </div>
  );
}
