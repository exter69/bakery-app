import { useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { isAuthenticated } from '../api/client';
import { useI18n } from '../i18n';
import {
  useBundles,
  useBundleWebSocket,
  useReserveBundle,
  useConfirmReservation,
  useCancelBundleReservation,
} from '../hooks/useBundles';
import { BundleCard } from '../components/BundleCard';
import { BundleMapView } from '../components/BundleMapView';
import { ReservationRail } from '../components/ReservationRail';
import { ImpactCard } from '../components/ImpactCard';
import type { Bundle, BundleFilters, BundleReservation, BundleType } from '../types/bundle';
import './BundlePage.css';

type ViewMode = 'list' | 'map';
type FilterKey = 'all' | 'pickupBefore19' | 'nearMe' | 'surprise' | 'compose';

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

export default function BundlePage() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [activeFilters, setActiveFilters] = useState<Set<FilterKey>>(new Set(['all']));
  const [userLatitude, setUserLatitude] = useState<number | undefined>(undefined);
  const [userLongitude, setUserLongitude] = useState<number | undefined>(undefined);
  const [geoAvailable, setGeoAvailable] = useState(false);

  // Reservation state
  const [activeReservation, setActiveReservation] = useState<BundleReservation | null>(null);
  const [reservingBundleId, setReservingBundleId] = useState<string | null>(null);

  // Mutation hooks
  const reserveBundle = useReserveBundle();
  const confirmReservation = useConfirmReservation();
  const cancelBundleReservation = useCancelBundleReservation();

  // Request geolocation on mount
  useEffect(() => {
    if (!navigator.geolocation) return;

    navigator.geolocation.getCurrentPosition(
      (position) => {
        setUserLatitude(position.coords.latitude);
        setUserLongitude(position.coords.longitude);
        setGeoAvailable(true);
      },
      () => {
        setGeoAvailable(false);
      }
    );
  }, []);

  // Build API filters from active filter set
  const apiFilters: BundleFilters = useMemo(() => {
    const filters: BundleFilters = {};

    if (activeFilters.has('surprise')) {
      filters.type = 'surprise' as BundleType;
    } else if (activeFilters.has('compose')) {
      filters.type = 'compose' as BundleType;
    }

    if (activeFilters.has('pickupBefore19')) {
      filters.pickupBefore = '19:00';
    }

    if (activeFilters.has('nearMe')) {
      filters.maxDistance = 500;
    }

    return filters;
  }, [activeFilters]);

  const { data, loading, error, refetch } = useBundles(apiFilters);

  // WebSocket: refetch on stock updates or expired events
  useBundleWebSocket({
    onStockUpdate: () => refetch(),
    onExpired: () => refetch(),
  });

  const toggleFilter = useCallback((key: FilterKey) => {
    setActiveFilters((prev) => {
      const next = new Set(prev);

      if (key === 'all') {
        return new Set(['all']);
      }

      next.delete('all');

      // Surprise and compose are mutually exclusive
      if (key === 'surprise') next.delete('compose');
      if (key === 'compose') next.delete('surprise');

      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }

      if (next.size === 0) {
        return new Set(['all']);
      }

      return next;
    });
  }, []);

  // Client-side filtering (distance) and sorting
  const bundles: Bundle[] = useMemo(() => {
    if (!data?.items) return [];

    let items = [...data.items];

    // Client-side distance filter
    if (activeFilters.has('nearMe') && userLatitude != null && userLongitude != null) {
      items = items.filter((bundle) => {
        const dist = haversineDistance(userLatitude, userLongitude, bundle.bakeryLatitude, bundle.bakeryLongitude);
        return dist <= 500;
      });
    }

    // Sort: by proximity if geolocation available, otherwise by publishedDate
    if (geoAvailable && userLatitude != null && userLongitude != null) {
      items.sort((a, b) => {
        const distA = haversineDistance(userLatitude, userLongitude, a.bakeryLatitude, a.bakeryLongitude);
        const distB = haversineDistance(userLatitude, userLongitude, b.bakeryLatitude, b.bakeryLongitude);
        return distA - distB;
      });
    } else {
      items.sort((a, b) => b.publishedDate.localeCompare(a.publishedDate));
    }

    return items;
  }, [data, activeFilters, userLatitude, userLongitude, geoAvailable]);

  // Find the bundle associated with the active reservation (for ReservationRail)
  const reservedBundle = useMemo(() => {
    if (!activeReservation) return null;
    return bundles.find((b) => b.id === activeReservation.bundleId) ?? null;
  }, [activeReservation, bundles]);

  const handleReserve = useCallback(async (bundleId: string) => {
    // Redirect unauthenticated users to login instead of calling the API
    if (!isAuthenticated()) {
      navigate('/login', { state: { from: '/paniers-du-soir' } });
      return;
    }
    setReservingBundleId(bundleId);
    try {
      const reservation = await reserveBundle.mutate(bundleId);
      setActiveReservation(reservation);
      refetch();
    } catch (err) {
      // 409 = sold out race condition
      const isConflict = err instanceof Error && err.message.includes('409');
      if (isConflict) {
        alert(t('bundles.error.unavailable'));
      }
      refetch();
    } finally {
      setReservingBundleId(null);
    }
  }, [reserveBundle, refetch, t, navigate]);

  const handleConfirm = useCallback(async () => {
    if (!activeReservation) return;
    try {
      await confirmReservation.mutate(activeReservation.bundleId);
      setActiveReservation(null);
    } catch {
      // Error state is managed by the hook
    }
  }, [activeReservation, confirmReservation]);

  const handleCancel = useCallback(async () => {
    if (!activeReservation) return;
    try {
      await cancelBundleReservation.mutate(activeReservation.id);
      setActiveReservation(null);
      refetch();
    } catch {
      // Error state is managed by the hook
    }
  }, [activeReservation, cancelBundleReservation, refetch]);

  return (
    <div className="bundle-page">
      {/* Header */}
      <header className="bundle-page__header">
        <div className="bundle-page__title-row">
          <h1 className="bundle-page__title">{t('bundles.title')}</h1>
          <span className="bundle-page__badge">{t('bundles.badge.antiWaste')}</span>
        </div>
        <p className="bundle-page__subtitle">{t('bundles.subtitle')}</p>
      </header>

      {/* View toggle */}
      <div className="bundle-page__view-toggle" role="tablist" aria-label="View mode">
        <button
          role="tab"
          className={`bundle-page__view-btn${viewMode === 'list' ? ' bundle-page__view-btn--active' : ''}`}
          onClick={() => setViewMode('list')}
          aria-selected={viewMode === 'list'}
        >
          {t('bundles.view.list')}
        </button>
        <button
          role="tab"
          className={`bundle-page__view-btn${viewMode === 'map' ? ' bundle-page__view-btn--active' : ''}`}
          onClick={() => setViewMode('map')}
          aria-selected={viewMode === 'map'}
        >
          {t('bundles.view.map')}
        </button>
      </div>

      {/* Filter bar */}
      <div className="bundle-page__filters" role="toolbar" aria-label="Bundle filters">
        <button
          className={`bundle-page__filter-btn${activeFilters.has('all') ? ' bundle-page__filter-btn--active' : ''}`}
          onClick={() => toggleFilter('all')}
          aria-pressed={activeFilters.has('all')}
        >
          {t('bundles.filter.all')}
        </button>
        <button
          className={`bundle-page__filter-btn${activeFilters.has('pickupBefore19') ? ' bundle-page__filter-btn--active' : ''}`}
          onClick={() => toggleFilter('pickupBefore19')}
          aria-pressed={activeFilters.has('pickupBefore19')}
        >
          {t('bundles.filter.pickupBefore19')}
        </button>
        <button
          className={`bundle-page__filter-btn${activeFilters.has('nearMe') ? ' bundle-page__filter-btn--active' : ''}`}
          onClick={() => toggleFilter('nearMe')}
          disabled={!geoAvailable}
          aria-pressed={activeFilters.has('nearMe')}
        >
          {t('bundles.filter.nearMe')}
        </button>
        <button
          className={`bundle-page__filter-btn${activeFilters.has('surprise') ? ' bundle-page__filter-btn--active' : ''}`}
          onClick={() => toggleFilter('surprise')}
          aria-pressed={activeFilters.has('surprise')}
        >
          {t('bundles.filter.surprise')}
        </button>
        <button
          className={`bundle-page__filter-btn${activeFilters.has('compose') ? ' bundle-page__filter-btn--active' : ''}`}
          onClick={() => toggleFilter('compose')}
          aria-pressed={activeFilters.has('compose')}
        >
          {t('bundles.filter.compose')}
        </button>
      </div>

      {/* Content area */}
      <div className="bundle-page__content">
        <div className="bundle-page__main">
          {loading && (
            <div className="bundle-page__loading">Loading...</div>
          )}

          {error && (
            <div className="bundle-page__error" role="alert">{error}</div>
          )}

          {!loading && !error && bundles.length === 0 && (
            <div className="bundle-page__empty">
              <p className="bundle-page__empty-text">
                No bundles available right now. Check back later!
              </p>
            </div>
          )}

          {!loading && !error && bundles.length > 0 && (
            <>
              {viewMode === 'list' ? (
                <div className="bundle-page__grid">
                  {bundles.map((bundle) => (
                    <BundleCard
                      key={bundle.id}
                      bundle={bundle}
                      userLatitude={userLatitude}
                      userLongitude={userLongitude}
                      onReserve={handleReserve}
                      reserveLoading={reservingBundleId === bundle.id}
                    />
                  ))}
                </div>
              ) : (
                <BundleMapView bundles={bundles} />
              )}
            </>
          )}

          {/* Impact card */}
          <div className="bundle-page__impact" aria-label="Community impact">
            <ImpactCard />
          </div>
        </div>

        {/* Sidebar: ReservationRail rendered when a reservation is active */}
        {activeReservation && reservedBundle && (
          <div className="bundle-page__sidebar">
            <ReservationRail
              reservation={activeReservation}
              bundle={reservedBundle}
              onConfirm={handleConfirm}
              onCancel={handleCancel}
              confirmLoading={confirmReservation.loading}
              cancelLoading={cancelBundleReservation.loading}
            />
          </div>
        )}
      </div>

      {/* Footer */}
      <footer className="bundle-page__footer">
        <p className="bundle-page__footer-note">{t('bundles.footer')}</p>
      </footer>
    </div>
  );
}
