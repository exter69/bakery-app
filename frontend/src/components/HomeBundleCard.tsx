import { Link } from 'react-router-dom';
import { useI18n } from '../i18n';
import type { Bundle } from '../types/bundle';
import './HomeBundleCard.css';

export interface HomeBundleCardProps {
  bundles: Bundle[];
  userLatitude?: number;
  userLongitude?: number;
}

/**
 * Compute distance between two geographic coordinates using the Haversine formula.
 * Returns distance in meters.
 */
function haversineDistance(
  lat1: number,
  lon1: number,
  lat2: number,
  lon2: number
): number {
  const R = 6371000;
  const toRad = (deg: number) => (deg * Math.PI) / 180;
  const dLat = toRad(lat2 - lat1);
  const dLon = toRad(lon2 - lon1);
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(toRad(lat1)) *
      Math.cos(toRad(lat2)) *
      Math.sin(dLon / 2) *
      Math.sin(dLon / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
}

/** Format a price in cents to euros string (e.g., 350 → "€3.50") */
function formatPrice(cents: number): string {
  return `€${(cents / 100).toFixed(2)}`;
}

/** Format distance in meters to a human-readable string */
function formatDistance(meters: number): string {
  if (meters < 1000) {
    return `${Math.round(meters)} m`;
  }
  return `${(meters / 1000).toFixed(1)} km`;
}

/**
 * Home page card showing available surplus bundles.
 * Displays nearest bundle expanded with full details + up to 3 compact bundles.
 * Does not render when no bundles are available.
 */
export function HomeBundleCard({ bundles, userLatitude, userLongitude }: HomeBundleCardProps) {
  const { t } = useI18n();

  // Only show published, non-sold-out bundles
  const availableBundles = bundles.filter(
    (b) => b.status === 'published' && b.quantityRemaining > 0
  );

  if (availableBundles.length === 0) {
    return null;
  }

  // Sort by distance if geolocation available
  const sorted = [...availableBundles];
  if (userLatitude != null && userLongitude != null) {
    sorted.sort((a, b) => {
      const distA = haversineDistance(userLatitude, userLongitude, a.bakeryLatitude, a.bakeryLongitude);
      const distB = haversineDistance(userLatitude, userLongitude, b.bakeryLatitude, b.bakeryLongitude);
      return distA - distB;
    });
  }

  const nearest = sorted[0];
  const additional = sorted.slice(1, 4); // up to 3 more

  const nearestDistance =
    userLatitude != null && userLongitude != null
      ? haversineDistance(userLatitude, userLongitude, nearest.bakeryLatitude, nearest.bakeryLongitude)
      : null;

  return (
    <section className="home-bundle-card" aria-label={t('bundles.home.title')}>
      {/* Header */}
      <div className="home-bundle-card__header">
        <h2 className="home-bundle-card__title">{t('bundles.home.title')}</h2>
        <span className="home-bundle-card__badge">{t('bundles.badge.antiWaste')}</span>
      </div>
      <p className="home-bundle-card__subtitle">{t('bundles.home.subtitle')}</p>

      {/* Nearest bundle — expanded */}
      <div className="home-bundle-card__expanded">
        <img
          src={nearest.photoUrl}
          alt={nearest.name}
          className="home-bundle-card__expanded-photo"
          loading="lazy"
        />
        <div className="home-bundle-card__expanded-body">
          <div className="home-bundle-card__expanded-meta">
            <span className="home-bundle-card__expanded-bakery">{nearest.bakeryName}</span>
            {nearestDistance != null && (
              <span className="home-bundle-card__expanded-distance">{formatDistance(nearestDistance)}</span>
            )}
            <span className="home-bundle-card__expanded-pickup">
              {nearest.pickupStartTime}–{nearest.pickupEndTime}
            </span>
          </div>
          <span className="home-bundle-card__expanded-stock">
            {t('bundles.stock.remaining').replace('{n}', String(nearest.quantityRemaining))}
          </span>

          {/* Contents */}
          <div className="home-bundle-card__expanded-contents">
            {nearest.type === 'compose' && nearest.items.length > 0 ? (
              <p className="home-bundle-card__items-list">
                {nearest.items.map((item, i) => (
                  <span key={i}>
                    {i > 0 && ', '}
                    {item.quantity}× {item.description}
                  </span>
                ))}
              </p>
            ) : nearest.type === 'surprise' ? (
              <p className="home-bundle-card__surprise-value">
                {nearest.description}
              </p>
            ) : null}
          </div>

          {/* Pricing */}
          <div className="home-bundle-card__expanded-pricing">
            <span className="home-bundle-card__original-price">{formatPrice(nearest.originalPrice)}</span>
            <span className="home-bundle-card__discounted-price">{formatPrice(nearest.discountedPrice)}</span>
          </div>

          <Link to="/paniers-du-soir" className="home-bundle-card__reserve-btn">
            {t('bundles.action.reserve')}
          </Link>
        </div>
      </div>

      {/* Additional bundles — compact rows */}
      {additional.length > 0 && (
        <div className="home-bundle-card__compact-list">
          {additional.map((bundle) => {
            const dist =
              userLatitude != null && userLongitude != null
                ? haversineDistance(userLatitude, userLongitude, bundle.bakeryLatitude, bundle.bakeryLongitude)
                : null;

            return (
              <div key={bundle.id} className="home-bundle-card__compact-row">
                <img
                  src={bundle.photoUrl}
                  alt={bundle.name}
                  className="home-bundle-card__compact-photo"
                  loading="lazy"
                />
                <div className="home-bundle-card__compact-info">
                  <span className="home-bundle-card__compact-bakery">{bundle.bakeryName}</span>
                  <span className="home-bundle-card__compact-meta">
                    {dist != null && <>{formatDistance(dist)} · </>}
                    {bundle.type === 'surprise' ? t('bundles.type.surprise') : t('bundles.type.compose')}
                    {' · '}
                    {bundle.pickupStartTime}–{bundle.pickupEndTime}
                  </span>
                </div>
                <div className="home-bundle-card__compact-pricing">
                  <span className="home-bundle-card__original-price">{formatPrice(bundle.originalPrice)}</span>
                  <span className="home-bundle-card__discounted-price">{formatPrice(bundle.discountedPrice)}</span>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* See all link */}
      <Link to="/paniers-du-soir" className="home-bundle-card__see-all">
        {t('bundles.home.seeAll')}
      </Link>
    </section>
  );
}
