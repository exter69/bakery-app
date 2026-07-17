import { useState, useEffect, useCallback } from 'react';
import { fetchMyBakery, fetchBakeryOrders, updateOrderStatus } from '../../api/seller';
import type { Order } from '../../api/seller';
import './Dashboard.css';

const PAGE_SIZE = 20;

const STATUS_OPTIONS = [
  { value: '', label: 'All statuses' },
  { value: 'confirmed', label: 'Confirmed' },
  { value: 'preparing', label: 'Preparing' },
  { value: 'ready', label: 'Ready' },
  { value: 'delivered', label: 'Delivered' },
  { value: 'cancelled', label: 'Cancelled' },
];

export default function DashboardOrders() {
  const [bakeryId, setBakeryId] = useState<string | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('');
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const loadOrders = useCallback(async (bId: string, p: number, status: string) => {
    try {
      const res = await fetchBakeryOrders(bId, { page: p, status: status || undefined });
      setOrders(res.items);
      setTotal(res.total);
    } catch {
      setMsg({ type: 'error', text: 'Failed to load orders.' });
    }
  }, []);

  useEffect(() => {
    fetchMyBakery()
      .then((b) => {
        if (b) {
          setBakeryId(b.id);
          return loadOrders(b.id, 1, '');
        }
      })
      .catch(() => setMsg({ type: 'error', text: 'Failed to load bakery.' }))
      .finally(() => setLoading(false));
  }, [loadOrders]);

  useEffect(() => {
    if (bakeryId) {
      loadOrders(bakeryId, page, statusFilter);
    }
  }, [bakeryId, page, statusFilter, loadOrders]);

  const handleStatusChange = async (orderId: string, newStatus: string) => {
    if (!bakeryId) return;
    try {
      await updateOrderStatus(orderId, newStatus);
      setMsg({ type: 'success', text: 'Order status updated.' });
      await loadOrders(bakeryId, page, statusFilter);
    } catch {
      setMsg({ type: 'error', text: 'Failed to update order status.' });
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const formatItems = (order: Order) =>
    order.items.map((i) => `${i.productName} x${i.quantity}`).join(', ');

  const formatTime = (iso: string) => {
    const d = new Date(iso);
    return d.toLocaleString(undefined, { dateStyle: 'short', timeStyle: 'short' });
  };

  const renderAction = (order: Order) => {
    switch (order.status) {
      case 'confirmed':
        return (
          <button className="dash-btn dash-btn--primary dash-btn--sm" onClick={() => handleStatusChange(order.id, 'preparing')}>
            Start Preparing
          </button>
        );
      case 'preparing':
        return (
          <button className="dash-btn dash-btn--primary dash-btn--sm" onClick={() => handleStatusChange(order.id, 'ready')}>
            Mark Ready
          </button>
        );
      case 'ready':
        return (
          <button className="dash-btn dash-btn--primary dash-btn--sm" onClick={() => handleStatusChange(order.id, 'delivered')}>
            Mark Delivered
          </button>
        );
      default:
        return <span style={{ color: '#94a3b8', fontSize: '0.8rem' }}>—</span>;
    }
  };

  if (loading) return <div className="dash-loading">Loading orders…</div>;

  if (!bakeryId) {
    return (
      <div className="dash-empty">
        <h1 className="dash-page__title">Orders</h1>
        <p style={{ marginTop: '1rem' }}>No bakery found.</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="dash-page__title">Orders</h1>
      <p className="dash-page__subtitle">View and manage incoming orders.</p>

      {msg && <div className={`dash-msg dash-msg--${msg.type}`}>{msg.text}</div>}

      <div className="dash-toolbar">
        <select
          className="dash-form__select"
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          style={{ maxWidth: 200 }}
        >
          {STATUS_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
      </div>

      {orders.length === 0 ? (
        <div className="dash-empty">No orders found.</div>
      ) : (
        <div className="dash-card" style={{ padding: 0, overflow: 'hidden' }}>
          <div className="dash-table-wrap">
            <table className="dash-table">
              <thead>
                <tr>
                  <th>Items</th>
                  <th>Scheduled</th>
                  <th>Status</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {orders.map((order) => (
                  <tr key={order.id}>
                    <td style={{ maxWidth: 250, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {formatItems(order)}
                    </td>
                    <td>{formatTime(order.scheduledTime)}</td>
                    <td>
                      <span className={`dash-badge dash-badge--${order.status}`}>
                        {order.status.replace('_', ' ')}
                      </span>
                    </td>
                    <td>{renderAction(order)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {totalPages > 1 && (
        <div className="dash-pagination">
          <button
            className="dash-btn dash-btn--secondary dash-btn--sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            Previous
          </button>
          <span>Page {page} of {totalPages}</span>
          <button
            className="dash-btn dash-btn--secondary dash-btn--sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}
