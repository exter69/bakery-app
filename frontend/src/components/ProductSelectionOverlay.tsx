import type { Product } from '../types/bakery';
import './ProductSelectionOverlay.css';

/** Item in the selection list with quantity */
export interface SelectedItem {
  product: Product;
  quantity: number;
}

interface ProductSelectionOverlayProps {
  /** Whether the overlay is currently active */
  isActive: boolean;
  /** Products available for selection */
  products: Product[];
  /** Currently selected items with quantities */
  selectedItems: SelectedItem[];
  /** Called when a product is clicked — adds or increments */
  onProductClick: (product: Product) => void;
  /** Called when user is done selecting */
  onDone: () => void;
}

/**
 * Dark overlay displayed during product selection mode.
 * Shows all products in a clickable grid. Clicking a product
 * adds it to the active panel or increments its quantity.
 *
 * Activates within 100ms via CSS class toggle (no heavy JS).
 */
export default function ProductSelectionOverlay({
  isActive,
  products,
  selectedItems,
  onProductClick,
  onDone,
}: ProductSelectionOverlayProps) {
  /** Find current quantity for a given product */
  function getQuantity(productId: string): number {
    const item = selectedItems.find((i) => i.product.id === productId);
    return item ? item.quantity : 0;
  }

  return (
    <div
      className={`product-selection-overlay${isActive ? ' product-selection-overlay--active' : ''}`}
      role="dialog"
      aria-modal="true"
      aria-label="Product selection"
      aria-hidden={!isActive}
    >
      {/* Dark backdrop */}
      <div className="product-selection-overlay__backdrop" onClick={onDone} />

      {/* Content on top */}
      <div className="product-selection-overlay__content">
        <div className="product-selection-overlay__header">
          <h2 className="product-selection-overlay__title">Select Products</h2>
          <button
            type="button"
            className="product-selection-overlay__done-btn"
            onClick={onDone}
          >
            Done
          </button>
        </div>

        <div className="product-selection-overlay__grid">
          {products.map((product) => {
            const qty = getQuantity(product.id);
            return (
              <div
                key={product.id}
                className="product-selection-overlay__product"
                role="button"
                tabIndex={isActive ? 0 : -1}
                aria-label={`Add ${product.name} to selection${qty > 0 ? `, currently ${qty} selected` : ''}`}
                onClick={() => onProductClick(product)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    onProductClick(product);
                  }
                }}
              >
                {product.photoUrl && (
                  <img
                    src={product.photoUrl}
                    alt={product.name}
                    className="product-selection-overlay__product-photo"
                    loading="lazy"
                  />
                )}
                <div className="product-selection-overlay__product-info">
                  <p className="product-selection-overlay__product-name">
                    {product.name}
                  </p>
                  <p className="product-selection-overlay__product-price">
                    €{(product.price / 100).toFixed(2)}
                  </p>
                </div>
                {qty > 0 && (
                  <span className="product-selection-overlay__quantity-badge">
                    {qty}
                  </span>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
