import { useI18n } from '../i18n';
import type { Bundle } from '../types/bundle';

export interface BundleMapViewProps {
  bundles: Bundle[];
  onMarkerClick?: (bundleId: string) => void;
}

/**
 * BundleMapView — map view showing bundle bakery locations.
 *
 * This is a placeholder implementation. For production, integrate
 * Leaflet (react-leaflet) or Mapbox GL JS to render an interactive
 * map with markers at each bakery's coordinates.
 *
 * Each marker would display the bundle's bakery position and,
 * on click, scroll to or highlight the corresponding BundleCard.
 */
export function BundleMapView({ bundles, onMarkerClick }: BundleMapViewProps) {
  const { t } = useI18n();

  return (
    <div
      className="bundle-map-view"
      role="region"
      aria-label={t('bundles.view.map')}
      style={{
        width: '100%',
        minHeight: '400px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '1rem',
        background: 'var(--bg-panel)',
        border: 'var(--border)',
        borderRadius: 'var(--radius-card)',
        padding: '2rem',
      }}
    >
      <span
        style={{
          fontSize: '2.5rem',
        }}
        aria-hidden="true"
      >
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><polygon points="1 6 1 22 8 18 16 22 23 18 23 2 16 6 8 2 1 6"/><line x1="8" y1="2" x2="8" y2="18"/><line x1="16" y1="6" x2="16" y2="22"/></svg>
      </span>
      <p
        style={{
          fontFamily: 'var(--font-hand-body)',
          fontSize: '1.1rem',
          color: 'var(--ink-muted)',
          margin: 0,
          textAlign: 'center',
        }}
      >
        {t('bundles.view.map')} — coming soon
      </p>
      <p
        style={{
          fontFamily: 'var(--font-hand-body)',
          fontSize: '0.9rem',
          color: 'var(--ink-muted)',
          margin: 0,
        }}
      >
        {bundles.length} {bundles.length === 1 ? 'bundle' : 'bundles'}
      </p>

      {/* Render hidden marker buttons for accessibility / future integration */}
      <div style={{ display: 'none' }}>
        {bundles.map((bundle) => (
          <button
            key={bundle.id}
            onClick={() => onMarkerClick?.(bundle.id)}
            aria-label={`${bundle.bakeryName} — ${bundle.name}`}
          >
            {bundle.bakeryName}
          </button>
        ))}
      </div>
    </div>
  );
}
