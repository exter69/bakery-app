import { useCallback, useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { fetchBakery, fetchMenu } from '../api/bakeries';
import { createOrder, createReservation } from '../api/orders';
import { isAuthenticated } from '../api/client';
import type { Bakery, Menu, Product } from '../types/bakery';
import OrderSidePanel from '../components/OrderSidePanel';
import type { OrderItem } from '../components/OrderSidePanel';
import ReservationSidePanel from '../components/ReservationSidePanel';
import type { ReservationItem } from '../components/ReservationSidePanel';
import ProductSelectionOverlay from '../components/ProductSelectionOverlay';
import './BakeryDetailPage.css';

type ActivePanel = 'order' | 'reservation' | null;

export default function BakeryDetailPage() {
  const { id } = useParams<{ id: string }>();

  const [bakery, setBakery] = useState<Bakery | null>(null);
  const [menu, setMenu] = useState<Menu | null>(null);
  const [loading, setLoading] = useState(true);
  const [menuError, setMenuError] = useState<string | null>(null);
  const [orderPanelOpen, setOrderPanelOpen] = useState(false);
  const [reservationPanelOpen, setReservationPanelOpen] = useState(false);
  const [orderItems, setOrderItems] = useState<OrderItem[]>([]);
  const [reservationItems, setReservationItems] = useState<ReservationItem[]>([]);
  const [selectionMode, setSelectionMode] = useState<ActivePanel>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

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

  // Handle product click during selection mode
  const handleProductClick = (product: Product) => {
    if (selectionMode === 'order') {
      setOrderItems((prev) => {
        const existing = prev.find((item) => item.product.id === product.id);
        if (existing) {
          return prev.map((item) =>
            item.product.id === product.id
              ? { ...item, quantity: item.quantity + 1 }
              : item
          );
        }
        return [...prev, { product, quantity: 1 }];
      });
    } else if (selectionMode === 'reservation') {
      setReservationItems((prev) => {
        const existing = prev.find((item) => item.product.id === product.id);
        if (existing) {
          return prev.map((item) =>
            item.product.id === product.id
              ? { ...item, quantity: item.quantity + 1 }
              : item
          );
        }
        return [...prev, { product, quantity: 1 }];
      });
    }
  };

  // Handle order submission → calls API, handles payment redirect
  const handleOrderSubmit = async (day: string, startTime: string, endTime: string) => {
    if (!id || orderItems.length === 0) return;

    setSubmitError(null);
    setSubmitting(true);

    try {
      const response = await createOrder({
        bakeryId: id,
        items: orderItems.map((item) => ({
          productId: item.product.id,
          quantity: item.quantity,
        })),
        scheduledDay: day,
        scheduledTime: { startTime, endTime },
      });

      // Redirect to payment URL
      if (response.paymentUrl) {
        window.location.href = response.paymentUrl;
      }
    } catch (err) {
      // Retain user selections, show error in the panel
      setSubmitError(err instanceof Error ? err.message : 'Order submission failed. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  // Handle reservation submission → calls API, no payment redirect
  const handleReservationSubmit = async (day: string, startTime: string, endTime: string) => {
    if (!id || reservationItems.length === 0) return;

    setSubmitError(null);
    setSubmitting(true);

    try {
      await createReservation({
        bakeryId: id,
        items: reservationItems.map((item) => ({
          productId: item.product.id,
          quantity: item.quantity,
        })),
        scheduledDay: day,
        scheduledTime: { startTime, endTime },
      });

      // Reservation confirmed - close panel and reset
      setReservationPanelOpen(false);
      setReservationItems([]);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Reservation submission failed. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  // Close order panel
  const handleCloseOrderPanel = () => {
    setOrderPanelOpen(false);
    setOrderItems([]);
    setSelectionMode(null);
    setSubmitError(null);
  };

  // Close reservation panel
  const handleCloseReservationPanel = () => {
    setReservationPanelOpen(false);
    setReservationItems([]);
    setSelectionMode(null);
    setSubmitError(null);
  };

  // Get flat list of products for overlay
  const allProducts = menu ? Object.values(menu).flat() : [];

  if (loading) {
    return (
      <div className="bakery-detail__loading">
        <div className="spinner" aria-label="Loading bakery details" />
        <p>Loading bakery details...</p>
      </div>
    );
  }

  if (!bakery) {
    return (
      <div className="bakery-detail__error">
        <p>Bakery not found.</p>
      </div>
    );
  }

  const menuCategories = menu ? Object.entries(menu) : [];
  const hasMenuItems = menuCategories.some(([, products]) => products.length > 0);

  return (
    <div className="bakery-detail">
      {/* Bakery Header */}
      <header className="bakery-detail__header">
        <img
          src={bakery.photoUrl}
          alt={bakery.name}
          className="bakery-detail__photo"
        />
        <div className="bakery-detail__info">
          <h1 className="bakery-detail__name">{bakery.name}</h1>
          <p className="bakery-detail__description">{bakery.description}</p>
          <p className="bakery-detail__address">
            <span className="bakery-detail__address-icon" aria-hidden="true">📍</span>
            {bakery.address}
          </p>
        </div>
      </header>

      {/* Action Buttons */}
      <div className="bakery-detail__actions">
        {isAuthenticated() ? (
          <>
            <button
              type="button"
              className="btn btn--primary"
              onClick={() => {
                setOrderPanelOpen(true);
                setSubmitError(null);
              }}
            >
              Place Order
            </button>
            <button
              type="button"
              className="btn btn--secondary"
              onClick={() => {
                setReservationPanelOpen(true);
                setSubmitError(null);
              }}
            >
              Make Reservation
            </button>
          </>
        ) : (
          <div className="bakery-detail__guest-banner">
            <p>
              <Link to="/login" className="bakery-detail__sign-in-link">Sign in</Link> to place orders and make reservations.
            </p>
          </div>
        )}
      </div>

      {/* Submission Error (visible at page level) */}
      {submitError && (
        <div className="bakery-detail__submit-error" role="alert">
          <p>{submitError}</p>
        </div>
      )}

      {/* Menu Section */}
      <section className="bakery-detail__menu">
        <h2 className="bakery-detail__menu-title">Menu</h2>

        {menuError && (
          <div className="bakery-detail__menu-error">
            <p>{menuError}</p>
            <button
              type="button"
              className="btn btn--outline"
              onClick={loadMenu}
            >
              Retry
            </button>
          </div>
        )}

        {!menuError && !hasMenuItems && (
          <div className="bakery-detail__menu-empty">
            <p>Menu is currently empty.</p>
          </div>
        )}

        {!menuError && hasMenuItems && (
          <div className="bakery-detail__categories">
            {menuCategories.map(([category, products]) => (
              <div key={category} className="menu-category">
                <h3 className="menu-category__title">{category}</h3>
                <div className="menu-category__grid">
                  {products.map((product) => (
                    <article
                      key={product.id}
                      className={`product-card ${selectionMode ? 'product-card--selectable' : ''}`}
                      onClick={selectionMode ? () => handleProductClick(product) : undefined}
                      role={selectionMode ? 'button' : undefined}
                      tabIndex={selectionMode ? 0 : undefined}
                      aria-label={selectionMode ? `Add ${product.name} to ${selectionMode}` : undefined}
                    >
                      {product.photoUrl && (
                        <img
                          src={product.photoUrl}
                          alt={product.name}
                          className="product-card__photo"
                          loading="lazy"
                        />
                      )}
                      <div className="product-card__content">
                        <h4 className="product-card__name">{product.name}</h4>
                        {product.description && (
                          <p className="product-card__description">
                            {product.description}
                          </p>
                        )}
                        <p className="product-card__price">
                          €{(product.price / 100).toFixed(2)}
                        </p>
                      </div>
                    </article>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Product Selection Overlay */}
      <ProductSelectionOverlay
        isActive={selectionMode !== null}
        products={allProducts}
        selectedItems={selectionMode === 'order' ? orderItems : reservationItems}
        onProductClick={handleProductClick}
        onDone={() => setSelectionMode(null)}
      />

      {/* Order Side Panel */}
      <OrderSidePanel
        isOpen={orderPanelOpen}
        onClose={handleCloseOrderPanel}
        bakerySchedule={bakery.schedule}
        onStartSelection={() => setSelectionMode('order')}
        items={orderItems}
        onSubmit={handleOrderSubmit}
        submitting={submitting}
        submitError={submitError}
      />

      {/* Reservation Side Panel */}
      <ReservationSidePanel
        isOpen={reservationPanelOpen}
        onClose={handleCloseReservationPanel}
        schedule={bakery.schedule}
        items={reservationItems}
        onItemsChange={setReservationItems}
        onStartSelection={() => setSelectionMode('reservation')}
        onSubmit={handleReservationSubmit}
        submitting={submitting}
        submitError={submitError}
      />
    </div>
  );
}
