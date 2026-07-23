import { useState, useEffect, useCallback } from 'react';
import { useI18n } from '../../i18n';
import { listDeliveries, editOrder, getProducts } from '../../api/b2b-client';
import { ErrorState } from '../../components/ErrorState';
import type { Product } from '../../types/bakery';

interface DeliveryEntry {
  id: string;
  bakeryId: string;
  status: string;
  items: { productId: string; quantity: number }[];
  createdAt: string;
  deliverySiteId?: string;
  subtotalHt?: number;
  discountAmount?: number;
  tvaAmount?: number;
}

const PAGE_SIZE = 20;

export default function LivraisonsPage() {
  const { t } = useI18n();
  const [deliveries, setDeliveries] = useState<DeliveryEntry[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<{ bakeryId?: string; status?: string; dateFrom?: string; dateTo?: string }>({});

  // Edit modal state
  const [editingOrderId, setEditingOrderId] = useState<string | null>(null);
  const [editItems, setEditItems] = useState<{ productId: string; quantity: number }[]>([]);
  const [editProducts, setEditProducts] = useState<Product[]>([]);
  const [editError, setEditError] = useState<string | null>(null);
  const [editSaving, setEditSaving] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await listDeliveries({ ...filters, page });
      setDeliveries(result.items);
      setTotal(result.total);
    } catch {
      setError(t('comptoir.common.loadError'));
    } finally {
      setLoading(false);
    }
  }, [page, filters, t]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const upcoming = deliveries.filter((d) => ['confirmed', 'preparing', 'ready'].includes(d.status));
  const past = deliveries.filter((d) => d.status === 'delivered');

  const formatDate = (iso: string) => {
    const d = new Date(iso);
    return d.toLocaleDateString();
  };

  const statusLabel = (status: string) => {
    const key = `comptoir.deliveries.status.${status}` as const;
    return t(key);
  };

  const handleEditOpen = async (delivery: DeliveryEntry) => {
    setEditingOrderId(delivery.id);
    setEditItems(delivery.items.map((i) => ({ ...i })));
    setEditError(null);
    try {
      const categorizedProducts = await getProducts(delivery.bakeryId);
      const allProducts = Object.values(categorizedProducts).flat();
      setEditProducts(allProducts);
    } catch {
      setEditProducts([]);
    }
  };

  const handleEditClose = () => {
    setEditingOrderId(null);
    setEditItems([]);
    setEditProducts([]);
    setEditError(null);
  };

  const handleEditQuantity = (productId: string, quantity: number) => {
    setEditItems((prev) =>
      prev.map((item) => item.productId === productId ? { ...item, quantity: Math.max(0, quantity) } : item)
    );
  };

  const handleEditSave = async () => {
    if (!editingOrderId) return;
    setEditSaving(true);
    setEditError(null);
    try {
      const filteredItems = editItems.filter((i) => i.quantity > 0);
      await editOrder(editingOrderId, filteredItems);
      handleEditClose();
      fetchData();
    } catch {
      setEditError(t('comptoir.error.generic'));
    } finally {
      setEditSaving(false);
    }
  };

  return (
    <div className="livraisons-page">
      <h1>{t('comptoir.deliveries.title')}</h1>

      <div className="livraisons-page__filters">
        <input
          type="date"
          value={filters.dateFrom ?? ''}
          onChange={(e) => setFilters((f) => ({ ...f, dateFrom: e.target.value || undefined }))}
          aria-label={t('comptoir.deliveries.filter.dateFrom')}
        />
        <input
          type="date"
          value={filters.dateTo ?? ''}
          onChange={(e) => setFilters((f) => ({ ...f, dateTo: e.target.value || undefined }))}
          aria-label={t('comptoir.deliveries.filter.dateTo')}
        />
        <select
          value={filters.status ?? ''}
          onChange={(e) => setFilters((f) => ({ ...f, status: e.target.value || undefined }))}
          aria-label={t('comptoir.deliveries.filter.status')}
        >
          <option value="">{t('comptoir.deliveries.filter.status')}</option>
          <option value="confirmed">{t('comptoir.deliveries.status.confirmed')}</option>
          <option value="preparing">{t('comptoir.deliveries.status.preparing')}</option>
          <option value="ready">{t('comptoir.deliveries.status.ready')}</option>
          <option value="delivered">{t('comptoir.deliveries.status.delivered')}</option>
        </select>
      </div>

      {loading ? (
        <p>{t('comptoir.common.loading')}</p>
      ) : error ? (
        <ErrorState message={error} onRetry={fetchData} retryLabel={t('comptoir.common.retry')} />
      ) : deliveries.length === 0 ? (
        <p className="livraisons-page__empty">{t('comptoir.deliveries.empty')}</p>
      ) : (
        <>
          {upcoming.length > 0 && (
            <section>
              <h2>{t('comptoir.deliveries.upcoming')}</h2>
              <table className="livraisons-page__table">
                <thead>
                  <tr>
                    <th>{t('comptoir.invoices.date')}</th>
                    <th>{t('comptoir.invoices.status')}</th>
                    <th>{t('comptoir.commander.quantity')}</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {upcoming.map((d) => (
                    <tr key={d.id}>
                      <td>{formatDate(d.createdAt)}</td>
                      <td>{statusLabel(d.status)}</td>
                      <td>{t('comptoir.deliveries.itemCount', { n: String(d.items.length) })}</td>
                      <td>
                        <button
                          type="button"
                          className="livraisons-page__edit-btn"
                          onClick={() => handleEditOpen(d)}
                        >
                          {t('comptoir.deliveries.edit')}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          )}

          {past.length > 0 && (
            <section>
              <h2>{t('comptoir.deliveries.past')}</h2>
              <table className="livraisons-page__table">
                <thead>
                  <tr>
                    <th>{t('comptoir.invoices.date')}</th>
                    <th>{t('comptoir.invoices.status')}</th>
                    <th>{t('comptoir.commander.quantity')}</th>
                  </tr>
                </thead>
                <tbody>
                  {past.map((d) => (
                    <tr key={d.id}>
                      <td>{formatDate(d.createdAt)}</td>
                      <td>{statusLabel(d.status)}</td>
                      <td>{t('comptoir.deliveries.itemCount', { n: String(d.items.length) })}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          )}

          <div className="livraisons-page__pagination">
            <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              {t('comptoir.common.previous')}
            </button>
            <span>{page} / {totalPages || 1}</span>
            <button type="button" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              {t('comptoir.common.next')}
            </button>
          </div>
        </>
      )}

      {/* Inline edit modal */}
      {editingOrderId && (
        <div className="livraisons-page__modal-overlay" onClick={handleEditClose}>
          <div className="livraisons-page__modal" onClick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
            <h3>{t('comptoir.deliveries.edit')}</h3>
            {editError && <p className="livraisons-page__modal-error" role="alert">{editError}</p>}
            <table className="livraisons-page__edit-table">
              <thead>
                <tr>
                  <th>{t('comptoir.commander.product')}</th>
                  <th>{t('comptoir.commander.quantity')}</th>
                </tr>
              </thead>
              <tbody>
                {editItems.map((item) => {
                  const product = editProducts.find((p) => p.id === item.productId);
                  return (
                    <tr key={item.productId}>
                      <td>{product?.name ?? item.productId}</td>
                      <td>
                        <input
                          type="number"
                          min="0"
                          value={item.quantity}
                          onChange={(e) => handleEditQuantity(item.productId, Number(e.target.value))}
                        />
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            <div className="livraisons-page__modal-actions">
              <button type="button" onClick={handleEditClose}>
                {t('comptoir.common.cancel')}
              </button>
              <button type="button" onClick={handleEditSave} disabled={editSaving}>
                {t('comptoir.common.save')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
