import { useCallback, useEffect, useState } from 'react';
import { fetchScheduleEntries, deleteOrder, deleteReservation } from '../api/orders';
import { ApiError } from '../api/client';
import { useWebSocket } from '../hooks/useWebSocket';
import { PushNotificationToggle } from '../components/PushNotificationToggle';
import type { ScheduleEntry, ScheduleQueryParams } from '../types/order';
import './ScheduleOrdersPage.css';

const PAGE_SIZE = 20;

const STATUS_OPTIONS = [
  { value: '', label: 'All statuses' },
  { value: 'pending_payment', label: 'Pending Payment' },
  { value: 'confirmed', label: 'Confirmed' },
  { value: 'preparing', label: 'Preparing' },
  { value: 'ready', label: 'Ready' },
  { value: 'delivered', label: 'Delivered' },
  { value: 'picked_up', label: 'Picked Up' },
  { value: 'cancelled', label: 'Cancelled' },
];

const SORT_FIELD_OPTIONS = [
  { value: 'scheduledTime', label: 'Scheduled Time' },
  { value: 'createdAt', label: 'Creation Date' },
];

const SORT_DIRECTION_OPTIONS = [
  { value: 'desc', label: 'Newest first' },
  { value: 'asc', label: 'Oldest first' },
];

function formatStatus(status: string): string {
  return status.replace(/_/g, ' ');
}

function formatDateTime(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
}

