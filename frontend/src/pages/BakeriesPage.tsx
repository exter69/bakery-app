import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { fetchBakeries } from '../api/bakeries';
import { useI18n } from '../i18n';
import type { BakeryCard } from '../types/bakery';
import './BakeriesPage.css';

type FilterChip = 'open' | 'all' | 'nearby';

const DAY_LABELS = ['M', 'T', 'W', 'T', 'F', 'S', 'S'];

export default function BakeriesPage() {
  const [bakeries, setBakeries] = useState<BakeryCard[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterChip>('open');
  const { t } = useI18n();

  useEffect(() => {
    let cancelled = false;

    function loadBakeries(location?: { lat: number; lng: number }) {
      fetchBakeries(1, location)
        .then((res) => {
          if (!cancelled) setBakeries(res.items);
        })
        .catch((err: Error) => {
          if (!cancelled) setError(err.message);
        })
        .finally(() => {
          if (!cancelled) setLoading(false);
        });
    }

    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          loadBakeries({
            lat: position.coords.latitude,
            lng: position.coords.longitude,
          });
        },
        () => {
          loadBakeries();
        },
      );
    } else {
      loadBakeries();
    }

    return () => {
      cancelled = true;
    };
  }, []);

  const todayIndex = (new Date().getDay() + 6) % 7; // Monday = 0

  const filteredBakeries =
    filter === 'open'
      ? bakeries.filter((b) => b.todaySchedule.isOpen)
      : bakeries;

  if (loading) {
    return <div className="page-loading">Loading bakeries...</div>;
  }

  if (error) {
    return <div className="page-error">{error}</div>;
  }

  if (bakeries.length === 0) {
    return (
      <div className="bakeries-page">
        <div className="bakeries-page__header">
          <h1 className="bakeries-page__title">{t('bakeries.title')}</h1>
          <p className="bakeries-page__subtitle">
            {t('bakeries.subtitle')}
          </p>
        </div>
        <div className="bakeries-empty">
          <p className="bakeries-empty__text">
            {t('bakeries.empty')}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="bakeries-page">
      <div className="bakeries-page__header">
        <h1 className="bakeries-page__title">{t('bakeries.title')}</h1>
        <p className="bakeries-page__subtitle">
          {t('bakeries.subtitle')}
        </p>
      </div>

      {/* Filter chips (visible on mobile, hidden on desktop) */}
      <div className="bakeries-filters">
        <button
          className={`filter-chip${filter === 'open' ? ' filter-chip--active' : ''}`}
          onClick={() => setFilter('open')}
        >
          {t('bakeries.openNow')}
        </button>
        <button
          className={`filter-chip${filter === 'all' ? ' filter-chip--active' : ''}`}
          onClick={() => setFilter('all')}
        >
          {t('bakeries.all')}
        </button>
        <button
          className={`filter-chip${filter === 'nearby' ? ' filter-chip--active' : ''}`}
          onClick={() => setFilter('nearby')}
        >
          {t('bakeries.nearby')}
        </button>
      </div>

      {/* Desktop: card grid */}
      <div className="bakery-grid">
        {filteredBakeries.map((bakery) => (
          <Link
            key={bakery.id}
            to={`/bakeries/${bakery.id}`}
            className={`bakery-card${!bakery.todaySchedule.isOpen ? ' bakery-card--closed' : ''}`}
          >
            <img
              src={bakery.photoUrl}
              alt={bakery.name}
              className="bakery-card__photo"
              loading="lazy"
            />
            <div className="bakery-card__info">
              <h2 className="bakery-card__name">{bakery.name}</h2>
              <p className={`bakery-card__schedule${!bakery.todaySchedule.isOpen ? ' bakery-card__schedule--closed' : ''}`}>
                {bakery.todaySchedule.isOpen
                  ? `${t('common.open')} · ${bakery.todaySchedule.openTime} – ${bakery.todaySchedule.closeTime}`
                  : t('bakeries.closedToday')}
              </p>
              {bakery.distance != null && (
                <p className="bakery-card__distance">
                  {bakery.distance < 1
                    ? `${Math.round(bakery.distance * 1000)} m ${t('common.away')}`
                    : `${bakery.distance} km ${t('common.away')}`}
                </p>
              )}
            </div>
          </Link>
        ))}
      </div>

      {/* Mobile: ledger rows */}
      <div className="bakery-ledger">
        {filteredBakeries.map((bakery) => (
          <Link
            key={bakery.id}
            to={`/bakeries/${bakery.id}`}
            className={`ledger-row${!bakery.todaySchedule.isOpen ? ' ledger-row--closed' : ''}`}
          >
            <img
              src={bakery.photoUrl}
              alt={bakery.name}
              className="ledger-row__photo"
              loading="lazy"
            />
            <div className="ledger-row__center">
              <span className="ledger-row__name">{bakery.name}</span>
              {bakery.distance != null && (
                <span className="ledger-row__address">
                  {bakery.distance < 1
                    ? `${Math.round(bakery.distance * 1000)} m`
                    : `${bakery.distance} km`}
                </span>
              )}
              <div className="ledger-row__week">
                {DAY_LABELS.map((label, i) => (
                  <span
                    key={i}
                    className={`day-chip${i === todayIndex ? ' day-chip--today' : ''}`}
                  >
                    {label}
                  </span>
                ))}
              </div>
            </div>
            <div className="ledger-row__hours">
              {bakery.todaySchedule.isOpen ? (
                <span className="ledger-row__time">
                  {bakery.todaySchedule.openTime} – {bakery.todaySchedule.closeTime}
                </span>
              ) : (
                <span className="ledger-row__closed-label">{t('common.closed')}</span>
              )}
            </div>
          </Link>
        ))}
      </div>

      {filteredBakeries.length === 0 && (
        <div className="bakeries-empty">
          <p className="bakeries-empty__text">
            No bakeries match this filter 🥐
          </p>
        </div>
      )}
    </div>
  );
}
