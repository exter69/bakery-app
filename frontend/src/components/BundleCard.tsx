import { useI18n } from '../i18n';
import type { Bundle } from '../types/bundle';
import './BundleCard.css';

export interface BundleCardProps {
  bundle: Bundle;
  userLatitude?: number;
  userLongitude?: number;
  onReserve: (bundleId: string) => void;
  reserveLoading?: boolean;
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
  const R = 6371000; // Earth radius in meters
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

export function BundleCard({
  bundle,
  userLatitude,
  userLongitude,
  onReserve,
  reserveLoading = false,
}: BundleCardProps) {
  const { t } = useI18n();
  const isSoldOut = bundle.status === 'sold_out' || bundle.quantityRemaining === 0;

  const distance =
    userLatitude != null && userLongitude != null
      ? haversineDistance(userLatitude, userLongitude, bundle.bakeryLatitude, bundle.bakeryLongitude)
      : null;

  const cardClassName = `bundle-card${isSoldOut ? ' bundle-card--sold-out' : ''}`;

  return (
    <article className={cardClassName} aria-label={bundle.name}>
      <div className="bundle-card__image-container">
        <img
          src={bundle.photoUrl}
          alt={bundle.name}
          className="bundle-card__photo"
          loading="lazy"
        />
        {isSoldOut ? (
          <span className="bundle-card__badge bundle-card__badge--sold-out" aria-label={t('bundles.status.soldOut')}>
            {t('bundles.status.soldOut')}
          </span>
        ) : (
          <span className="bundle-card__badge bundle-card__badge--stock" aria-label={t('bundles.stock.remaining').replace('{n}', String(bundle.quantityRemaining))}>
            {t('bundles.stock.remaining').replace('{n}', String(bundle.quantityRemaining))}
          </span>
        )}
      </div>

      <div className="bundle-card__body">
        <div className="bundle-card__header">
          <h3 className="bundle-card__name">{bundle.name}</h3>
          <span
            className={`bundle-card__type ${bundle.type === 'surprise' ? 'bundle-card__type--surprise' : 'bundle-card__type--compose'}`}
            aria-label={bundle.type === 'surprise' ? t('bundles.type.surprise') : t('bundles.type.compose')}
          >
            {bundle.type === 'surprise' ? t('bundles.type.surprise') : t('bundles.type.compose')}
          </span>
        </div>

        <div className="bundle-card__meta">
          <span className="bundle-card__bakery">{bundle.bakeryName}</span>
          {distance != null && (
            <span className="bundle-card__distance">{formatDistance(distance)}</span>
          )}
          <span className="bundle-card__pickup">
            retrait{' '}
            <span className="bundle-card__pickup-time">
              {bundle.pickupStartTime}–{bundle.pickupEndTime}
            </span>
          </span>
        </div>

        {isSoldOut ? (
          <div className="bundle-card__sold-out-info">
            <p className="bundle-card__sold-out-bakery">{bundle.bakeryName}</p>
            {distance != null && (
              <p className="bundle-card__sold-out-distance">{formatDistance(distance)}</p>
            )}
            <p className="bundle-card__sold-out-message">{t('bundles.soldOutMessage')}</p>
          </div>
        ) : (
          <div className="bundle-card__contents">
            {bundle.type === 'compose' && bundle.items.length > 0 ? (
              <p className="bundle-card__items-list">
                {bundle.items.map((item, i) => (
                  <span key={i}>
                    {i > 0 && ', '}
                    {item.quantity}× {item.description}
                  </span>
                ))}
              </p>
            ) : bundle.type === 'surprise' ? (
              <p className="bundle-card__surprise-value">
                valeur estimée {formatPrice(bundle.estimatedValue)} — {bundle.description}
              </p>
            ) : null}
          </div>
        )}

        <div className="bundle-card__footer">
          <div className="bundle-card__pricing">
            <span className="bundle-card__original-price">{formatPrice(bundle.originalPrice)}</span>
            <span className="bundle-card__discounted-price">{formatPrice(bundle.discountedPrice)}</span>
          </div>

          <button
            className="bundle-card__reserve-btn"
            onClick={() => onReserve(bundle.id)}
            disabled={isSoldOut || reserveLoading}
            aria-label={`${t('bundles.action.reserve')} ${bundle.name}`}
          >
            {reserveLoading ? '...' : t('bundles.action.reserve')}
          </button>
        </div>
      </div>
    </article>
  );
}
