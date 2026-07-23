import { useCallback, useEffect, useMemo, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { fetchBakery, fetchMenu } from '../api/bakeries';
import { createOrder, createReservation, consumeReorderData } from '../api/orders';
import { isAuthenticated } from '../api/client';
import ProductSelectionModal from '../components/ProductSelectionModal';
import HealthScoreDisplay from '../components/HealthScoreDisplay';
import { AllergenIndicator } from '../components/AllergenIndicator';
import AllergenDetailModal from '../components/AllergenDetailModal';
import AllergenInfoIcon from '../components/AllergenInfoIcon';
import type { OrderItem, DeliveryFrequency } from '../components/ProductSelectionModal';
import type { Bakery, DaySchedule, Menu, Product } from '../types/bakery';
import './BakeryDetailPage.css';

function formatTime(t: { hour: number; minute: number }): string {
  return `${String(t.hour).padStart(2, '0')}:${String(t.minute).padStart(2, '0')}`;
}

export default function BakeryDetailPage() {
  const { id } = useParams<{ id: string }>();

  // Data fetching state
  const [bakery, setBakery] = useState<Bakery | null>(null);
  const [menu, setMenu] = useState<Menu | null>(null);
  const [loading, setLoading] = useState(true);
  const [menuError, setMenuError] = useState<string | null>(null);

  // Modal state
  const [modalOpen, setModalOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [initialOrderItems, setInitialOrderItems] = useState<OrderItem[]>([]);
  const [unavailableItems, setUnavailableItems] = useState<{ productName: string }[]>([]);

  // Allergen detail modal state
  const [allergenModalProduct, setAllergenModalProduct] = useState<Product | null>(null);

  // Mobile active category filter
  const [activeCategory, setActiveCategory] = useState<string | null>(null);

  // User location for travel time
  const [userPos, setUserPos] = useState<{ lat: number; lng: number } | null>(null);

  // ─── Data Fetching ───────────────────────────────────────────────────────────

  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => setUserPos({ lat: pos.coords.latitude, lng: pos.coords.longitude }),
        () => {} // silently ignore
      );
    }
  }, []);

  const loadMenu = useCallback(async () => {
    if (!id) return;
    setMenuError(null);
    try {
      const data = await fetchMenu(id);
      setMenu(data);
    } catch (err) {
      setMenuError(err instanceof Error ? err.message : 'Failed to load menu');
    }
  }, [id]);

  useEffect(() => {
    if (!id) return;
    async function load() {
      setLoading(true);
      try {
        const [bakeryData] = await Promise.all([fetchBakery(id!), loadMenu()]);
        setBakery(bakeryData);
      } catch (err) {
        setMenuError(err instanceof Error ? err.message : 'Failed to load bakery');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [id, loadMenu]);

  // Set initial active category for mobile
  useEffect(() => {
    if (menu && !activeCategory) {
      const categories = Object.keys(menu);
      if (categories.length > 0) setActiveCategory(categories[0]);
    }
  }, [menu, activeCategory]);

  // ─── Derived Data ────────────────────────────────────────────────────────────

  const scheduleMap = useMemo(() => {
    const map = new Map<string, DaySchedule>();
    if (bakery) {
      for (const ds of bakery.schedule) {
        map.set(ds.day, ds);
      }
    }
    return map;
  }, [bakery]);

  const todayHours = useMemo(() => {
    if (!bakery) return null;
    const days = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];
    const today = days[new Date().getDay()];
    const s = scheduleMap.get(today);
    if (!s || !s.isOpen) return 'Closed today';
    return `Open ${formatTime(s.openTime)}–${formatTime(s.closeTime)}`;
  }, [bakery, scheduleMap]);

  const travelInfo = useMemo(() => {
    if (!userPos || !bakery) return null;
    const lat1 = userPos.lat, lng1 = userPos.lng;
    const lat2 = bakery.latitude ?? 0;
    const lng2 = bakery.longitude ?? 0;
    if (!lat2 && !lng2) return null;

    const R = 6371;
    const dLat = (lat2 - lat1) * Math.PI / 180;
    const dLng = (lng2 - lng1) * Math.PI / 180;
    const a = Math.sin(dLat / 2) ** 2 +
      Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
      Math.sin(dLng / 2) ** 2;
    const dist = R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

    const walkMin = Math.round(dist / 5 * 60);
    const bikeMin = Math.round(dist / 15 * 60);
    const driveMin = Math.round(dist / 30 * 60);

    return { distanceKm: Math.round(dist * 10) / 10, walkMin, bikeMin, driveMin };
  }, [userPos, bakery]);

  // Menu data for the modal
  const menuCategories = menu ? Object.entries(menu) : [];
  const hasMenuItems = menuCategories.some(([, products]) => products.length > 0);

  const allProducts = useMemo<Product[]>(() => {
    if (!menu) return [];
    return Object.values(menu).flat();
  }, [menu]);

  // Consume re-order data from sessionStorage once menu is loaded
  useEffect(() => {
    if (!menu || !id) return;
    const reorderData = consumeReorderData();
    if (!reorderData || reorderData.bakeryId !== id) return;

    const products = Object.values(menu).flat();
    const matched: OrderItem[] = [];
    const unavailable: { productName: string }[] = [];

    for (const reorderItem of reorderData.items) {
      const product = products.find(
        (p) => p.id === reorderItem.productId && p.isAvailable
      );
      if (product) {
        matched.push({ product, quantity: reorderItem.quantity });
      } else {
        unavailable.push({ productName: reorderItem.productName });
      }
    }

    if (matched.length > 0 || unavailable.length > 0) {
      setInitialOrderItems(matched);
      setUnavailableItems(unavailable);
      setModalOpen(true);
    }
  // Only run once when menu loads — intentionally omitting id from deps
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [menu]);

  const categories = useMemo(() => {
    if (!menu) return [];
    return Object.keys(menu);
  }, [menu]);

  const productsByCategory = useMemo<Record<string, Product[]>>(() => {
    return menu ?? {};
  }, [menu]);

  // ─── Modal Submit Handler ────────────────────────────────────────────────────

  const handleModalSubmit = async (
    items: OrderItem[],
    days: string[],
    time: string,
    mode: 'delivery' | 'reservation',
    frequency?: DeliveryFrequency,
  ) => {
    if (!id || items.length === 0 || days.length === 0 || !time) return;
    setSubmitError(null);
    try {
      const requestItems = items.map((item) => ({
        productId: item.product.id,
        quantity: item.quantity,
      }));

      if (mode === 'delivery') {
        const response = await createOrder({
          bakeryId: id,
          items: requestItems,
          scheduledDay: days[0],
          scheduledTime: { startTime: time, endTime: time },
          recurring: frequency === 'weekly',
          recurringDays: frequency === 'weekly' ? days : undefined,
        });
        if (response.paymentUrl) {
          window.location.href = response.paymentUrl;
          return;
        }
      } else {
        await createReservation({
          bakeryId: id,
          items: requestItems,
          scheduledDay: days[0],
          scheduledTime: { startTime: time, endTime: time },
        });
      }
      setModalOpen(false);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Order submission failed.');
    }
  };

  // ─── Render ──────────────────────────────────────────────────────────────────

  if (loading) {
    return (
      <div className="bakery-page bakery-page--loading">
        <div className="spinner" aria-label="Loading bakery details" />
        <p>Loading bakery details...</p>
      </div>
    );
  }

  if (!bakery) {
    return (
      <div className="bakery-page bakery-page--error">
        <p>Bakery not found.</p>
      </div>
    );
  }

  return (
    <div className="bakery-page">
      <div className="bakery-page__content">
        {/* Desktop header */}
        <header className="bakery-page__header">
          <div className="bakery-page__photo-wrap">
            {bakery.photoUrl ? (
              <img src={bakery.photoUrl} alt={bakery.name} className="bakery-page__photo" />
            ) : (
              <div className="bakery-page__photo-placeholder" aria-label={`${bakery.name} photo`} />
            )}
          </div>
          <div className="bakery-page__info">
            <h1 className="bakery-page__name">{bakery.name}</h1>
            <p className="bakery-page__address">{bakery.address}</p>
            <p className="bakery-page__hours">{todayHours}</p>
            {travelInfo && (
              <p className="bakery-page__travel">
                📍 {travelInfo.distanceKm} km — 🚶 {travelInfo.walkMin} min · 🚲 {travelInfo.bikeMin} min · 🚗 {travelInfo.driveMin} min
              </p>
            )}
          </div>
        </header>

        {/* Mobile hero */}
        <div className="bakery-page__mobile-hero">
          {bakery.photoUrl ? (
            <img src={bakery.photoUrl} alt={bakery.name} className="bakery-page__mobile-hero-img" />
          ) : (
            <div className="bakery-page__mobile-hero-placeholder" aria-label={`${bakery.name} photo`} />
          )}
          <div className="bakery-page__mobile-info">
            <h1 className="bakery-page__name">{bakery.name}</h1>
            <p className="bakery-page__hours">{todayHours}</p>
            {travelInfo && (
              <p className="bakery-page__travel">
                📍 {travelInfo.distanceKm} km — 🚶 {travelInfo.walkMin} min · 🚲 {travelInfo.bikeMin} min · 🚗 {travelInfo.driveMin} min
              </p>
            )}
          </div>
        </div>

        {/* Mobile category chips */}
        {hasMenuItems && (
          <div className="bakery-page__category-chips">
            {menuCategories.map(([category]) => (
              <button
                key={category}
                type="button"
                className={`bakery-page__category-chip ${activeCategory === category ? 'bakery-page__category-chip--active' : ''}`}
                onClick={() => setActiveCategory(category)}
              >
                {category}
              </button>
            ))}
          </div>
        )}

        {/* Menu section */}
        <section className="bakery-page__menu">
          {menuError && (
            <div className="bakery-page__menu-error">
              <p>{menuError}</p>
              <button type="button" className="bakery-page__retry-btn" onClick={loadMenu}>
                Retry
              </button>
            </div>
          )}

          {!menuError && !hasMenuItems && (
            <div className="bakery-page__menu-empty">
              <p>Menu is currently empty.</p>
            </div>
          )}

          {!menuError && hasMenuItems && (
            <div className="bakery-page__categories">
              {menuCategories.map(([category, products]) => (
                <div
                  key={category}
                  className={`bakery-page__category ${activeCategory !== null && activeCategory !== category ? 'bakery-page__category--hidden-mobile' : ''}`}
                >
                  <h3 className="bakery-page__category-title">{category}</h3>

                  {/* Desktop product grid */}
                  <div className="bakery-page__product-grid">
                    {products.map((product) => (
                      <article key={product.id} className="product-card">
                        {product.photoUrl ? (
                          <img src={product.photoUrl} alt={product.name} className="product-card__photo" loading="lazy" />
                        ) : (
                          <div className="product-card__photo-placeholder" aria-hidden="true" />
                        )}
                        <div className="product-card__body">
                          <span className="product-card__name">{product.name}</span>
                          <span className="product-card__price">€{(product.price / 100).toFixed(2)}</span>
                          {product.healthScore != null && <HealthScoreDisplay score={product.healthScore} />}
                        </div>
                        {product.allergens?.length > 0 && (
                          <AllergenIndicator
                            allergens={product.allergens}
                            productName={product.name}
                            onOpenModal={() => setAllergenModalProduct(product)}
                          />
                        )}
                      </article>
                    ))}
                  </div>

                  {/* Mobile product rows */}
                  <div className="bakery-page__product-rows">
                    {products.map((product) => (
                      <div key={product.id} className="product-row">
                        {product.photoUrl ? (
                          <img src={product.photoUrl} alt={product.name} className="product-row__photo" loading="lazy" />
                        ) : (
                          <div className="product-row__photo-placeholder" aria-hidden="true" />
                        )}
                        <div className="product-row__info">
                          <span className="product-row__name">{product.name}</span>
                          <span className="product-row__price">€{(product.price / 100).toFixed(2)}</span>
                          {product.healthScore != null && <HealthScoreDisplay score={product.healthScore} />}
                        </div>
                        {product.allergens?.length > 0 && (
                          <AllergenIndicator
                            allergens={product.allergens}
                            productName={product.name}
                            onOpenModal={() => setAllergenModalProduct(product)}
                          />
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Sign-in banner when not authenticated */}
        {!isAuthenticated() && (
          <div className="bakery-page__sign-in-banner">
            <Link to="/login" className="bakery-page__sign-in-link">Sign in</Link> to place orders
          </div>
        )}

        {/* Floating order button (desktop) */}
        {isAuthenticated() && (
          <button
            type="button"
            className="bakery-page__order-btn"
            onClick={() => setModalOpen(true)}
          >
            Start a delivery order →
          </button>
        )}

        {/* Submit error toast */}
        {submitError && (
          <div className="bakery-page__sign-in-banner" role="alert">
            {submitError}
          </div>
        )}
      </div>

      {/* Product Selection Modal */}
      <ProductSelectionModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        bakeryName={bakery.name}
        products={allProducts}
        categories={categories}
        productsByCategory={productsByCategory}
        schedule={bakery.schedule}
        minDeliveryAmount={bakery.minDeliveryAmount}
        onSubmit={handleModalSubmit}
        initialItems={initialOrderItems}
        unavailableItems={unavailableItems}
      />

      {/* Allergen Detail Modal */}
      {allergenModalProduct && (
        <AllergenDetailModal
          isOpen={!!allergenModalProduct}
          onClose={() => setAllergenModalProduct(null)}
          productName={allergenModalProduct.name}
          allergens={allergenModalProduct.allergens}
        />
      )}

      {/* Page-level allergen info floating button */}
      <AllergenInfoIcon />
    </div>
  );
}
