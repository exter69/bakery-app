import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { fetchBakeries } from '../api/bakeries';
import type { BakeryCard } from '../types/bakery';

export default function BakeriesPage() {
  const [bakeries, setBakeries] = useState<BakeryCard[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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
          // User denied geolocation or error — load without location
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
          <h1 className="bakeries-page__title">Bakeries</h1>
          <p className="bakeries-page__subtitle">
            Discover bakeries delivering near you
          </p>
        </div>
        <div className="bakeries-empty">
          <p className="bakeries-empty__text">
            No bakeries in your area yet 🥐 — Check back soon, new bakeries are
            joining every day!
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="bakeries-page">
      <div className="bakeries-page__header">
        <h1 className="bakeries-page__title">Bakeries</h1>
        <p className="bakeries-page__subtitle">
          Discover bakeries delivering near you
        </p>
      </div>
      <div className="bakery-grid">
        {bakeries.map((bakery) => (
          <Link
            key={bakery.id}
            to={`/bakeries/${bakery.id}`}
            className="bakery-card"
          >
            <img
              src={bakery.photoUrl}
              alt={bakery.name}
              className="bakery-card__photo"
              loading="lazy"
            />
            <div className="bakery-card__info">
              <h2 className="bakery-card__name">{bakery.name}</h2>
              <p className="bakery-card__schedule">
                {bakery.todaySchedule.isOpen
                  ? `Open ${bakery.todaySchedule.openTime} – ${bakery.todaySchedule.closeTime}`
                  : 'Closed today'}
              </p>
              {bakery.distance != null && (
                <p className="bakery-card__distance">
                  {bakery.distance < 1
                    ? `${Math.round(bakery.distance * 1000)} m away`
                    : `${bakery.distance} km away`}
                </p>
              )}
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
