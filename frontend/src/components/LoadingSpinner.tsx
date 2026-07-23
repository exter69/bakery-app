/**
 * Full-page loading spinner used as the Suspense fallback for lazy-loaded routes.
 * Re-uses the existing .spinner class defined in project CSS.
 */
export default function LoadingSpinner() {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '40vh',
      }}
    >
      <div className="spinner" aria-label="Loading" role="status" />
    </div>
  );
}
