import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { fetchBakeries } from '../api/bakeries';
import type { BakeryCard } from '../types/bakery';

export default function BakeryListPage() {
  const [bakeries, setBakeries] = useState<BakeryCard[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchBakeries()
      .then((res) => setBakeries(res.items))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="page-loading">Loading bakeries...</div>;
  }

  if (error) {
    return <div className="page-error">{error}</div>;
  }

  if (bakeries.length === 0) {
    return (
      <div className="empty-state">
        <p>No bakeries are currently listed.</p>
      </div>
    );
  }

  return (
    <div className="bakery-list">
      <h1>Bakeries</h1>
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
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
