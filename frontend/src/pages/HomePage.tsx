import { lazy, Suspense, useEffect, useState, useRef, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { fetchBakeries } from '../api/bakeries';
import { isAuthenticated, clearToken } from '../api/client';
import { useI18n } from '../i18n';
import { useBundles } from '../hooks/useBundles';
import Footer from '../components/Footer';
import { HomeBundleCard } from '../components/HomeBundleCard';
import LanguageSwitcher from '../components/LanguageSwitcher';
import { ThemeSwitcher } from '../components/ThemeSwitcher';
import type { BakeryCard } from '../types/bakery';
import './HomePage.css';

// Lazy-load the map component to defer the heavy leaflet bundle
const BakeryMap = lazy(() => import('../components/BakeryMap'));

export default function HomePage() {
  const navigate = useNavigate();
  const authenticated = isAuthenticated();
  const [bakeries, setBakeries] = useState<BakeryCard[]>([]);
  const [userLatitude, setUserLatitude] = useState<number | undefined>(undefined);
  const [userLongitude, setUserLongitude] = useState<number | undefined>(undefined);
  const { t } = useI18n();

  // Fetch bundles for the HomeBundleCard
  const { data: bundlesData } = useBundles({});

  // Hover indicator for nav links
  const heroLinksRef = useRef<HTMLDivElement>(null);
  const [hoverIndicator, setHoverIndicator] = useState<{ left: number; width: number } | null>(null);

  const handleLinkHover = useCallback((e: React.MouseEvent<HTMLAnchorElement>) => {
    if (!heroLinksRef.current) return;
    const containerRect = heroLinksRef.current.getBoundingClientRect();
    const linkRect = e.currentTarget.getBoundingClientRect();
    setHoverIndicator({
      left: linkRect.left - containerRect.left,
      width: linkRect.width,
    });
  }, []);

  const handleLinksLeave = useCallback(() => {
    setHoverIndicator(null);
  }, []);

  useEffect(() => {
    fetchBakeries()
      .then((res) => setBakeries(res.items.slice(0, 5)))
      .catch(() => {});
  }, []);

  // Request geolocation for HomeBundleCard distance display
  useEffect(() => {
    if (!navigator.geolocation) return;
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setUserLatitude(position.coords.latitude);
        setUserLongitude(position.coords.longitude);
      },
      () => {
        // Geolocation denied or unavailable — gracefully degrade
      }
    );
  }, []);

  function handleSignOut() {
    clearToken();
    navigate('/login');
  }

  const openBakeries = bakeries.filter((b) => b.todaySchedule.isOpen);

  return (
    <div className="home-page">
      {/* Hero with real bakery photo */}
      <section className="home-hero">
        <div className="home-hero__overlay" />

        {/* Floating pill navbar */}
        <nav className="pill-nav pill-nav--hero">
          <Link to="/" className="pill-nav__brand">Ma Boulangerie</Link>
          <div className="pill-nav__spacer" />
          <div className="pill-nav__links" ref={heroLinksRef} onMouseLeave={handleLinksLeave}>
            {hoverIndicator && (
              <span
                className="pill-nav__indicator"
                style={{ left: hoverIndicator.left, width: hoverIndicator.width }}
              />
            )}
            <Link to="/bakeries" className="pill-nav__link" onMouseEnter={handleLinkHover}>{t('nav.bakeries')}</Link>
            {authenticated && <Link to="/schedule" className="pill-nav__link" onMouseEnter={handleLinkHover}>{t('nav.orders')}</Link>}
            <Link to="/about" className="pill-nav__link" onMouseEnter={handleLinkHover}>{t('nav.about')}</Link>
          </div>
          <ThemeSwitcher />
          <LanguageSwitcher />
          <div className="pill-nav__actions">
            {authenticated ? (
              <button onClick={handleSignOut} className="pill-nav__btn pill-nav__btn--accent">{t('nav.signOut')}</button>
            ) : (
              <Link to="/login" className="pill-nav__btn pill-nav__btn--accent">{t('nav.signIn')}</Link>
            )}
          </div>
        </nav>

        <div className="home-hero__tagline">
          {t('home.tagline')}
        </div>
      </section>

      {/* Body */}
      <div className="home-body">
        {/* Top row: Entry cards (left) + Quick reserve starts here (right) */}
        <div className="home-top-row">
          {/* Left: entry cards + open now below */}
          <div className="home-top-row__left">
            {/* Entry cards — typographic style */}
            <div className="entry-cards">
              <Link to="/bakeries" className="entry-card">
                <span className="entry-card__eyebrow">LIVRAISON</span>
                <span className="entry-card__title">{t('home.orderDelivery')}</span>
                <span className="entry-card__subtitle">{t('home.payOnline')}</span>
                <span className="entry-card__cta">
                  <span>{t('bakeries.title')}</span>
                  <span>→</span>
                </span>
              </Link>
              <Link to="/bakeries" className="entry-card">
                <span className="entry-card__eyebrow">RETRAIT</span>
                <span className="entry-card__title">{t('home.reservePickup')}</span>
                <span className="entry-card__subtitle">{t('home.payAtCounter')}</span>
                <span className="entry-card__cta">
                  <span>{t('bakeries.title')}</span>
                  <span>→</span>
                </span>
              </Link>
            </div>

            {/* Open right now — in a card */}
            <section className="open-now-card">
              <h2 className="open-now-card__heading">{t('home.openNow')}</h2>
              {openBakeries.length > 0 ? (
                <ul className="open-now-card__list">
                  {openBakeries.slice(0, 5).map((bakery) => (
                    <li key={bakery.id}>
                      <Link to={`/bakeries/${bakery.id}`} className="open-now-card__row">
                        <img
                          src={bakery.photoUrl}
                          alt={bakery.name}
                          className="open-now-card__photo"
                          loading="lazy"
                        />
                        <div className="open-now-card__info">
                          <span className="open-now-card__name">{bakery.name}</span>
                          <span className="open-now-card__details">
                            {t('common.open')} {t('common.till')} {bakery.todaySchedule.closeTime}
                            {bakery.distance != null && (
                              <> · {bakery.distance < 1
                                ? `${Math.round(bakery.distance * 1000)}m`
                                : `${bakery.distance}km`}</>
                            )}
                          </span>
                        </div>
                        <span className="open-now-card__hours">
                          {bakery.todaySchedule.openTime} – {bakery.todaySchedule.closeTime}
                        </span>
                      </Link>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="open-now-card__empty">{t('home.noOpen')}</p>
              )}
              <Link to="/bakeries" className="open-now-card__see-all">
                {t('home.seeAll')}
              </Link>
            </section>

            {/* Map preview — lazy-loaded to defer leaflet bundle */}
            <section className="open-now-card" style={{ marginTop: '1.5rem' }}>
              <h2 className="open-now-card__heading">Bakeries near you</h2>
              <Suspense fallback={<div style={{ height: 320, display: 'flex', alignItems: 'center', justifyContent: 'center' }}><div className="spinner" aria-label="Loading map" role="status" /></div>}>
                <BakeryMap bakeries={bakeries} />
              </Suspense>
            </section>

            {/* Surplus bundles — HomeBundleCard */}
            {bundlesData && bundlesData.items.length > 0 && (
              <div style={{ marginTop: '1.5rem' }}>
                <HomeBundleCard
                  bundles={bundlesData.items}
                  userLatitude={userLatitude}
                  userLongitude={userLongitude}
                />
              </div>
            )}
          </div>

          {/* Right: Quick reserve widget */}
          <aside className="home-top-row__sidebar">
            <QuickReserveWidget bakeries={bakeries} />
          </aside>
        </div>
      </div>

      <Footer />
    </div>
  );
}


// --- Quick Reserve Widget ---

interface QuickReserveWidgetProps {
  bakeries: BakeryCard[];
}

// Demo products for the sweet category (in a real app, fetched from the bakery's menu)
const SWEET_PRODUCTS = [
  { id: 'qr-1', name: 'Croissant', price: 150 },
  { id: 'qr-2', name: 'Pain au Chocolat', price: 180 },
  { id: 'qr-3', name: 'Éclair au Café', price: 380 },
  { id: 'qr-4', name: 'Tarte aux Pommes', price: 450 },
  { id: 'qr-5', name: 'Kouign-Amann', price: 320 },
];

function getPickupTime(): string {
  const d = new Date(Date.now() + 30 * 60 * 1000); // 30 min from now
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  return `${days[d.getDay()]} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}

function QuickReserveWidget({ bakeries }: QuickReserveWidgetProps) {
  const { t } = useI18n();
  const [quantities, setQuantities] = useState<Record<string, number>>({});
  const [submitted, setSubmitted] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  // Pick a random bakery
  const selectedBakery = bakeries.length > 0
    ? bakeries[Math.floor(Math.random() * bakeries.length)]
    : null;

  const totalItems = Object.values(quantities).reduce((s, q) => s + q, 0);
  const totalPrice = SWEET_PRODUCTS.reduce((s, p) => s + (quantities[p.id] || 0) * p.price, 0);
  const pickupTime = getPickupTime();

  const increment = (id: string) => {
    setQuantities((prev) => ({ ...prev, [id]: (prev[id] || 0) + 1 }));
  };

  const decrement = (id: string) => {
    setQuantities((prev) => {
      const cur = prev[id] || 0;
      if (cur <= 0) return prev;
      const next = { ...prev, [id]: cur - 1 };
      if (next[id] === 0) delete next[id];
      return next;
    });
  };

  const [needsLogin, setNeedsLogin] = useState(false);

  const handleSubmit = () => {
    if (totalItems === 0) return;
    if (!isAuthenticated()) {
      setNeedsLogin(true);
      return;
    }
    setSubmitted(true);
  };

  const handleConfirm = () => {
    setConfirmed(true);
  };

  if (confirmed) {
    return (
      <div className="quick-reserve">
        <h2 className="quick-reserve__heading">{t('home.quickReserve')}</h2>
        <div style={{ textAlign: 'center', padding: '2rem 0' }}>
          <span style={{ fontSize: '2rem' }}>✓</span>
          <p style={{ margin: '0.5rem 0 0', color: 'var(--ink)' }}>Reservation confirmed!</p>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.85rem', color: 'var(--ink-muted)' }}>
            Pickup at {pickupTime}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="quick-reserve">
      <h2 className="quick-reserve__heading">{t('home.quickReserve')}</h2>

      {/* Bakery slot — shows random bakery photo + name */}
      <div className="quick-reserve__bakery-slot">
        {selectedBakery ? (
          <div className="quick-reserve__bakery-info">
            <img
              src={selectedBakery.photoUrl}
              alt={selectedBakery.name}
              className="quick-reserve__bakery-img"
            />
            <span className="quick-reserve__bakery-name">{selectedBakery.name}</span>
          </div>
        ) : (
          <span className="quick-reserve__bakery-placeholder">bakery</span>
        )}
      </div>

      {/* Animated content area */}
      <div className={`quick-reserve__content-area ${submitted ? 'quick-reserve__content-area--summary' : ''}`}>
        {/* Selection view */}
        <div className={`quick-reserve__view ${!submitted ? 'quick-reserve__view--visible' : 'quick-reserve__view--hidden'}`}>
          <div className="quick-reserve__items">
            {SWEET_PRODUCTS.map((product) => {
              const qty = quantities[product.id] || 0;
              return (
                <div key={product.id} className="quick-reserve__product-row">
                  <span className="quick-reserve__product-name">{product.name}</span>
                  <div className="quick-reserve__stepper">
                    <button
                      type="button"
                      className="quick-reserve__stepper-btn"
                      onClick={() => decrement(product.id)}
                      disabled={qty === 0}
                    >
                      −
                    </button>
                    <span className="quick-reserve__stepper-qty">{qty}</span>
                    <button
                      type="button"
                      className="quick-reserve__stepper-btn"
                      onClick={() => increment(product.id)}
                    >
                      +
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
          <button
            type="button"
            className="quick-reserve__submit-btn"
            disabled={totalItems === 0}
            onClick={handleSubmit}
          >
            {totalItems === 0
              ? 'Select items'
              : `Add ${totalItems} item${totalItems > 1 ? 's' : ''} · €${(totalPrice / 100).toFixed(2)}`}
          </button>
          {needsLogin && (
            <div className="quick-reserve__login-hint">
              <p>Sign in to complete your reservation</p>
              <Link to="/login" className="quick-reserve__login-link">Sign in →</Link>
            </div>
          )}
        </div>

        {/* Summary view */}
        <div className={`quick-reserve__view ${submitted ? 'quick-reserve__view--visible' : 'quick-reserve__view--hidden'}`}>
          <div className="quick-reserve__summary">
            <h3 className="quick-reserve__summary-title">Order Summary</h3>
            {SWEET_PRODUCTS.filter((p) => (quantities[p.id] || 0) > 0).map((product) => (
              <div key={product.id} className="quick-reserve__summary-row">
                <span>{quantities[product.id]}× {product.name}</span>
                <span>€{((quantities[product.id] || 0) * product.price / 100).toFixed(2)}</span>
              </div>
            ))}
            <div className="quick-reserve__summary-total">
              <span>Total</span>
              <span>€{(totalPrice / 100).toFixed(2)}</span>
            </div>
          </div>
          <button
            type="button"
            className="quick-reserve__back-btn"
            onClick={() => setSubmitted(false)}
          >
            ← Modify selection
          </button>
          <div className="quick-reserve__footer" onClick={handleConfirm} style={{ cursor: 'pointer' }}>
            <span className="quick-reserve__footer-line">
              {pickupTime} · €{(totalPrice / 100).toFixed(2)}
            </span>
            <span className="quick-reserve__footer-sub">{t('home.payAtPickup')}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
