import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { DaySchedule, Product } from '../types/bakery';
import { AllergenIndicator } from './AllergenIndicator';
import AllergenDetailModal from './AllergenDetailModal';
import HealthScoreDisplay from './HealthScoreDisplay';
import './ProductSelectionModal.css';

/** An item selected for the order */
export interface OrderItem {
  product: Product;
  quantity: number;
}

export type DeliveryFrequency = 'one-time' | 'weekly';

export interface ProductSelectionModalProps {
  isOpen: boolean;
  onClose: () => void;
  bakeryName: string;
  products: Product[];
  categories: string[];
  productsByCategory: Record<string, Product[]>;
  schedule: DaySchedule[];
  minDeliveryAmount?: number; // in cents
  onSubmit: (items: OrderItem[], days: string[], time: string, mode: 'delivery' | 'reservation', frequency?: DeliveryFrequency) => void;
  /** Pre-filled items for re-order flow */
  initialItems?: OrderItem[];
  /** Items that are no longer available (shown as struck-through) */
  unavailableItems?: { productName: string }[];
}

const DAY_ORDER = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
const DAY_LABELS: Record<string, string> = {
  monday: 'Mon',
  tuesday: 'Tue',
  wednesday: 'Wed',
  thursday: 'Thu',
  friday: 'Fri',
  saturday: 'Sat',
  sunday: 'Sun',
};

