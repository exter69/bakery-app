import { useState, useEffect, useCallback } from 'react';
import {
  fetchRecurringOrders,
  pauseRecurringOrder,
  resumeRecurringOrder,
  deleteRecurringOrder,
  fetchProfile,
  updateHoliday,
} from '../api/recurring';
import type { RecurringOrder, UserProfile } from '../api/recurring';
import './RecurringOrdersPage.css';

const PAGE_SIZE = 20;

export default function RecurringOrdersPage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [holidayMode, setHolidayMode] = useState(false);
  const [holidayFrom, setHolidayFrom] = useState('');
  const [holidayTo, setHolidayTo] = useState('');
  const [holidaySaving, setHolidaySaving] = useState(false);

  const [orders, setOrders] = useState<RecurringOrder[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const loadOrders = useCallback(async (p: number) => {
    try {
      const res = await fetchRecurringOrders(p);
      setOrders(res.items);
      setTotal(res.total);
    } catch {
      setMsg({ type: 'error', text: 'Failed to load recurring orders.' });
    }
  }, []);

  useEffect(() => {
    async function init() {
      try {
        const [prof] = await Promise.all([fetchProfile(), loadOrders(1)]);
        setProfile(prof);
        setHolidayMode(prof.holidayMode);
        setHolidayFrom(prof.holidayFrom ?? '');
        setHolidayTo(prof.holidayTo ?? '');
      } catch {
        setMsg({ type: 'error', text: 'Failed to load profile.' });
      } finally {
        setLoading(false);
      }
    }
    init();
  }, [loadOrders]);

  useEffect(() => {
    if (!loading) {
      loadOrders(page);
    }
  }, [page, loading, loadOrders]);

  async function handleHolidaySave() {
    setHolidaySaving(true);
    setMsg(null);
    try {
      const updated = await updateHoliday({
        holidayMode,
        holidayFrom: holidayMode ? holidayFrom || undefined : undefined,
        holidayTo: holidayMode ? holidayTo || undefined : undefined,
      });
      setProfile(updated);
      setMsg({ type: 'success', text: 'Holiday settings saved.' });
    } catch {
      setMsg({ type: 'error', text: 'Failed to save holiday settings.' });
    } finally {
      setHolidaySaving(false);
    }
  }

  async function handlePause(id: string) {
    try {
      await pauseRecurringOrder(id);
      await loadOrders(page);
      setMsg({ type: 'success', text: 'Order paused.' });
    } catch {
      setMsg({ type: 'error', text: 'Failed to pause order.' });
    }
  }

  async function handleResume(id: string) {
    try {
      await resumeRecurringOrder(id);
      await loadOrders(page);
      setMsg({ type: 'success', text: 'Order resumed.' });
    } catch {
      setMsg({ type: 'error', text: 'Failed to resume order.' });
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteRecurringOrder(id);
      await loadOrders(page);
      setMsg({ type: 'success', text: 'Order deleted.' });
    } catch {
      setMsg({ type: 'error', text: 'Failed to delete order.' });
    }
  }

  const totalPages = Math.ceil(total / PAGE_SIZE);

  if (loading) return <div className="recurring-loading">Loading…</div>;

  return (
    <div className="recurring-page">
      <h1 className="recurring-page__title">Recurring Orders</h1>
      <p className="recurring-page__subtitle">Manage your recurring orders and holiday mode.</p>

      {msg && <div className={`recurring-msg recurring-msg--${msg.type}`}>{msg.text}</div>}

      {/* Holiday Mode Card */}
      <div className="holiday-card">
        <div className="holiday-card__header">
          <h2 className="holiday-card__title">Holiday Mode</h2>
          <label className="recurring-toggle">
            <input
              type="checkbox"
              checked={holidayMode}
              onChange={(e) => setHolidayMode(e.target.checked)}
            />
            <span className="recurring-toggle__slider" />
          </label>
        </div>

        {profile?.holidayMode && (
          <div className="holiday-banner">
            Holiday mode active — recurring orders paused
            {profile.holidayFrom && profile.holidayTo && (
              <> ({profile.holidayFrom} to {profile.holidayTo})</>
            )}
          </div>
        )}

        {holidayMode && (
          <div className="holiday-card__dates">
            <div className="holiday-card__field">
              <label className="holiday-card__label">From</label>
              <input
                type="date"
                className="holiday-card__input"
                value={holidayFrom}
                onChange={(e) => setHolidayFrom(e.target.value)}
              />
            </div>
            <div className="holiday-card__field">
              <label className="holiday-card__label">To</label>
              <input
                type="date"
                className="holiday-card__input"
                value={holidayTo}
                onChange={(e) => setHolidayTo(e.target.value)}
              />
            </div>
          </div>
        )}

        <div className="holiday-card__actions">
          <button
            className="recurring-btn recurring-btn--primary"
            onClick={handleHolidaySave}
            disabled={holidaySaving}
          >
            {holidaySaving ? 'Saving…' : 'Save'}
          </button>
        </div>
      </div>

      {/* Recurring Orders Table */}
      <div className="recurring-orders-section">
        <h2 className="recurring-orders-section__title">My Recurring Orders</h2>

        {orders.length === 0 ? (
          <div className="recurring-empty">
            No recurring orders yet. Set one up when placing an order!
          </div>
        ) : (
          <>
            <div className="recurring-table-wrap">
              <table className="recurring-table">
                <thead>
                  <tr>
                    <th>Bakery</th>
                    <th>Items</th>
                    <th>Day</th>
                    <th>Time</th>
                    <th>Frequency</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {orders.map((order) => (
                    <tr key={order.id}>
                      <td>{order.bakeryName}</td>
                      <td style={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {order.items.map((i) => i.productName).join(', ')}
                      </td>
                      <td>{order.scheduledDay}</td>
                      <td>{order.scheduledTime.startTime}–{order.scheduledTime.endTime}</td>
                      <td style={{ textTransform: 'capitalize' }}>{order.frequency}</td>
                      <td>
                        <span className={`recurring-badge recurring-badge--${order.active ? 'active' : 'paused'}`}>
                          {order.active ? 'Active' : 'Paused'}
                        </span>
                      </td>
                      <td>
                        <div className="recurring-actions">
                          {order.active ? (
                            <button
                              className="recurring-btn recurring-btn--secondary"
                              onClick={() => handlePause(order.id)}
                            >
                              Pause
                            </button>
                          ) : (
                            <button
                              className="recurring-btn recurring-btn--primary"
                              onClick={() => handleResume(order.id)}
                            >
                              Resume
                            </button>
                          )}
                          <button
                            className="recurring-btn recurring-btn--danger"
                            onClick={() => handleDelete(order.id)}
                          >
                            Delete
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {totalPages > 1 && (
              <div className="recurring-pagination">
                <button
                  className="recurring-btn recurring-btn--secondary"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  Previous
                </button>
                <span>Page {page} of {totalPages}</span>
                <button
                  className="recurring-btn recurring-btn--secondary"
                  disabled={page >= totalPages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
