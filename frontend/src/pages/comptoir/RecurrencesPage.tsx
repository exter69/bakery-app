import { useState, useEffect, useCallback } from 'react';
import { useI18n } from '../../i18n';
import {
  fetchRecurringOrders,
  createRecurringOrder,
  pauseRecurringOrder,
  resumeRecurringOrder,
  deleteRecurringOrder,
  type RecurringOrder,
  type CreateRecurringOrderPayload,
} from '../../api/recurring';
import { listApprovedBakeries, getProducts } from '../../api/b2b-client';
import { ErrorState } from '../../components/ErrorState';
import type { Bakery, Product } from '../../types/bakery';

const DAYS_OF_WEEK = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];

export default function RecurrencesPage() {
  const { t } = useI18n();
  const [orders, setOrders] = useState<RecurringOrder[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchRecurringOrders(page);
      setOrders(result.items);
      setTotal(result.total);
    } catch {
      setError(t('comptoir.common.loadError'));
    } finally {
      setLoading(false);
    }
  }, [page, t]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const totalPages = Math.ceil(total / 20);

  // Pause/resume toggle with optimistic update
  const handleToggle = async (order: RecurringOrder) => {
    const prev = orders;
    setOrders((o) => o.map((r) => r.id === order.id ? { ...r, active: !r.active } : r));
    try {
      if (order.active) {
        await pauseRecurringOrder(order.id);
      } else {
        await resumeRecurringOrder(order.id);
      }
    } catch {
      setOrders(prev);
    }
  };

  // Delete with confirmation
  const handleDelete = async (id: string) => {
    try {
      await deleteRecurringOrder(id);
      setOrders((o) => o.filter((r) => r.id !== id));
      setDeleteConfirmId(null);
    } catch {
      setError(t('comptoir.error.generic'));
      setDeleteConfirmId(null);
    }
  };

  return (
    <div className="recurrences-page">
      <div className="recurrences-page__header">
        <h1>{t('comptoir.recurrences.title')}</h1>
        <button
          type="button"
          className="recurrences-page__new-btn"
          onClick={() => setShowForm(!showForm)}
        >
          {t('comptoir.recurrences.new')}
        </button>
      </div>

      {showForm && (
        <CreationForm
          onCreated={(order) => {
            setOrders((prev) => [order, ...prev]);
            setShowForm(false);
          }}
          onCancel={() => setShowForm(false)}
        />
      )}

      {loading ? (
        <p>{t('comptoir.common.loading')}</p>
      ) : error ? (
        <ErrorState message={error} onRetry={fetchData} retryLabel={t('comptoir.common.retry')} />
      ) : orders.length === 0 ? (
        <p className="recurrences-page__empty">{t('comptoir.recurrences.empty')}</p>
      ) : (
        <>
          <table className="recurrences-page__table">
            <thead>
              <tr>
                <th>{t('comptoir.invoices.bakery')}</th>
                <th>{t('comptoir.recurrences.frequency')}</th>
                <th>{t('comptoir.commander.quantity')}</th>
                <th>{t('comptoir.pricing.totalTtc')}</th>
                <th>{t('comptoir.invoices.status')}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => (
                <tr key={order.id}>
                  <td>{order.bakeryName}</td>
                  <td>{order.frequency}</td>
                  <td>{order.items.length}</td>
                  <td>{(order.items.reduce((sum, i) => sum + i.subtotal, 0) / 100).toFixed(2)} EUR</td>
                  <td>
                    <span className={`recurrences-page__badge ${order.active ? 'recurrences-page__badge--active' : ''}`}>
                      {order.active ? t('comptoir.recurrences.active') : t('comptoir.recurrences.inactive')}
                    </span>
                  </td>
                  <td className="recurrences-page__actions">
                    <button type="button" onClick={() => handleToggle(order)}>
                      {order.active ? t('comptoir.recurrences.deactivate') : t('comptoir.recurrences.activate')}
                    </button>
                    {deleteConfirmId === order.id ? (
                      <>
                        <button type="button" onClick={() => handleDelete(order.id)}>
                          {t('comptoir.common.confirm')}
                        </button>
                        <button type="button" onClick={() => setDeleteConfirmId(null)}>
                          {t('comptoir.common.cancel')}
                        </button>
                      </>
                    ) : (
                      <button type="button" onClick={() => setDeleteConfirmId(order.id)}>
                        {t('comptoir.common.delete')}
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {totalPages > 1 && (
            <div className="recurrences-page__pagination">
              <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                {t('comptoir.common.previous')}
              </button>
              <span>{page} / {totalPages}</span>
              <button type="button" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
                {t('comptoir.common.next')}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

// --- Creation Form ---

interface CreationFormProps {
  onCreated: (order: RecurringOrder) => void;
  onCancel: () => void;
}

function CreationForm({ onCreated, onCancel }: CreationFormProps) {
  const { t } = useI18n();
  const [bakeries, setBakeries] = useState<Bakery[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const [bakeryId, setBakeryId] = useState('');
  const [items, setItems] = useState<{ productId: string; quantity: number }[]>([]);
  const [frequency, setFrequency] = useState('weekly');
  const [scheduledDay, setScheduledDay] = useState('monday');
  const [startTime, setStartTime] = useState('08:00');
  const [endTime, setEndTime] = useState('10:00');
  const [selectionMode, setSelectionMode] = useState('fixed');

  useEffect(() => {
    listApprovedBakeries().then(setBakeries).catch(() => setBakeries([]));
  }, []);

  useEffect(() => {
    if (bakeryId) {
      getProducts(bakeryId)
        .then((categorized) => setProducts(Object.values(categorized).flat()))
        .catch(() => setProducts([]));
    } else {
      setProducts([]);
    }
  }, [bakeryId]);

  const handleAddItem = (productId: string) => {
    if (!productId || items.some((i) => i.productId === productId)) return;
    setItems((prev) => [...prev, { productId, quantity: 1 }]);
  };

  const handleRemoveItem = (productId: string) => {
    setItems((prev) => prev.filter((i) => i.productId !== productId));
  };

  const handleItemQuantity = (productId: string, quantity: number) => {
    setItems((prev) =>
      prev.map((i) => i.productId === productId ? { ...i, quantity: Math.max(1, quantity) } : i)
    );
  };

  const handleSubmit = async () => {
    setFormError(null);
    if (!bakeryId || items.length === 0) {
      setFormError(t('comptoir.recurrences.form.validationError'));
      return;
    }
    setSubmitting(true);
    try {
      const payload: CreateRecurringOrderPayload = {
        bakeryId,
        items: items.map((i) => ({ productId: i.productId, quantity: i.quantity })),
        scheduledDay,
        scheduledTime: { startTime, endTime },
        frequency,
        selectionMode,
      };
      const created = await createRecurringOrder(payload);
      onCreated(created);
    } catch {
      setFormError(t('comptoir.error.generic'));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="recurrences-page__form">
      {formError && <p className="recurrences-page__form-error" role="alert">{formError}</p>}

      {/* Bakery picker */}
      <label className="recurrences-page__field">
        <span>{t('comptoir.recurrences.form.bakery')}</span>
        <select value={bakeryId} onChange={(e) => setBakeryId(e.target.value)}>
          <option value="">{t('comptoir.commander.selectBakery')}</option>
          {bakeries.map((b) => (
            <option key={b.id} value={b.id}>{b.name}</option>
          ))}
        </select>
      </label>

      {/* Product picker */}
      {bakeryId && products.length > 0 && (
        <div className="recurrences-page__field">
          <span>{t('comptoir.recurrences.form.products')}</span>
          <select onChange={(e) => handleAddItem(e.target.value)} value="">
            <option value="">{t('comptoir.recurrences.form.addProduct')}</option>
            {products
              .filter((p) => p.isAvailable && !items.some((i) => i.productId === p.id))
              .map((p) => (
                <option key={p.id} value={p.id}>{p.name} - {(p.price / 100).toFixed(2)} EUR</option>
              ))}
          </select>
          {items.length > 0 && (
            <table className="recurrences-page__items-table">
              <thead>
                <tr>
                  <th>{t('comptoir.commander.product')}</th>
                  <th>{t('comptoir.commander.quantity')}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => {
                  const prod = products.find((p) => p.id === item.productId);
                  return (
                    <tr key={item.productId}>
                      <td>{prod?.name ?? item.productId}</td>
                      <td>
                        <input
                          type="number"
                          min="1"
                          value={item.quantity}
                          onChange={(e) => handleItemQuantity(item.productId, Number(e.target.value))}
                        />
                      </td>
                      <td>
                        <button type="button" onClick={() => handleRemoveItem(item.productId)}>
                          {t('comptoir.common.delete')}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Frequency */}
      <label className="recurrences-page__field">
        <span>{t('comptoir.recurrences.frequency')}</span>
        <select value={frequency} onChange={(e) => setFrequency(e.target.value)}>
          <option value="weekly">{t('comptoir.recurrences.weekly')}</option>
          <option value="biweekly">{t('comptoir.recurrences.form.biweekly')}</option>
        </select>
      </label>

      {/* Scheduled day */}
      <label className="recurrences-page__field">
        <span>{t('comptoir.recurrences.form.day')}</span>
        <select value={scheduledDay} onChange={(e) => setScheduledDay(e.target.value)}>
          {DAYS_OF_WEEK.map((day) => (
            <option key={day} value={day}>{t(`comptoir.recurrences.form.days.${day}`)}</option>
          ))}
        </select>
      </label>

      {/* Scheduled time */}
      <div className="recurrences-page__field recurrences-page__field--row">
        <label>
          <span>{t('comptoir.recurrences.form.startTime')}</span>
          <input type="time" value={startTime} onChange={(e) => setStartTime(e.target.value)} />
        </label>
        <label>
          <span>{t('comptoir.recurrences.form.endTime')}</span>
          <input type="time" value={endTime} onChange={(e) => setEndTime(e.target.value)} />
        </label>
      </div>

      {/* Selection mode */}
      <fieldset className="recurrences-page__field">
        <legend>{t('comptoir.recurrences.form.selectionMode')}</legend>
        <label>
          <input type="radio" name="selectionMode" value="fixed" checked={selectionMode === 'fixed'} onChange={() => setSelectionMode('fixed')} />
          {t('comptoir.recurrences.form.modeFixed')}
        </label>
        <label>
          <input type="radio" name="selectionMode" value="bakeryChoice" checked={selectionMode === 'bakeryChoice'} onChange={() => setSelectionMode('bakeryChoice')} />
          {t('comptoir.recurrences.form.modeBakeryChoice')}
        </label>
        <label>
          <input type="radio" name="selectionMode" value="randomFavorites" checked={selectionMode === 'randomFavorites'} onChange={() => setSelectionMode('randomFavorites')} />
          {t('comptoir.recurrences.form.modeRandomFavorites')}
        </label>
      </fieldset>

      <div className="recurrences-page__form-actions">
        <button type="button" onClick={onCancel}>
          {t('comptoir.common.cancel')}
        </button>
        <button type="button" onClick={handleSubmit} disabled={submitting}>
          {t('comptoir.common.save')}
        </button>
      </div>
    </div>
  );
}
