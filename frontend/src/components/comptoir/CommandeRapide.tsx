import { useState, useEffect } from 'react';
import { useI18n } from '../../i18n';
import { getProducts, getLastOrder } from '../../api/b2b-client';
import { SavedListPicker } from './SavedListPicker';
import type { Product } from '../../types/bakery';
import type { UseB2BCartReturn } from '../../hooks/useB2BCart';

interface Props {
  bakeryId: string;
  bakeryName: string;
  cart: UseB2BCartReturn;
}

export function CommandeRapide({ bakeryId, bakeryName, cart }: Props) {
  const { t } = useI18n();
  const [products, setProducts] = useState<Record<string, Product[]>>({});
  const [loading, setLoading] = useState(true);
  const [savingList, setSavingList] = useState(false);
  const [listName, setListName] = useState('');

  useEffect(() => {
    setLoading(true);
    getProducts(bakeryId)
      .then(setProducts)
      .catch(() => setProducts({}))
      .finally(() => setLoading(false));
  }, [bakeryId]);

  const handleQuantityChange = (product: Product, qty: number) => {
    if (qty <= 0) {
      cart.removeItem(bakeryId, product.id);
    } else {
      const currentQty = cart.getItemQuantity(bakeryId, product.id);
      if (currentQty === 0) {
        cart.addItem(bakeryId, bakeryName, {
          productId: product.id,
          productName: product.name,
          unitPrice: product.price,
        }, qty);
      } else {
        cart.updateQuantity(bakeryId, product.id, qty);
      }
    }
  };

  const handleRepeatLast = async () => {
    try {
      const lastOrder = await getLastOrder(bakeryId);
      if (!lastOrder || !lastOrder.items) return;
      // Clear the current group first
      cart.clearGroup(bakeryId);
      // Add all items from the last order
      const allProducts = Object.values(products).flat();
      for (const item of lastOrder.items) {
        const product = allProducts.find((p) => p.id === item.productId);
        if (product && product.isAvailable) {
          cart.addItem(bakeryId, bakeryName, {
            productId: product.id,
            productName: product.name,
            unitPrice: product.price,
          }, item.quantity);
        }
      }
    } catch {
      // Silently fail if no previous order
    }
  };

  const handleSaveList = async () => {
    if (!listName.trim()) return;
    const items = cart.getGroupItems(bakeryId);
    if (items.length === 0) return;
    setSavingList(true);
    try {
      const { createSavedList } = await import('../../api/b2b-client');
      await createSavedList({
        bakeryId,
        name: listName.trim(),
        items: items.map((i) => ({ productId: i.productId, quantity: i.quantity })),
      });
      setListName('');
    } catch {
      // Silently fail
    } finally {
      setSavingList(false);
    }
  };

  const handleListSelect = (listItems: { productId: string; quantity: number }[]) => {
    cart.clearGroup(bakeryId);
    const allProducts = Object.values(products).flat();
    for (const item of listItems) {
      const product = allProducts.find((p) => p.id === item.productId);
      if (product && product.isAvailable) {
        cart.addItem(bakeryId, bakeryName, {
          productId: product.id,
          productName: product.name,
          unitPrice: product.price,
        }, item.quantity);
      }
    }
  };

  if (loading) return <p>{t('comptoir.common.loading')}</p>;

  const categories = Object.keys(products);

  return (
    <div className="commande-rapide">
      <div className="commande-rapide__actions">
        <SavedListPicker bakeryId={bakeryId} onSelect={handleListSelect} />
        <button type="button" className="commande-rapide__btn" onClick={handleRepeatLast}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
          {t('comptoir.commander.repeatLast')}
        </button>
        <div className="commande-rapide__save-list">
          <input
            type="text"
            value={listName}
            onChange={(e) => setListName(e.target.value)}
            placeholder={t('comptoir.commander.saveList')}
          />
          <button type="button" onClick={handleSaveList} disabled={savingList || !listName.trim()}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
          </button>
        </div>
      </div>

      <table className="commande-rapide__table">
        <thead>
          <tr>
            <th>{t('comptoir.commander.product')}</th>
            <th>{t('comptoir.commander.unitPrice')}</th>
            <th>{t('comptoir.commander.quantity')}</th>
          </tr>
        </thead>
        <tbody>
          {categories.map((category) => (
            <>{/* Category header row */}
              <tr key={`cat-${category}`} className="commande-rapide__category-row">
                <td colSpan={3}>{category}</td>
              </tr>
              {products[category].map((product) => {
                const qty = cart.getItemQuantity(bakeryId, product.id);
                return (
                  <tr
                    key={product.id}
                    className={product.isAvailable ? '' : 'commande-rapide__row--disabled'}
                  >
                    <td>{product.name}</td>
                    <td>{(product.price / 100).toFixed(2)} EUR</td>
                    <td>
                      {product.isAvailable ? (
                        <input
                          type="number"
                          min="0"
                          value={qty}
                          onChange={(e) => handleQuantityChange(product, parseInt(e.target.value) || 0)}
                          aria-label={`${t('comptoir.commander.quantity')} ${product.name}`}
                        />
                      ) : (
                        <span className="commande-rapide__unavailable">{t('comptoir.commander.unavailable')}</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </>
          ))}
        </tbody>
      </table>
    </div>
  );
}
