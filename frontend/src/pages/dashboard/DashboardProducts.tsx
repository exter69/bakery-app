import { useState, useEffect, useCallback, useMemo } from 'react';
import { fetchMyBakery, fetchProducts, createProduct, updateProduct } from '../../api/seller';
import { ApiError } from '../../api/client';
import type { Product } from '../../types/bakery';
import { FilterChips } from '../../components/pro/FilterChips';
import { ProductCard } from '../../components/pro/ProductCard';
import { ErrorBanner } from '../../components/pro/ErrorBanner';
import AllergenMultiSelect from '../../components/dashboard/AllergenMultiSelect';
import HealthScoreInput from '../../components/dashboard/HealthScoreInput';
import ImageUpload from '../../components/dashboard/ImageUpload';
import './DashboardProducts.css';
import './Dashboard.css';

interface ProductForm {
  name: string;
  description: string;
  price: string;
  category: string;
  photoUrl: string;
  allergens: string[];
  healthScore: number | null;
}

const emptyForm: ProductForm = {
  name: '',
  description: '',
  price: '',
  category: '',
  photoUrl: '',
  allergens: [],
  healthScore: null,
};

/** Day labels for availability toggles */
const DAY_LABELS = ['L', 'M', 'M', 'J', 'V', 'S', 'D'] as const;

/** All categories filter value */
const ALL_FILTER = 'toutes';

