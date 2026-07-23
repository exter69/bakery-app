import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchOrderHistory, storeReorderData } from '../api/orders';
import { ApiError } from '../api/client';
import { useI18n } from '../i18n';
import type { ScheduleEntry } from '../types/order';
import './OrderHistoryPage.css';

const PAGE_SIZE = 20;

function formatDate(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleDateString(undefined, {
    dateStyle: 'long',
  });
}

function formatCurrency(cents: number): string {
  return `€${(cents / 100).toFixed(2)}`;
}

function formatItemsList(items: { productName: string; quantity: number }[]): string {
  return items.map((item) => `${item.quantity}x ${item.productName}`).join(', ');
}

export default function OrderHistoryPage() {
  const { t } = useI18n();
  const navigate = useNavigate();

  const [entries, setEntries] = useState<ScheduleEntry[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const loadHistory = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetchOrderHistory(page);
      setEntries(response.items);
      setTotal(response.total);
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Failed to load order history';
      setError(message);
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    loadHistory();
  }, [loadHistory]);

  function handleReorder(entry: ScheduleEntry) {
    storeReorderData({
      bakeryId: entry.bakeryId,
      items: entry.items.map((item) => ({
        productId: item.productId,
        productName: item.productName,
        quantity: item.quantity,
      })),
    });
    navigate(`/bakeries/${entry.bakeryId}`);
  }

  return (
    <div className="order-history">
      <h1 className="order-history__title">{t('history.title')}</h1>

      {/* Loading */}
      {loading && (
        <div className="order-history__loading">
          <div className="spinner" />
          <p>Loading...</p>
        </div>
      )}

      {/* Error */}
      {!loading && error && (
        <div className="order-history__error">
          <p>{error}</p>
          <button className="btn btn--outline" onClick={loadHistory}>
            Retry
          </button>
        </div>
      )}

      {/* Empty state */}
      {!loading && !error && entries.length === 0 && (
        <div className="order-history__empty">
          <p>{t('history.empty')}</p>
        </div>
      )}

      {/* Order cards */}
      {!loading && !error && entries.length > 0 && (
        <>
          <div className="order-history__list">
            {entries.map((entry) => (
              <article key={entry.id} className="order-history__card">
                <div className="order-history__card-header">
                  <time className="order-history__date">{formatDate(entry.createdAt)}</time>
                  <span className={`order-history__status-badge order-history__status-badge--${entry.status}`}>
                    {entry.status === 'delivered' ? t('history.status.delivered') : t('history.status.pickedUp')}
                  </span>
                </div>

                <div className="order-history__card-body">
                  <div className="order-history__meta">
                    <span className={`order-history__type-badge order-history__type-badge--${entry.type}`}>
                      {entry.type === 'order' ? t('history.type.delivery') : t('history.type.pickup')}
                    </span>
                  </div>
                  <p className="order-history__items">{formatItemsList(entry.items)}</p>
                  <p className="order-history__total">{formatCurrency(entry.totalAmount)}</p>
                </div>

                <div className="order-history__card-footer">
                  <button
                    type="button"
                    className="order-history__reorder-btn"
                    onClick={() => handleReorder(entry)}
                  >
                    {t('history.reorder')}
                  </button>
                </div>
              </article>
            ))}
          </div>

          {/* Pagination */}
          <div className="order-history__pagination">
            <button
              className="order-history__page-btn"
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
            >
              {t('history.previous')}
            </button>
            <span className="order-history__page-info">
              {page} / {totalPages}
            </span>
            <button
              className="order-history__page-btn"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              {t('history.next')}
            </button>
          </div>
        </>
      )}
    </div>
  );
}
