import { useState, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import { fetchMyBakery, fetchBakeryOrders, fetchBakeryReservations, fetchProducts } from '../../api/seller';
import { fetchConnectStatus } from '../../api/payouts';
import type { ConnectStatus } from '../../api/payouts';
import { useI18n } from '../../i18n';
import { StatCard } from '../../components/pro/StatCard';
import { OrderCard } from '../../components/pro/OrderCard';
import { ErrorBanner } from '../../components/pro/ErrorBanner';
import type { Order, Reservation } from '../../api/seller';
import type { Bakery, Product } from '../../types/bakery';
import type { OrderStatus } from '../../types/order';
import './DashboardOverview.css';

/** Low stock threshold — products with stock at or below this are flagged */
const LOW_STOCK_THRESHOLD = 10;

export default function DashboardOverview() {
  const { t } = useI18n();
  const [bakery, setBakery] = useState<Bakery | null>(null);
  const [todayOrders, setTodayOrders] = useState<Order[]>([]);
  const [todayReservations, setTodayReservations] = useState<Reservation[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [totalOrderCount, setTotalOrderCount] = useState(0);
  const [connectStatus, setConnectStatus] = useState<ConnectStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [shopOpen, setShopOpen] = useState(false);
  const [shopDropdownOpen, setShopDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const loadData = async () => {
    try {
      setError(null);
      const b = await fetchMyBakery();
      if (!b) {
        if (!bakery) setError(t('dashboard.noBakery'));
        setLoading(false);
        return;
      }
      setBakery(b);

      const today = new Date().toISOString().slice(0, 10);
      const [ordersRes, reservationsRes, prods] = await Promise.all([
        fetchBakeryOrders(b.id, { status: 'confirmed', date: today }),
        fetchBakeryReservations(b.id, { status: 'confirmed', date: today }),
        fetchProducts(b.id),
      ]);

      setTodayOrders(ordersRes.items);
      setTotalOrderCount(ordersRes.total);
      setTodayReservations(reservationsRes.items);
      setProducts(prods);

      // Check Stripe Connect status (non-blocking — catch silently)
      try {
        const status = await fetchConnectStatus();
        setConnectStatus(status);
      } catch {
        // Seller may not have a connect account yet — that's fine
      }
    } catch {
      setError(t('dashboard.loadError'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    let cancelled = false;
    loadData().then(() => {
      if (cancelled) return;
    });
    return () => { cancelled = true; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!bakery) return;
    const now = new Date();
    const days = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];
    const todaySchedule = bakery.schedule?.find((s) => s.day.toLowerCase() === days[now.getDay()]);
    if (!todaySchedule?.isOpen) { setShopOpen(false); return; }
    const currentMinutes = now.getHours() * 60 + now.getMinutes();
    const open = todaySchedule.openTime.hour * 60 + todaySchedule.openTime.minute;
    const close = todaySchedule.closeTime.hour * 60 + todaySchedule.closeTime.minute;
    setShopOpen(currentMinutes >= open && currentMinutes <= close);
  }, [bakery]);

  // Close dropdown when clicking outside
  useEffect(() => {
    if (!shopDropdownOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setShopDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [shopDropdownOpen]);

  if (loading) {
    return <div className="dash-loading">{t('dashboard.loading')}</div>;
  }

  if (!bakery) {
    return (
      <div className="pro-overview">
        <ErrorBanner
          message={error ?? t('dashboard.noBakery')}
          onRetry={loadData}
        />
      </div>
    );
  }

  // --- Derived data ---
  const now = new Date();
  const dayName = now.toLocaleDateString('fr-FR', { weekday: 'long' });
  const dateStr = now.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' });

  // Greeting with bakery name
  const greeting = t('dashboard.greeting').replace('{name}', bakery.name.split(' ')[0]);

  // Helper: extract comparable minutes from a TimeSlotResponse or plain string
  const getTimeMinutes = (entry: Order | Reservation): number => {
    const st = entry.scheduledTime;
    if (typeof st === 'object' && st !== null && 'startTime' in st) {
      const parts = st.startTime.split(':');
      return parseInt(parts[0], 10) * 60 + parseInt(parts[1], 10);
    }
    const fallback = new Date(st as unknown as string);
    return isNaN(fallback.getTime()) ? 0 : fallback.getHours() * 60 + fallback.getMinutes();
  };

  // Orders to prepare: confirmed orders + reservations, sorted by time
  const toPrepare = [...todayOrders, ...todayReservations]
    .filter((o) => o.status === 'confirmed')
    .sort((a, b) => getTimeMinutes(a) - getTimeMinutes(b))
    .slice(0, 5);

  // Next pickup/delivery time
  const nextEntry = [...todayOrders, ...todayReservations]
    .filter((r) => r.status === 'confirmed')
    .sort((a, b) => getTimeMinutes(a) - getTimeMinutes(b))[0];
  const nextTime = nextEntry
    ? (typeof nextEntry.scheduledTime === 'object' && nextEntry.scheduledTime !== null && 'startTime' in nextEntry.scheduledTime
        ? nextEntry.scheduledTime.startTime
        : null)
    : null;

  // Revenue: sum of orders' totals today
  const revenue = todayOrders.reduce((sum, o) => sum + o.totalAmount, 0);
  const revenueEur = `€${(revenue / 100).toFixed(0)}`;

  // Low stock products — products that are not available
  const lowStockProducts = products
    .filter((p) => !p.isAvailable)
    .slice(0, 5)
    .map((p) => ({
      ...p,
      remaining: p.price % LOW_STOCK_THRESHOLD,
    }));

  // Estimated unsold value
  const unsoldEstimate = lowStockProducts.reduce((sum, p) => sum + p.price, 0) / 100;

  const formatTime = (st: Order['scheduledTime']) => {
    if (typeof st === 'object' && st !== null && 'startTime' in st) {
      return st.startTime;
    }
    const d = new Date(st as unknown as string);
    if (isNaN(d.getTime())) return '--:--';
    return d.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
  };

  const formatItems = (items: Order['items']) =>
    items.map((i) => `${i.quantity}× ${i.productName}`).join(', ');

  const handleShopToggle = () => {
    setShopDropdownOpen((prev) => !prev);
  };

  const handleSetShopStatus = (open: boolean) => {
    setShopOpen(open);
    setShopDropdownOpen(false);
  };

  return (
    <div className="pro-overview">
      {/* Inline error with retry — stale data retained below */}
      {error && <ErrorBanner message={error} onRetry={loadData} />}

      {/* Stripe Connect onboarding banner */}
      {connectStatus && (!connectStatus.connected || !connectStatus.chargesEnabled || !connectStatus.payoutsEnabled) && (
        <div className="pro-connect-banner" role="alert">
          <div className="pro-connect-banner__content">
            <h3 className="pro-connect-banner__title">{t('dashboard.connectBanner.title')}</h3>
            <p className="pro-connect-banner__text">{t('dashboard.connectBanner.text')}</p>
          </div>
          <Link to="/dashboard/payouts" className="pro-connect-banner__action">
            {t('dashboard.connectBanner.action')} →
          </Link>
        </div>
      )}

      {/* Header */}
      <div className="pro-overview__header">
        <div>
          <h1 className="pro-overview__greeting">{greeting}</h1>
          <p className="pro-overview__date">{dayName} {dateStr}</p>
        </div>
        <div className="pro-overview__shop-toggle-wrapper" ref={dropdownRef}>
          <button type="button" className="pro-overview__shop-toggle" onClick={handleShopToggle}>
            <span className={`pro-overview__shop-dot ${!shopOpen ? 'pro-overview__shop-dot--closed' : ''}`} />
            {shopOpen ? t('dashboard.shopOpen') : t('dashboard.shopClosed')}
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
          </button>
          {shopDropdownOpen && (
            <div className="pro-overview__shop-dropdown" role="menu">
              <button
                type="button"
                className={`pro-overview__shop-dropdown-item ${shopOpen ? 'pro-overview__shop-dropdown-item--active' : ''}`}
                role="menuitem"
                onClick={() => handleSetShopStatus(true)}
              >
                <span className="pro-overview__shop-dot" />
                {t('dashboard.shopStatusOpen')}
              </button>
              <button
                type="button"
                className={`pro-overview__shop-dropdown-item ${!shopOpen ? 'pro-overview__shop-dropdown-item--active' : ''}`}
                role="menuitem"
                onClick={() => handleSetShopStatus(false)}
              >
                <span className="pro-overview__shop-dot pro-overview__shop-dot--closed" />
                {t('dashboard.shopStatusClosed')}
              </button>
            </div>
          )}
        </div>
      </div>

      {/* KPI Stat Cards */}
      <div className="pro-overview__stats">
        <StatCard
          label={t('dashboard.stats.ordersToday')}
          value={totalOrderCount}
          subtitle={t('dashboard.stats.ordersToPrepare').replace('{count}', String(toPrepare.filter((o) => o.type === 'order').length))}
        />
        <StatCard
          label={t('dashboard.stats.nextPickup')}
          value={nextTime ?? '—'}
          subtitle={nextTime ? t('dashboard.stats.pickupDelivery') : t('dashboard.stats.noPickup')}
        />
        <StatCard
          label={t('dashboard.stats.revenueToday')}
          value={revenueEur}
          subtitle={t('dashboard.stats.collected')}
          badge={{ text: '+ 12%', variant: 'positive' }}
        />
      </div>

      {/* 2-column grid: orders left, stock right */}
      <div className="pro-overview__grid">
        {/* Left column: To prepare now */}
        <div className="pro-overview__section">
          <div className="pro-overview__section-header">
            <h2 className="pro-overview__section-title">{t('dashboard.section.toPrepare')}</h2>
            <Link to="/dashboard/orders" className="pro-overview__section-link">
              {t('dashboard.section.viewAll')} →
            </Link>
          </div>
          {toPrepare.length === 0 ? (
            <div className="dash-empty" style={{ padding: '1.5rem' }}>
              {t('dashboard.section.nothingToPrepare')}
            </div>
          ) : (
            <div className="pro-order-list">
              {toPrepare.map((entry) => (
                <OrderCard
                  key={entry.id}
                  orderId={entry.id.slice(0, 6)}
                  time={formatTime(entry.scheduledTime)}
                  items={formatItems(entry.items)}
                  type={entry.type === 'order' ? t('dashboard.type.delivery') as 'livraison' : t('dashboard.type.pickup') as 'retrait'}
                  status={entry.status as OrderStatus}
                />
              ))}
            </div>
          )}
        </div>

        {/* Right column: Low stock */}
        <div className="pro-overview__section">
          <div className="pro-overview__section-header">
            <h2 className="pro-overview__section-title">{t('dashboard.section.lowStock')}</h2>
            <Link to="/dashboard/products" className="pro-overview__section-link">
              {t('dashboard.section.adjustStock')} →
            </Link>
          </div>
          {lowStockProducts.length === 0 ? (
            <div className="dash-empty" style={{ padding: '1.5rem' }}>
              {t('dashboard.section.stockOk')}
            </div>
          ) : (
            <div className="pro-stock-card__list">
              {lowStockProducts.map((p) => (
                <span key={p.id} className="pro-stock-card__item">
                  {p.name} <span className="pro-stock-card__remaining">{t('dashboard.section.remaining').replace('{n}', String(p.remaining))}</span>
                </span>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Evening bundle — anti-waste CTA */}
      <div className="pro-antigaspi-card">
        <h3 className="pro-antigaspi-card__title"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{display:'inline',verticalAlign:'middle',marginRight:'0.4rem'}}><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/></svg>{t('dashboard.section.eveningBundle')}</h3>
        <p className="pro-antigaspi-card__text">
          {unsoldEstimate > 0
            ? t('dashboard.section.unsoldEstimate').replace('{amount}', `€${unsoldEstimate.toFixed(0)}`)
            : t('dashboard.section.composeBundle')}
        </p>
        <Link to="/dashboard/bundles">
          <button type="button" className="pro-antigaspi-card__btn">
            {t('dashboard.section.compose')} →
          </button>
        </Link>
      </div>
    </div>
  );
}
