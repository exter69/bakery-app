import { useState, useEffect, useMemo } from 'react';
import { fetchMyBakery, fetchProducts } from '../../api/seller';
import { useI18n } from '../../i18n';
import { StockStepper } from '../../components/pro/StockStepper';
import { calculateBundlePrice, capQuantity } from './bundle-utils';
import type { BundleItem } from './bundle-utils';
import type { Bakery, Product } from '../../types/bakery';
import './DashboardBundles.css';

/** Generate time options in 30-minute increments from 16:00 to 21:00 */
function generateTimeOptions(): string[] {
  const options: string[] = [];
  for (let h = 16; h <= 21; h++) {
    options.push(`${String(h).padStart(2, '0')}:00`);
    if (h < 21) {
      options.push(`${String(h).padStart(2, '0')}:30`);
    }
  }
  return options;
}

const TIME_OPTIONS = generateTimeOptions();

/** Format cents to euro display string */
function formatEur(cents: number): string {
  return `€${(cents / 100).toFixed(2)}`;
}

/** Derive closing time string from bakery schedule */
function getClosingTime(bakery: Bakery | null): string {
  if (!bakery?.schedule?.length) return '19:00';
  const days = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];
  const today = days[new Date().getDay()];
  const todaySchedule = bakery.schedule.find((s) => s.day.toLowerCase() === today);
  if (!todaySchedule?.isOpen) return '19:00';
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${pad(todaySchedule.closeTime.hour)}:${pad(todaySchedule.closeTime.minute)}`;
}

export default function DashboardBundles() {
  const { t } = useI18n();
  const [bakery, setBakery] = useState<Bakery | null>(null);
  const [items, setItems] = useState<BundleItem[]>([]);
  const [basketCount, setBasketCount] = useState(1);
  const [pickupStart, setPickupStart] = useState('18:30');
  const [pickupEnd, setPickupEnd] = useState('19:00');
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadData() {
      try {
        const b = await fetchMyBakery();
        if (!b || cancelled) {
          setLoading(false);
          return;
        }
        setBakery(b);

        const products = await fetchProducts(b.id);
        if (cancelled) return;

        // Map products to BundleItems — use a deterministic "remaining" derived from product data
        const bundleItems: BundleItem[] = products
          .filter((p: Product) => p.isAvailable)
          .map((p: Product) => ({
            productId: p.id,
            name: p.name,
            remaining: Math.max(1, (p.price % 10) + 1),
            selected: false,
            quantity: 1,
            price: p.price,
          }));

        setItems(bundleItems);
      } catch {
        setMsg({ type: 'error', text: t('dashboard.bundles.loadError') });
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    loadData();
    return () => { cancelled = true; };
  }, [t]);

  const closingTime = getClosingTime(bakery);

  // Toggle product selection
  function handleToggle(productId: string) {
    setItems((prev) =>
      prev.map((item) =>
        item.productId === productId
          ? { ...item, selected: !item.selected }
          : item
      )
    );
  }

  // Change quantity for a product
  function handleQuantityChange(productId: string, newQuantity: number) {
    setItems((prev) =>
      prev.map((item) =>
        item.productId === productId
          ? { ...item, quantity: capQuantity(newQuantity, item.remaining) }
          : item
      )
    );
  }

  // Compute pricing
  const pricing = useMemo(() => calculateBundlePrice(items), [items]);
  const hasSelection = items.some((item) => item.selected);
  const selectedItems = items.filter((item) => item.selected);

  // Discount percentage display
  const discountPct = pricing.originalPrice > 0
    ? Math.round((1 - pricing.discountedPrice / pricing.originalPrice) * 100)
    : 0;

  // Publish handler — stub with toast for now
  async function handlePublish() {
    if (!hasSelection) return;

    const bundleData = {
      bakeryId: bakery?.id,
      name: `Panier ${bakery?.name ?? 'du soir'}`,
      items: selectedItems.map((item) => ({
        productId: item.productId,
        quantity: item.quantity,
      })),
      originalPrice: pricing.originalPrice,
      discountedPrice: pricing.discountedPrice,
      basketCount,
      pickupStart,
      pickupEnd,
    };

    console.log('[DashboardBundles] Publishing bundle:', bundleData);
    setMsg({ type: 'success', text: t('dashboard.bundles.publishSuccess').replace('{count}', String(basketCount)) });
  }

  if (loading) {
    return <div className="dash-loading">{t('dashboard.bundles.loading')}</div>;
  }

  if (!bakery) {
    return (
      <div className="dash-empty">
        <p>{t('dashboard.bundles.noBakery')}</p>
      </div>
    );
  }

  return (
    <div className="bundle-composer">
      {/* Header */}
      <header className="bundle-composer__header">
        <div className="bundle-composer__title-row">
          <h1 className="bundle-composer__title">{t('dashboard.bundles.title')}</h1>
          <span className="bundle-composer__badge">{t('dashboard.bundles.badge')}</span>
        </div>
        <p className="bundle-composer__subtitle">
          {t('dashboard.bundles.closingTime').replace('{time}', closingTime)} · {t('dashboard.bundles.publishHint')}
        </p>
      </header>

      {/* Toast message */}
      {msg && (
        <div className={`bundle-composer__toast bundle-composer__toast--${msg.type}`}>
          {msg.text}
        </div>
      )}

      {/* Split panels */}
      <div className="bundle-composer__panels">
        {/* Left panel: product checklist */}
        <div className="bundle-composer__left">
          <div className="bundle-composer__section-label">{t('dashboard.bundles.step1')}</div>
          <div className="bundle-composer__section-hint">
            {t('dashboard.bundles.step1Hint')}
          </div>

          {items.length === 0 ? (
            <div className="bundle-composer__empty">{t('dashboard.bundles.emptyProducts')}</div>
          ) : (
            <div className="bundle-composer__product-list">
              {items.map((item) => (
                <div key={item.productId} className="bundle-composer__product-row">
                  <input
                    type="checkbox"
                    className="bundle-composer__checkbox"
                    checked={item.selected}
                    onChange={() => handleToggle(item.productId)}
                    aria-label={`${item.name}`}
                  />
                  <div className="bundle-composer__product-info">
                    <span className="bundle-composer__product-name">{item.name}</span>
                    <span className="bundle-composer__product-remaining">
                      {t('dashboard.bundles.remaining').replace('{n}', String(item.remaining))}
                    </span>
                  </div>
                  {item.selected && (
                    <div className="bundle-composer__product-stepper">
                      <StockStepper
                        value={item.quantity}
                        min={1}
                        max={item.remaining}
                        onChange={(val) => handleQuantityChange(item.productId, val)}
                      />
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Right panel: client preview */}
        <div className="bundle-composer__right">
          <div className="bundle-composer__preview-label">{t('dashboard.bundles.preview')}</div>

          <div className="bundle-composer__preview-card">
            <div className="bundle-composer__preview-name">
              {t('dashboard.bundles.bundleName').replace('{name}', bakery.name)}
            </div>
            <div className="bundle-composer__preview-time">
              {t('dashboard.bundles.pickupTime').replace('{start}', pickupStart).replace('{end}', pickupEnd)}
            </div>

            {selectedItems.length > 0 ? (
              <ul className="bundle-composer__preview-items">
                {selectedItems.map((item) => (
                  <li key={item.productId}>{item.quantity}x {item.name}</li>
                ))}
              </ul>
            ) : (
              <p className="bundle-composer__empty">
                {t('dashboard.bundles.selectForPreview')}
              </p>
            )}

            {hasSelection && (
              <div className="bundle-composer__preview-pricing">
                <span className="bundle-composer__preview-original">
                  {formatEur(pricing.originalPrice)}
                </span>
                <span className="bundle-composer__preview-discounted">
                  {formatEur(pricing.discountedPrice)}
                </span>
              </div>
            )}

            <button
              type="button"
              className="bundle-composer__preview-btn"
              disabled
              aria-label={t('dashboard.bundles.reserve')}
            >
              {t('dashboard.bundles.reserve')}
            </button>
          </div>

          {/* Controls below preview */}
          <div className="bundle-composer__controls">
            {/* Price summary */}
            {hasSelection && (
              <div className="bundle-composer__control-row">
                <span className="bundle-composer__control-label">{t('dashboard.bundles.price')}</span>
                <span>{formatEur(pricing.discountedPrice)}</span>
                <span className="bundle-composer__control-suffix">(-{discountPct}%)</span>
              </div>
            )}

            {/* Basket count stepper */}
            <div className="bundle-composer__control-row">
              <span className="bundle-composer__control-label">{t('dashboard.bundles.quantity')}</span>
              <StockStepper
                value={basketCount}
                min={1}
                onChange={setBasketCount}
              />
              <span className="bundle-composer__control-suffix">{t('dashboard.bundles.baskets')}</span>
            </div>

            {/* Pickup time window */}
            <div className="bundle-composer__control-row">
              <span className="bundle-composer__control-label">{t('dashboard.bundles.pickup')}</span>
              <select
                className="bundle-composer__time-select"
                value={pickupStart}
                onChange={(e) => setPickupStart(e.target.value)}
                aria-label={t('dashboard.bundles.pickup')}
              >
                {TIME_OPTIONS.map((opt) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
              <span className="bundle-composer__time-separator">—</span>
              <select
                className="bundle-composer__time-select"
                value={pickupEnd}
                onChange={(e) => setPickupEnd(e.target.value)}
                aria-label={t('dashboard.bundles.pickup')}
              >
                {TIME_OPTIONS.map((opt) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </div>

      {/* Footer: publish button */}
      <footer className="bundle-composer__footer">
        <button
          type="button"
          className="bundle-composer__publish-btn"
          disabled={!hasSelection}
          onClick={handlePublish}
        >
          {t('dashboard.bundles.publish')}
        </button>
      </footer>
    </div>
  );
}