function generateTimeSlots(open: { hour: number; minute: number }, close: { hour: number; minute: number }): string[] {
  const slots: string[] = [];
  let h = open.hour;
  let m = open.minute;
  const closeMin = close.hour * 60 + close.minute;
  while (h * 60 + m < closeMin) {
    slots.push(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
    m += 30;
    if (m >= 60) { h += 1; m = 0; }
  }
  return slots;
}

export default function ProductSelectionModal({
  isOpen,
  onClose,
  bakeryName,
  products,
  categories,
  productsByCategory,
  schedule,
  minDeliveryAmount,
  onSubmit,
  initialItems,
  unavailableItems,
}: ProductSelectionModalProps) {
  // Selection state
  const [orderItems, setOrderItems] = useState<OrderItem[]>([]);
  const [selectedDays, setSelectedDays] = useState<string[]>([]);
  const [selectedTime, setSelectedTime] = useState<string>('');
  const [orderMode, setOrderMode] = useState<'delivery' | 'reservation'>('delivery');
  const [deliveryFrequency, setDeliveryFrequency] = useState<DeliveryFrequency>('one-time');
  const [activeCategory, setActiveCategory] = useState<string | null>(null);

  // Allergen detail modal state
  const [allergenModalProduct, setAllergenModalProduct] = useState<Product | null>(null);

  // Mobile tray state
  const [trayExpanded, setTrayExpanded] = useState(false);

  // Hover-expand: track which card is expanded (for touch)
  const [expandedCardId, setExpandedCardId] = useState<string | null>(null);

  // Refs
  const modalRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  // Set initial active category
  useEffect(() => {
    if (categories.length > 0 && !activeCategory) {
      setActiveCategory(null); // null = show all
    }
  }, [categories, activeCategory]);

  // Pre-fill order items for re-order flow
  useEffect(() => {
    if (initialItems && initialItems.length > 0) {
      setOrderItems(initialItems);
    }
  }, [initialItems]);

  // Lock body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      previousFocusRef.current = document.activeElement as HTMLElement;
      document.body.style.overflow = 'hidden';
      // Focus the modal
      setTimeout(() => modalRef.current?.focus(), 50);
    } else {
      document.body.style.overflow = '';
      // Restore focus
      previousFocusRef.current?.focus();
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  // Schedule helpers
  const scheduleMap = useMemo(() => {
    const map = new Map<string, DaySchedule>();
    for (const ds of schedule) {
      map.set(ds.day, ds);
    }
    return map;
  }, [schedule]);

  const openDays = useMemo(() => {
    return DAY_ORDER.filter((day) => {
      const s = scheduleMap.get(day);
      return s && s.isOpen;
    });
  }, [scheduleMap]);

  const timeSlots = useMemo(() => {
    if (selectedDays.length === 0) return [];
    const s = scheduleMap.get(selectedDays[0]);
    if (!s || !s.isOpen) return [];
    return generateTimeSlots(s.openTime, s.closeTime);
  }, [selectedDays, scheduleMap]);

  // Order total in cents
  const total = useMemo(() => {
    return orderItems.reduce((sum, item) => sum + item.quantity * item.product.price, 0);
  }, [orderItems]);

  // Delivery minimum check
  const belowMinimum = orderMode === 'delivery' && minDeliveryAmount != null && total > 0 && total < minDeliveryAmount;

  // CTA enabled state
  const canSubmit = orderItems.length > 0 && selectedDays.length > 0 && selectedTime !== '' && !belowMinimum;

  // Product actions
  const quantityMap = useMemo(() => {
    const map = new Map<string, number>();
    for (const item of orderItems) {
      map.set(item.product.id, item.quantity);
    }
    return map;
  }, [orderItems]);

  const getQuantity = useCallback((productId: string): number => {
    return quantityMap.get(productId) ?? 0;
  }, [quantityMap]);

  const incrementProduct = useCallback((product: Product) => {
    setOrderItems((prev) => {
      const existing = prev.find((i) => i.product.id === product.id);
      if (existing) {
        return prev.map((i) =>
          i.product.id === product.id ? { ...i, quantity: i.quantity + 1 } : i
        );
      }
      return [...prev, { product, quantity: 1 }];
    });
  }, []);

  const decrementProduct = useCallback((productId: string) => {
    setOrderItems((prev) => {
      const existing = prev.find((i) => i.product.id === productId);
      if (!existing) return prev;
      if (existing.quantity <= 1) {
        return prev.filter((i) => i.product.id !== productId);
      }
      return prev.map((i) =>
        i.product.id === productId ? { ...i, quantity: i.quantity - 1 } : i
      );
    });
  }, []);

  const removeProduct = useCallback((productId: string) => {
    setOrderItems((prev) => prev.filter((i) => i.product.id !== productId));
  }, []);

  // Day toggle — reservation (pickup) mode only allows single day
  const toggleDay = useCallback((day: string) => {
    setSelectedDays((prev) => {
      if (orderMode === 'reservation') {
        // Single selection for pickup
        return prev.includes(day) ? [] : [day];
      }
      // Multi-select for delivery
      if (prev.includes(day)) {
        return prev.filter((d) => d !== day);
      }
      return [...prev, day];
    });
  }, [orderMode]);

  // When switching to reservation mode, trim to single day
  useEffect(() => {
    if (orderMode === 'reservation' && selectedDays.length > 1) {
      setSelectedDays(prev => [prev[0]]);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orderMode]);

  // Submit
  const handleSubmit = () => {
    if (!canSubmit) return;
    onSubmit(orderItems, selectedDays, selectedTime, orderMode, orderMode === 'delivery' ? deliveryFrequency : undefined);
  };

  // Close handler
  const handleClose = () => {
    onClose();
  };

  // Escape key
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      handleClose();
    }
  };

  // Products to display (filtered by category)
  const displayProducts = useMemo(() => {
    if (!activeCategory) return products;
    return productsByCategory[activeCategory] ?? [];
  }, [activeCategory, products, productsByCategory]);

  if (!isOpen) return null;

  return (
    <>
    <div
      className="psm-backdrop"
      onClick={handleClose}
    >
      <div
        ref={modalRef}
        className="psm"
        role="dialog"
        aria-modal="true"
        aria-label={`Compose your order — ${bakeryName}`}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Header */}
        <div className="psm__header">
          <h2 className="psm__title">Compose your order — {bakeryName}</h2>
          <button
            type="button"
            className="psm__close-btn"
            onClick={handleClose}
            aria-label="Close modal"
          >
            ×
          </button>
        </div>

        {/* Body: two-pane layout */}
        <div className="psm__body">
          {/* Left pane: products */}
          <div className="psm__products-pane">
            {/* Category filter chips */}
            <div className="psm__categories">
              <button
                type="button"
                className={`psm__category-chip ${activeCategory === null ? 'psm__category-chip--active' : ''}`}
                onClick={() => setActiveCategory(null)}
              >
                All
              </button>
              {categories.map((cat) => (
                <button
                  key={cat}
                  type="button"
                  className={`psm__category-chip ${activeCategory === cat ? 'psm__category-chip--active' : ''}`}
                  onClick={() => setActiveCategory(cat)}
                >
                  {cat}
                </button>
              ))}
            </div>

            {/* Product grid */}
            <div className="psm__product-grid">
              {displayProducts.map((product) => {
                const qty = getQuantity(product.id);
                const isExpanded = expandedCardId === product.id;
                return (
                  <div
                    key={product.id}
                    className={`psm__card ${qty > 0 ? 'psm__card--selected' : ''} ${isExpanded ? 'psm__card--expanded' : ''}`}
                    role="button"
                    tabIndex={0}
                    aria-label={`${product.name}, €${(product.price / 100).toFixed(2)}${qty > 0 ? `, ${qty} in order` : ''}`}
                    onClick={() => {
                      // Touch behavior: first tap expands, second tap adds
                      if ('ontouchstart' in window) {
                        if (!isExpanded) {
                          setExpandedCardId(product.id);
                        } else {
                          incrementProduct(product);
                        }
                      }
                    }}
                    onMouseEnter={() => setExpandedCardId(product.id)}
                    onMouseLeave={() => setExpandedCardId(null)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        incrementProduct(product);
                      }
                    }}
                  >
                    {/* Image */}
                    <div className="psm__card-image-wrap">
                      {product.photoUrl ? (
                        <img
                          src={product.photoUrl}
                          alt={product.name}
                          className="psm__card-image"
                          loading="lazy"
                        />
                      ) : (
                        <div className="psm__card-image-placeholder" aria-hidden="true" />
                      )}
                      {qty > 0 && (
                        <span className="psm__card-qty-badge">×{qty}</span>
                      )}
                    </div>

                    {/* Info (always visible) */}
                    <div className="psm__card-info">
                      <p className="psm__card-name">{product.name}</p>
                      <p className="psm__card-price">€{(product.price / 100).toFixed(2)}</p>
                      {product.healthScore != null && <HealthScoreDisplay score={product.healthScore} />}
                    </div>

                    {/* Allergen indicator */}
                    {product.allergens?.length > 0 && (
                      <AllergenIndicator
                        allergens={product.allergens}
                        productName={product.name}
                        onOpenModal={() => setAllergenModalProduct(product)}
                      />
                    )}

                    {/* Expanded details (visible on hover/expand) */}
                    <div className="psm__card-details">
                      {product.description && (
                        <p className="psm__card-description">{product.description}</p>
                      )}
                      {product.healthScore != null && <HealthScoreDisplay score={product.healthScore} />}
                      {/* Stepper + Ajouter */}
                      <div
                        className="psm__card-actions"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <div className="psm__card-stepper">
                          <button
                            type="button"
                            className="psm__stepper-btn"
                            onClick={() => decrementProduct(product.id)}
                            disabled={qty === 0}
                            aria-label={`Remove one ${product.name}`}
                          >
                            −
                          </button>
                          <span className="psm__stepper-count">{qty}</span>
                          <button
                            type="button"
                            className="psm__stepper-btn psm__stepper-btn--plus"
                            onClick={() => incrementProduct(product)}
                            aria-label={`Add one ${product.name}`}
                          >
                            +
                          </button>
                        </div>

                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Right pane: setup (desktop) */}
          <div className="psm__setup-pane">
            <SetupContent
              orderMode={orderMode}
              setOrderMode={setOrderMode}
              deliveryFrequency={deliveryFrequency}
              setDeliveryFrequency={setDeliveryFrequency}
              openDays={openDays}
              selectedDays={selectedDays}
              toggleDay={toggleDay}
              timeSlots={timeSlots}
              selectedTime={selectedTime}
              setSelectedTime={setSelectedTime}
              orderItems={orderItems}
              removeProduct={removeProduct}
              total={total}
              belowMinimum={belowMinimum}
              minDeliveryAmount={minDeliveryAmount}
              canSubmit={canSubmit}
              onSubmit={handleSubmit}
              unavailableItems={unavailableItems}
            />
          </div>
        </div>

        {/* Mobile bottom tray */}
        <div className={`psm__mobile-tray ${trayExpanded ? 'psm__mobile-tray--expanded' : ''}`}>
          <div
            className="psm__mobile-tray-header"
            onClick={() => setTrayExpanded(!trayExpanded)}
            role="button"
            tabIndex={0}
            aria-expanded={trayExpanded}
            aria-label="Order summary"
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                setTrayExpanded(!trayExpanded);
              }
            }}
          >
            <span className="psm__mobile-tray-mode">
              {orderMode === 'delivery'
                ? deliveryFrequency === 'weekly' ? 'Weekly' : 'Delivery'
                : 'Pickup'}
            </span>
            <span className="psm__mobile-tray-info">
              {selectedDays.length > 0
                ? selectedDays.map((d) => DAY_LABELS[d]).join(', ')
                : 'No day selected'}
              {selectedTime && ` · ${selectedTime}`}
            </span>
            <span className="psm__mobile-tray-total">
              €{(total / 100).toFixed(2)}
            </span>
            <button
              type="button"
              className="psm__mobile-tray-cta"
              disabled={!canSubmit}
              onClick={(e) => {
                e.stopPropagation();
                handleSubmit();
              }}
            >
              Validate →
            </button>
          </div>

          {trayExpanded && (
            <div className="psm__mobile-tray-body">
              <SetupContent
                orderMode={orderMode}
                setOrderMode={setOrderMode}
                deliveryFrequency={deliveryFrequency}
                setDeliveryFrequency={setDeliveryFrequency}
                openDays={openDays}
                selectedDays={selectedDays}
                toggleDay={toggleDay}
                timeSlots={timeSlots}
                selectedTime={selectedTime}
                setSelectedTime={setSelectedTime}
                orderItems={orderItems}
                removeProduct={removeProduct}
                total={total}
                belowMinimum={belowMinimum}
                minDeliveryAmount={minDeliveryAmount}
                canSubmit={canSubmit}
                onSubmit={handleSubmit}
                unavailableItems={unavailableItems}
              />
            </div>
          )}
        </div>
      </div>
    </div>

      {allergenModalProduct && (
        <AllergenDetailModal
          isOpen={!!allergenModalProduct}
          onClose={() => setAllergenModalProduct(null)}
          productName={allergenModalProduct.name}
          allergens={allergenModalProduct.allergens}
        />
      )}
    </>
  );
}

