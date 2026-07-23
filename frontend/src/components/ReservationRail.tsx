import { useI18n } from '../i18n';
import type { Bundle, BundleReservation } from '../types/bundle';
import './ReservationRail.css';

export interface ReservationRailProps {
  reservation: BundleReservation;
  bundle: Bundle;
  onConfirm: () => void;
  onCancel: () => void;
  confirmLoading?: boolean;
  cancelLoading?: boolean;
}

/** Format a price in cents to euros string (e.g., 350 → "€3.50") */
function formatPrice(cents: number): string {
  return `€${(cents / 100).toFixed(2)}`;
}

export function ReservationRail({
  reservation,
  bundle,
  onConfirm,
  onCancel,
  confirmLoading = false,
  cancelLoading = false,
}: ReservationRailProps) {
  const { t } = useI18n();

  return (
    <aside className="reservation-rail" aria-label={t('bundles.reservation.title')}>
      <h2 className="reservation-rail__title">{t('bundles.reservation.title')}</h2>

      <div className="reservation-rail__bundle-info">
        <p className="reservation-rail__bundle-name">
          1× {bundle.name}
        </p>
        <p className="reservation-rail__price">
          {formatPrice(bundle.discountedPrice)}
        </p>
      </div>

      <p className="reservation-rail__pickup">
        Retrait {bundle.pickupStartTime}–{bundle.pickupEndTime} · {t('bundles.payment')}
      </p>

      <p className="reservation-rail__warning" role="alert">
        {t('bundles.warning.pickup')} {bundle.pickupEndTime}
      </p>

      <button
        className="reservation-rail__confirm-btn"
        onClick={onConfirm}
        disabled={confirmLoading || reservation.status !== 'pending'}
        aria-label={`${t('bundles.action.confirm')} ${bundle.name}`}
      >
        {confirmLoading ? '...' : `${t('bundles.action.confirm')} →`}
      </button>

      <button
        className="reservation-rail__cancel-btn"
        onClick={onCancel}
        disabled={cancelLoading}
        aria-label={`Annuler réservation ${bundle.name}`}
      >
        {cancelLoading ? '...' : 'Annuler'}
      </button>
    </aside>
  );
}
