import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { fetchBakeries } from '../api/bakeries';
import type { BakeryCard } from '../types/bakery';
import './HomePage.css';

export default function HomePage() {
  const [bakeries, setBakeries] = useState<BakeryCard[]>([]);

  useEffect(() => {
    fetchBakeries()
      .then((res) => setBakeries(res.items.slice(0, 4)))
      .catch(() => {
        /* silently ignore — section just won't show */
      });
  }, []);

  return (
    <div className="home-page">
      {/* Hero */}
      <section className="home-hero">
        <h1 className="home-hero__title">Fresh bakery, delivered to your door</h1>
        <p className="home-hero__subtitle">
          Choose your favorites, pick the day and time — we handle the rest. Or
          reserve for pickup at your local bakery.
        </p>
        <div className="home-hero__actions">
          <Link to="/bakeries" className="home-hero__btn home-hero__btn--primary">
            Start Ordering
          </Link>
          <a href="#how-it-works" className="home-hero__btn home-hero__btn--secondary">
            How It Works
          </a>
        </div>
      </section>

      {/* How It Works */}
      <section className="home-section" id="how-it-works">
        <h2 className="home-section__title">How It Works</h2>
        <div className="how-it-works">
          <div className="how-step">
            <div className="how-step__number">1</div>
            <h3 className="how-step__title">Browse &amp; Choose</h3>
            <p className="how-step__desc">
              Explore bakeries near you and pick your favorite breads, pastries,
              and treats.
            </p>
          </div>
          <div className="how-step">
            <div className="how-step__number">2</div>
            <h3 className="how-step__title">Schedule Delivery</h3>
            <p className="how-step__desc">
              Choose the day and time, and set it recurring if you want — weekly
              auto-order so you never have to think about it.
            </p>
          </div>
          <div className="how-step">
            <div className="how-step__number">3</div>
            <h3 className="how-step__title">Enjoy Fresh</h3>
            <p className="how-step__desc">
              Wake up to fresh bakery products at your door. It&apos;s that
              simple.
            </p>
          </div>
        </div>
      </section>

      {/* Delivery & Pickup */}
      <section className="home-section">
        <h2 className="home-section__title">Our Services</h2>
        <div className="feature-cards">
          <div className="feature-card">
            <div className="feature-card__icon">🚲</div>
            <h3 className="feature-card__title">Delivery — The Morning Ritual</h3>
            <p className="feature-card__desc">
              Scheduled delivery of bakery products straight to your door. Choose
              the time and day, make it recurring — weekly auto-order so you never
              have to think about it.
            </p>
            <p className="feature-card__note">
              Holiday mode coming soon — pause when you&apos;re away.
            </p>
          </div>
          <div className="feature-card">
            <div className="feature-card__icon">🏪</div>
            <h3 className="feature-card__title">Pickup — Skip the Queue</h3>
            <p className="feature-card__desc">
              Reserve your items and pick them up at the bakery, no waiting.
              Browse the menu, select what you want, and it&apos;ll be ready when
              you arrive.
            </p>
          </div>
        </div>
      </section>

      {/* Nearby Bakeries */}
      {bakeries.length > 0 && (
        <section className="home-section">
          <div className="nearby-header">
            <h2 className="nearby-header__title">Bakeries Near You</h2>
            <Link to="/bakeries" className="nearby-header__link">
              View All →
            </Link>
          </div>
          <div className="nearby-grid">
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
                  <h3 className="bakery-card__name">{bakery.name}</h3>
                  <p className="bakery-card__schedule">
                    {bakery.todaySchedule.isOpen
                      ? `Open ${bakery.todaySchedule.openTime} – ${bakery.todaySchedule.closeTime}`
                      : 'Closed today'}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