export default function ScheduleOrdersPage() {
  const [entries, setEntries] = useState<ScheduleEntry[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters & sorting
  const [statusFilter, setStatusFilter] = useState('');
  const [sortBy, setSortBy] = useState<'scheduledTime' | 'createdAt'>('scheduledTime');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');

  // Delete state
  const [deleteTarget, setDeleteTarget] = useState<ScheduleEntry | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [toast, setToast] = useState<{ message: string; isError: boolean } | null>(null);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  // Real-time updates via WebSocket
  const { lastEvent } = useWebSocket();

  const loadEntries = useCallback(async () => {
    setLoading(true);
    setError(null);

    const params: ScheduleQueryParams = {
      page,
      sortBy,
      sortDirection,
    };
    if (statusFilter) {
      params.status = statusFilter;
    }

    try {
      const response = await fetchScheduleEntries(params);
      setEntries(response.items);
      setTotal(response.total);
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Failed to load orders';
      setError(message);
    } finally {
      setLoading(false);
    }
  }, [page, statusFilter, sortBy, sortDirection]);

  useEffect(() => {
    loadEntries();
  }, [loadEntries]);

  // Reset page when filters change
  useEffect(() => {
    setPage(1);
  }, [statusFilter, sortBy, sortDirection]);

  // Refresh list when a real-time order/reservation event arrives
  useEffect(() => {
    if (lastEvent?.type === 'order_status' || lastEvent?.type === 'reservation_status') {
      loadEntries();
    }
  }, [lastEvent, loadEntries]);

  function showToast(message: string, isError = false) {
    setToast({ message, isError });
    setTimeout(() => setToast(null), 4000);
  }

  function handleDeleteClick(entry: ScheduleEntry) {
    setDeleteTarget(entry);
  }

  function handleCancelDelete() {
    setDeleteTarget(null);
  }

  async function handleConfirmDelete() {
    if (!deleteTarget) return;

    setDeleting(true);
    try {
      if (deleteTarget.type === 'order') {
        await deleteOrder(deleteTarget.id);
      } else {
        await deleteReservation(deleteTarget.id);
      }
      setDeleteTarget(null);
      showToast('Successfully cancelled');
      loadEntries();
    } catch (err) {
      setDeleteTarget(null);
      if (err instanceof ApiError) {
        if (err.status === 403) {
          showToast('You do not have permission to delete this item', true);
        } else if (err.status === 404) {
          showToast('This item was not found', true);
        } else {
          showToast(err.message || 'Cannot delete this item', true);
        }
      } else {
        showToast('Failed to delete. Please try again.', true);
      }
    } finally {
      setDeleting(false);
    }
  }

  return (
    <div className="schedule-orders">
      <h1 className="schedule-orders__title">My Orders &amp; Reservations</h1>

      {/* Push notification opt-in */}
      <PushNotificationToggle />

      {/* Filter & Sort Controls */}
      <div className="schedule-orders__controls">
        <div className="schedule-orders__control-group">
          <label htmlFor="status-filter">Status</label>
          <select
            id="status-filter"
            className="schedule-orders__select"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div className="schedule-orders__control-group">
          <label htmlFor="sort-field">Sort by</label>
          <select
            id="sort-field"
            className="schedule-orders__select"
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'scheduledTime' | 'createdAt')}
          >
            {SORT_FIELD_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div className="schedule-orders__control-group">
          <label htmlFor="sort-direction">Direction</label>
          <select
            id="sort-direction"
            className="schedule-orders__select"
            value={sortDirection}
            onChange={(e) => setSortDirection(e.target.value as 'asc' | 'desc')}
          >
            {SORT_DIRECTION_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Loading */}
      {loading && (
        <div className="schedule-orders__loading">
          <div className="spinner" />
          <p>Loading orders…</p>
        </div>
      )}

      {/* Error */}
      {!loading && error && (
        <div className="schedule-orders__error">
          <p>{error}</p>
          <button className="btn btn--outline" onClick={loadEntries}>
            Retry
          </button>
        </div>
      )}

      {/* Empty state */}
      {!loading && !error && entries.length === 0 && (
        <div className="schedule-orders__empty">
          <p>No orders or reservations found.</p>
        </div>
      )}

      {/* Table */}
      {!loading && !error && entries.length > 0 && (
        <>
          <div className="schedule-orders__table-wrapper">
            <table className="schedule-orders__table">
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Items</th>
                  <th>Scheduled Time</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((entry) => (
                  <tr key={entry.id}>
                    <td>
                      <span
                        className={`schedule-orders__type-badge schedule-orders__type-badge--${entry.type}`}
                      >
                        {entry.type === 'order' ? 'Order' : 'Reservation'}
                      </span>
                    </td>
                    <td>{entry.items.map((item) => item.productName).join(', ')}</td>
                    <td>{formatDateTime(entry.scheduledTime)}</td>
                    <td>
                      <span
                        className={`schedule-orders__status-badge schedule-orders__status-badge--${entry.status}`}
                      >
                        {formatStatus(entry.status)}
                      </span>
                    </td>
                    <td>
                      <button
                        className="schedule-orders__delete-btn"
                        onClick={() => handleDeleteClick(entry)}
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="schedule-orders__pagination">
            <button
              className="schedule-orders__page-btn"
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
            >
              Previous
            </button>
            <span className="schedule-orders__page-info">
              Page {page} of {totalPages}
            </span>
            <button
              className="schedule-orders__page-btn"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
            </button>
          </div>
        </>
      )}

      {/* Delete confirmation dialog */}
      {deleteTarget && (
        <div className="schedule-orders__confirm-overlay" onClick={handleCancelDelete}>
          <div
            className="schedule-orders__confirm-dialog"
            onClick={(e) => e.stopPropagation()}
            role="dialog"
            aria-modal="true"
            aria-labelledby="confirm-delete-title"
          >
            <h3 id="confirm-delete-title">Confirm Deletion</h3>
            <p>
              Are you sure you want to delete this {deleteTarget.type}? This action cannot be undone.
            </p>
            <div className="schedule-orders__confirm-actions">
              <button
                className="schedule-orders__confirm-cancel"
                onClick={handleCancelDelete}
                disabled={deleting}
              >
                Cancel
              </button>
              <button
                className="schedule-orders__confirm-delete"
                onClick={handleConfirmDelete}
                disabled={deleting}
              >
                {deleting ? 'Deleting…' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Toast notification */}
      {toast && (
        <div
          className={`schedule-orders__toast ${toast.isError ? 'schedule-orders__toast--error' : ''}`}
          role="alert"
        >
          {toast.message}
        </div>
      )}
    </div>
  );
}
