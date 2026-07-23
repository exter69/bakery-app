import { memo } from 'react';
import type { Product } from '../../types/bakery';
import { StockStepper } from './StockStepper';
import './ProductCard.css';

export interface ProductCardProps {
  product: Product;
  stock: number;
  onStockChange: (productId: string, delta: number) => void;
  onToggleVisibility: (productId: string) => void;
}

export const ProductCard = memo(function ProductCard({
  product,
  stock,
  onStockChange,
  onToggleVisibility,
}: ProductCardProps) {
  const isHidden = !product.isAvailable;

  return (
    <article
      className={`product-card ${isHidden ? 'product-card--hidden' : ''}`}
    >
      {/* Photo */}
      <div className="product-card__photo-wrapper">
        {product.photoUrl ? (
          <img
            className="product-card__photo"
            src={product.photoUrl}
            alt={product.name}
            loading="lazy"
          />
        ) : (
          <div className="product-card__photo-placeholder" aria-hidden="true">
            photo
          </div>
        )}
      </div>

      {/* Info */}
      <div className="product-card__info">
        <h3 className="product-card__name">{product.name}</h3>
        {product.description && (
          <p className="product-card__description">{product.description}</p>
        )}
        {product.allergens.length > 0 && (
          <p className="product-card__allergens">
            allergènes&nbsp;: {product.allergens.join(', ')}
          </p>
        )}
      </div>

      {/* Actions column */}
      <div className="product-card__actions">
        <span className="product-card__price">
          {(product.price / 100).toFixed(2)}&nbsp;€
        </span>

        <StockStepper
          value={stock}
          onChange={(newValue) => onStockChange(product.id, newValue - stock)}
        />

        <button
          type="button"
          className={`product-card__visibility ${isHidden ? 'product-card__visibility--hidden' : 'product-card__visibility--visible'}`}
          onClick={() => onToggleVisibility(product.id)}
          aria-pressed={product.isAvailable}
          aria-label={isHidden ? 'Rendre visible' : 'Masquer le produit'}
        >
          {isHidden ? 'masqué' : 'en vente'}
        </button>
      </div>
    </article>
  );
});