// ─── Setup Content (shared between desktop pane and mobile tray) ─────────────

interface SetupContentProps {
  orderMode: 'delivery' | 'reservation';
  setOrderMode: (mode: 'delivery' | 'reservation') => void;
  deliveryFrequency: DeliveryFrequency;
  setDeliveryFrequency: (freq: DeliveryFrequency) => void;
  openDays: string[];
  selectedDays: string[];
  toggleDay: (day: string) => void;
  timeSlots: string[];
  selectedTime: string;
  setSelectedTime: (time: string) => void;
  orderItems: OrderItem[];
  removeProduct: (productId: string) => void;
  total: number;
  belowMinimum: boolean;
  minDeliveryAmount?: number;
  canSubmit: boolean;
  onSubmit: () => void;
  unavailableItems?: { productName: string }[];
}

function SetupContent({
  orderMode,
  setOrderMode,
  deliveryFrequency,
  setDeliveryFrequency,
  openDays,
  selectedDays,
  toggleDay,
  timeSlots,
  selectedTime,
  setSelectedTime,
  orderItems,
  removeProduct,
  total,
  belowMinimum,
  minDeliveryAmount,
  canSubmit,
  onSubmit,
  unavailableItems,
}: SetupContentProps) {
  return (
    <div className="psm__setup">
      <h3 className="psm__setup-heading">Your order</h3>

      {/* Mode toggle */}
      <div className="psm__mode-toggle">
        <button
          type="button"
          className={`psm__mode-chip ${orderMode === 'delivery' ? 'psm__mode-chip--active' : ''}`}
          onClick={() => setOrderMode('delivery')}
          aria-pressed={orderMode === 'delivery'}
        >
          Delivery
        </button>
        <button
          type="button"
          className={`psm__mode-chip ${orderMode === 'reservation' ? 'psm__mode-chip--active' : ''}`}
          onClick={() => setOrderMode('reservation')}
          aria-pressed={orderMode === 'reservation'}
        >
          Pickup
        </button>
      </div>

      {/* Delivery frequency toggle (only in delivery mode) */}
      {orderMode === 'delivery' && (
        <div className="psm__frequency-toggle">
          <button
            type="button"
            className={`psm__freq-chip ${deliveryFrequency === 'one-time' ? 'psm__freq-chip--active' : ''}`}
            onClick={() => setDeliveryFrequency('one-time')}
            aria-pressed={deliveryFrequency === 'one-time'}
          >
            One-time
          </button>
          <button
            type="button"
            className={`psm__freq-chip ${deliveryFrequency === 'weekly' ? 'psm__freq-chip--active' : ''}`}
            onClick={() => setDeliveryFrequency('weekly')}
            aria-pressed={deliveryFrequency === 'weekly'}
          >
            Weekly
          </button>
        </div>
      )}

      {/* Weekly info banner */}
      {orderMode === 'delivery' && deliveryFrequency === 'weekly' && (
        <div className="psm__weekly-info" role="note">
          🔄 This order will repeat every week on the selected day(s). You will be charged weekly.
        </div>
      )}

      {/* Day chips */}
      <div className="psm__day-section">
        <p className="psm__day-label">
          {orderMode === 'reservation'
            ? 'Pick a day'
            : deliveryFrequency === 'weekly'
              ? 'Delivery day(s)'
              : 'Select day(s)'}
        </p>
        <div className="psm__day-chips">
          {openDays.map((day) => (
            <button
              key={day}
              type="button"
              className={`psm__day-chip ${selectedDays.includes(day) ? 'psm__day-chip--active' : ''}`}
              onClick={() => toggleDay(day)}
              aria-pressed={selectedDays.includes(day)}
            >
              {DAY_LABELS[day]}
            </button>
          ))}
        </div>
      </div>

      {/* Time slot dropdown */}
      <div className="psm__time-slot">
        <select
          className="psm__time-select"
          value={selectedTime}
          onChange={(e) => setSelectedTime(e.target.value)}
          disabled={timeSlots.length === 0}
          aria-label="Select time slot"
        >
          <option value="">Select a time…</option>
          {timeSlots.map((slot) => (
            <option key={slot} value={slot}>
              {slot}
            </option>
          ))}
        </select>
      </div>

      {/* Minimum delivery warning */}
      {belowMinimum && minDeliveryAmount != null && (
        <div className="psm__min-warning" role="alert">
          Minimum delivery: €{(minDeliveryAmount / 100).toFixed(2)}
        </div>
      )}

      {/* Unavailable items from re-order */}
      {unavailableItems && unavailableItems.length > 0 && (
        <div className="psm__unavailable-items">
          {unavailableItems.map((item, idx) => (
            <div key={idx} className="psm__unavailable-item">
              <span className="psm__unavailable-item-name">{item.productName}</span>
              <span className="psm__unavailable-item-label">No longer available</span>
            </div>
          ))}
        </div>
      )}

      {/* Line items */}
      <div className="psm__line-items">
        {orderItems.length === 0 ? (
          <p className="psm__line-items-empty">No products selected</p>
        ) : (
          <>
            {orderItems.map((item) => (
              <div key={item.product.id} className="psm__line-item">
                <span className="psm__line-item-text">
                  {item.quantity}× {item.product.name} — €{((item.quantity * item.product.price) / 100).toFixed(2)}
                </span>
                <button
                  type="button"
                  className="psm__line-item-remove"
                  onClick={() => removeProduct(item.product.id)}
                  aria-label={`Remove ${item.product.name}`}
                >
                  ×
                </button>
              </div>
            ))}
            <div className="psm__line-items-divider" />
            <div className="psm__line-items-total">
              <span>Total</span>
              <span>€{(total / 100).toFixed(2)}</span>
            </div>
            {orderMode === 'delivery' && deliveryFrequency === 'weekly' && (
              <p className="psm__line-items-recur">
                Charged weekly · €{(total / 100).toFixed(2)}/week
              </p>
            )}
          </>
        )}
      </div>

      {/* Submit CTA (desktop only — mobile uses tray button) */}
      <button
        type="button"
        className="psm__submit-btn"
        disabled={!canSubmit}
        onClick={onSubmit}
        aria-disabled={!canSubmit}
      >
        {orderMode === 'delivery' && deliveryFrequency === 'weekly' ? 'Subscribe →' : 'Validate →'}
      </button>
    </div>
  );
}