export default function DashboardProducts() {
  const [bakeryId, setBakeryId] = useState<string | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Stock is local state — resets to zero each evening
  const [stockMap, setStockMap] = useState<Map<string, number>>(new Map());

  // Category filter
  const [selectedCategory, setSelectedCategory] = useState<string>(ALL_FILTER);

  // Day availability toggles
  const [activeDays, setActiveDays] = useState<boolean[]>([false, false, true, true, true, false, false]);

  // Modal state for product creation
  const [showModal, setShowModal] = useState(false);
  const [form, setForm] = useState<ProductForm>(emptyForm);
  const [submitting, setSubmitting] = useState(false);

  const loadProducts = useCallback(async (bId: string) => {
    try {
      const prods = await fetchProducts(bId);
      setProducts(prods);
      // Initialize stock to 0 for all products (resets nightly)
      setStockMap(new Map(prods.map((p) => [p.id, 0])));
      setError(null);
    } catch {
      setError('Impossible de charger les produits.');
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    async function init() {
      const b = await fetchMyBakery();
      if (cancelled) return;
      if (b) {
        setBakeryId(b.id);
        await loadProducts(b.id);
      } else {
        setError('Aucune boulangerie trouvée.');
      }
      setLoading(false);
    }
    init();
    return () => { cancelled = true; };
  }, [loadProducts]);

  // Build category chip options from product data
  const categoryOptions = useMemo(() => {
    const cats = new Set(products.map((p) => p.category.toLowerCase()));
    const options: { value: string; label: string }[] = [
      { value: ALL_FILTER, label: 'Toutes' },
    ];
    // Prioritize known categories in order, then any extras
    const knownOrder = ['viennoiseries', 'pains', 'pâtisseries'];
    const knownLabels: Record<string, string> = {
      viennoiseries: 'Viennoiseries',
      pains: 'Pains',
      pâtisseries: 'Pâtisseries',
    };
    for (const cat of knownOrder) {
      if (cats.has(cat)) {
        options.push({ value: cat, label: knownLabels[cat] });
        cats.delete(cat);
      }
    }
    // Add remaining categories
    for (const cat of cats) {
      options.push({ value: cat, label: cat.charAt(0).toUpperCase() + cat.slice(1) });
    }
    return options;
  }, [products]);

  // Filtered products
  const filteredProducts = useMemo(() => {
    if (selectedCategory === ALL_FILTER) return products;
    return products.filter((p) => p.category.toLowerCase() === selectedCategory);
  }, [products, selectedCategory]);

  // Stock change handler — updates local state and calls API
  const handleStockChange = useCallback(
    async (productId: string, delta: number) => {
      setStockMap((prev) => {
        const copy = new Map(prev);
        const current = copy.get(productId) ?? 0;
        const next = Math.max(0, current + delta);
        copy.set(productId, next);
        return copy;
      });
      // Call API to persist the stock change concept via updateProduct
      try {
        await updateProduct(productId, {});
      } catch (err) {
        // 409 Conflict: another user modified the product concurrently
        if (err instanceof ApiError && err.status === 409) {
          setError('Les données ont été modifiées. Rechargement…');
          if (bakeryId) {
            await loadProducts(bakeryId);
          }
        }
        // Other errors are silent — stock is local anyway
      }
    },
    [bakeryId, loadProducts],
  );

  // Visibility toggle
  const handleToggleVisibility = useCallback(
    async (productId: string) => {
      const product = products.find((p) => p.id === productId);
      if (!product) return;
      try {
        const updated = await updateProduct(productId, { isAvailable: !product.isAvailable });
        setProducts((prev) => prev.map((p) => (p.id === updated.id ? updated : p)));
      } catch {
        setError('Impossible de modifier la visibilité.');
      }
    },
    [products],
  );

  // Day toggle handler
  const toggleDay = (index: number) => {
    setActiveDays((prev) => prev.map((active, i) => (i === index ? !active : active)));
  };

  // Product creation form
  const openAddModal = () => {
    setForm(emptyForm);
    setShowModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!bakeryId) return;
    setSubmitting(true);
    try {
      const priceInCents = Math.round(parseFloat(form.price) * 100);
      await createProduct(bakeryId, {
        name: form.name,
        description: form.description,
        price: priceInCents,
        category: form.category,
        photoUrl: form.photoUrl,
        allergens: form.allergens,
        healthScore: form.healthScore,
      });
      setShowModal(false);
      await loadProducts(bakeryId);
    } catch {
      setError('Impossible de créer le produit.');
    } finally {
      setSubmitting(false);
    }
  };

  // Loading state
  if (loading) {
    return <div className="dash-loading">Chargement des produits…</div>;
  }

  // Error state with retry — full-page error only when no stale data exists
  if (error && products.length === 0) {
    return (
      <div className="pro-products">
        <h1 className="pro-products__title">Menu &amp; stock</h1>
        <ErrorBanner
          message={error}
          onRetry={bakeryId ? () => loadProducts(bakeryId) : undefined}
        />
      </div>
    );
  }

  return (
    <div className="pro-products">
      {/* Header */}
      <div className="pro-products__header">
        <h1 className="pro-products__title">Menu &amp; stock</h1>
      </div>

      {/* Inline error (non-blocking — stale data retained) */}
      {error && <ErrorBanner message={error} onRetry={bakeryId ? () => loadProducts(bakeryId) : undefined} />}

      {/* Toolbar: category chips + add button */}
      <div className="pro-products__toolbar">
        <FilterChips
          options={categoryOptions}
          selected={selectedCategory}
          onChange={setSelectedCategory}
          variant="category"
        />
        <button
          type="button"
          className="pro-products__add-btn"
          onClick={openAddModal}
        >
          + Nouveau produit
        </button>
      </div>

      {/* Product cards list */}
      {filteredProducts.length === 0 ? (
        <div className="pro-products__empty">
          Aucun produit dans cette catégorie.
        </div>
      ) : (
        <div className="pro-products__list">
          {filteredProducts.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              stock={stockMap.get(product.id) ?? 0}
              onStockChange={handleStockChange}
              onToggleVisibility={handleToggleVisibility}
            />
          ))}
        </div>
      )}

      {/* Day availability toggles */}
      <div className="pro-products__availability">
        <p className="pro-products__availability-label">
          Disponibilité par défaut
        </p>
        <div className="pro-products__day-toggles" role="group" aria-label="Jours de disponibilité">
          {DAY_LABELS.map((label, index) => (
            <button
              key={index}
              type="button"
              className={`pro-products__day-toggle ${activeDays[index] ? 'pro-products__day-toggle--active' : ''}`}
              onClick={() => toggleDay(index)}
              aria-pressed={activeDays[index]}
              aria-label={`${label} ${activeDays[index] ? 'actif' : 'inactif'}`}
            >
              {label}
            </button>
          ))}
        </div>
        <p className="pro-products__note">
          le stock se remet à zéro chaque soir ↺
        </p>
      </div>

      {/* Product creation modal */}
      {showModal && (
        <div className="dash-modal-overlay" onClick={() => setShowModal(false)}>
          <div className="dash-modal" onClick={(e) => e.stopPropagation()}>
            <h2 className="dash-modal__title">Nouveau produit</h2>
            <form className="dash-form" onSubmit={handleSubmit}>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-name">Nom</label>
                <input
                  id="prod-name"
                  className="dash-form__input"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  required
                />
              </div>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-desc">Description</label>
                <textarea
                  id="prod-desc"
                  className="dash-form__textarea"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  rows={2}
                />
              </div>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-price">Prix (€)</label>
                <input
                  id="prod-price"
                  className="dash-form__input"
                  type="number"
                  step="0.01"
                  min="0"
                  value={form.price}
                  onChange={(e) => setForm({ ...form, price: e.target.value })}
                  required
                />
              </div>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-category">Catégorie</label>
                <input
                  id="prod-category"
                  className="dash-form__input"
                  value={form.category}
                  onChange={(e) => setForm({ ...form, category: e.target.value })}
                  required
                />
              </div>
              <ImageUpload
                value={form.photoUrl}
                onChange={(url) => setForm({ ...form, photoUrl: url })}
                label="Photo"
                type="products"
              />
              <AllergenMultiSelect
                selected={form.allergens}
                onChange={(allergens) => setForm({ ...form, allergens })}
              />
              <HealthScoreInput
                value={form.healthScore}
                onChange={(healthScore) => setForm({ ...form, healthScore })}
              />
              <div className="dash-modal__actions">
                <button
                  type="button"
                  className="dash-btn dash-btn--secondary"
                  onClick={() => setShowModal(false)}
                >
                  Annuler
                </button>
                <button
                  type="submit"
                  className="dash-btn dash-btn--primary"
                  disabled={submitting}
                >
                  {submitting ? 'Création…' : 'Créer'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
