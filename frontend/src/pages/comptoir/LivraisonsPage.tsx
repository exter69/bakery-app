import { useState, useEffect, useCallback } from 'react';
import { useI18n } from '../../i18n';
import { listDeliveries } from '../../api/b2b-client';

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
  const [filters, setFilters] = useState<{ bakeryId?: string; status?: string; dateFrom?: string; dateTo?: string }>({});

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const result = await listDeliveries({ ...filters, page });
      setDeliveries(result.items);
      setTotal(result.total);
    } catch {
      setDeliveries([]);
    } finally {
      setLoading(false);
    }
  }, [page, filters]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const totalPages = Math.ceil(total / PAGE_SIZE);

  // Separate upcoming from past based on status
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
                      <td>{d.items.length} items</td>
                      <td>
                        <button type="button" className="livraisons-page__edit-btn">
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
                      <td>{d.items.length} items</td>
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
    </div>
  );
}
