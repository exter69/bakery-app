import { useState, useEffect, useCallback } from 'react';
import { fetchMyBakery, fetchBakeryReservations, updateReservationStatus } from '../../api/seller';
import type { Reservation } from '../../api/seller';
import './Dashboard.css';

const PAGE_SIZE = 20;

const STATUS_OPTIONS = [
  { value: '', label: 'All statuses' },
  { value: 'confirmed', label: 'Confirmed' },
  { value: 'ready', label: 'Ready' },
  { value: 'picked_up', label: 'Picked Up' },
  { value: 'cancelled', label: 'Cancelled' },
];

export default function DashboardReservations() {
  const [bakeryId, setBakeryId] = useState<string | null>(null);
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('');
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const loadReservations = useCallback(async (bId: string, p: number, status: string) => {
    try {
      const res = await fetchBakeryReservations(bId, { page: p, status: status || undefined });
      setReservations(res.items);
      setTotal(res.total);
    } catch {
      setMsg({ type: 'error', text: 'Failed to load reservations.' });
    }
  }, []);

  useEffect(() => {
    fetchMyBakery()
      .then((b) => {
        if (b) {
          setBakeryId(b.id);
          return loadReservations(b.id, 1, '');
        }
      })
      .catch(() => setMsg({ type: 'error', text: 'Failed to load bakery.' }))
      .finally(() => setLoading(false));
  }, [loadReservations]);

  useEffect(() => {
    if (bakeryId) {
      loadReservations(bakeryId, page, statusFilter);
    }
  }, [bakeryId, page, statusFilter, loadReservations]);

  const handleStatusChange = async (reservationId: string, newStatus: string) => {
    if (!bakeryId) return;
    try {
      await updateReservationStatus(reservationId, newStatus);
      setMsg({ type: 'success', text: 'Reservation status updated.' });
      await loadReservations(bakeryId, page, statusFilter);
    } catch {
      setMsg({ type: 'error', text: 'Failed to update reservation status.' });
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const formatItems = (reservation: Reservation) =>
    reservation.items.map((i) => `${i.productName} x${i.quantity}`).join(', ');

  const formatTime = (scheduledTime: { startTime: string; endTime: string }) => {
    return `${scheduledTime.startTime} – ${scheduledTime.endTime}`;
  };

  const renderAction = (reservation: Reservation) => {
    switch (reservation.status) {
      case 'confirmed':
        return (
          <button className="dash-btn dash-btn--primary dash-btn--sm" onClick={() => handleStatusChange(reservation.id, 'ready')}>
            Mark Ready
          </button>
        );
      case 'ready':
        return (
          <button className="dash-btn dash-btn--primary dash-btn--sm" onClick={() => handleStatusChange(reservation.id, 'picked_up')}>
            Mark Picked Up
          </button>
        );
      default:
        return <span style={{ color: '#94a3b8', fontSize: '0.8rem' }}>—</span>;
    }
  };

  if (loading) return <div className="dash-loading">Loading reservations…</div>;

  if (!bakeryId) {
    return (
      <div className="dash-empty">
        <h1 className="dash-page__title">Reservations</h1>
        <p style={{ marginTop: '1rem' }}>No bakery found.</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="dash-page__title">Reservations</h1>
      <p className="dash-page__subtitle">View and manage incoming reservations.</p>

      {msg && <div className={`dash-msg dash-msg--${msg.type}`}>{msg.text}</div>}

      <div className="dash-toolbar">
        <select
          className="dash-form__select"
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          
        >
          {STATUS_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
      </div>

      {reservations.length === 0 ? (
        <div className="dash-empty">No reservations found.</div>
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
                {reservations.map((reservation) => (
                  <tr key={reservation.id}>
                    <td style={{ maxWidth: 250, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {formatItems(reservation)}
                    </td>
                    <td>{formatTime(reservation.scheduledTime)}</td>
                    <td>
                      <span className={`dash-badge dash-badge--${reservation.status}`}>
                        {reservation.status.replace('_', ' ')}
                      </span>
                    </td>
                    <td>{renderAction(reservation)}</td>
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
