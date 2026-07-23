import { useState, useEffect, useCallback, useRef } from 'react';
import { fetchMyBakery, fetchBakeryOrders, updateOrderStatus } from '../../api/seller';
import { useWebSocket } from '../../hooks/useWebSocket';
import { useI18n } from '../../i18n';
import { FilterChips } from '../../components/pro/FilterChips';
import { OrderCard } from '../../components/pro/OrderCard';
import { ErrorBanner } from '../../components/pro/ErrorBanner';
import {
  groupOrdersByStatus,
  isAdjacentTransition,
  COLUMN_ORDER,
  type KanbanStatus,
} from './kanban-utils';
import type { Order } from '../../api/seller';
import type { OrderStatus } from '../../types/order';
import './DashboardOrders.css';

/** Filter chip options */
type DeliveryFilter = 'all' | 'livraison' | 'retrait';

/** Format today's date using locale day name */
function formatDayName(date: Date, locale: string): string {
  return date.toLocaleDateString(locale, { weekday: 'long' });
}

/** Format time from TimeSlotResponse object as HH:MM */
function formatTime(scheduledTime: Order['scheduledTime']): string {
  if (typeof scheduledTime === 'object' && scheduledTime !== null && 'startTime' in scheduledTime) {
    return scheduledTime.startTime;
  }
  return '--:--';
}

/** Format item summary for order card */
function formatItems(order: Order): string {
  return order.items.map((i) => `${i.quantity}× ${i.productName}`).join(', ');
}

/** Map order type to card delivery type */
function getDeliveryType(order: Order): 'livraison' | 'retrait' {
  return order.type === 'reservation' ? 'retrait' : 'livraison';
}

export default function DashboardOrders() {
  const { t } = useI18n();
  const [bakeryId, setBakeryId] = useState<string | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<DeliveryFilter>('all');
  const [selectedDate, setSelectedDate] = useState<string>(
    new Date().toISOString().slice(0, 10)
  );

  // Drag-and-drop state
  const [draggedOrderId, setDraggedOrderId] = useState<string | null>(null);
  const [dragOverColumn, setDragOverColumn] = useState<KanbanStatus | null>(null);
  const dragCounterRef = useRef<Map<KanbanStatus, number>>(new Map());

  // Toast state
  const [toast, setToast] = useState<{ message: string; isError: boolean } | null>(null);
  const toastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Real-time updates via WebSocket
  const { lastEvent } = useWebSocket();

  /** Column display labels from translations */
  const COLUMN_LABELS: Record<KanbanStatus, string> = {
    confirmed: t('dashboard.orders.column.confirmed'),
    preparing: t('dashboard.orders.column.preparing'),
    ready: t('dashboard.orders.column.ready'),
    delivered: t('dashboard.orders.column.delivered'),
  };

  const FILTER_OPTIONS: { value: DeliveryFilter; label: string }[] = [
    { value: 'livraison', label: t('dashboard.orders.filter.delivery') },
    { value: 'retrait', label: t('dashboard.orders.filter.pickup') },
    { value: 'all', label: t('dashboard.orders.filter.all') },
  ];

  function showToast(message: string, isError = false) {
    if (toastTimerRef.current) clearTimeout(toastTimerRef.current);
    setToast({ message, isError });
    toastTimerRef.current = setTimeout(() => setToast(null), 4000);
  }

  const loadOrders = useCallback(async (bId: string) => {
    try {
      setError(null);
      const res = await fetchBakeryOrders(bId);
      setOrders(res.items);
    } catch {
      setError(t('dashboard.orders.loadError'));
    }
  }, [t]);

  useEffect(() => {
    fetchMyBakery()
      .then((b) => {
        if (b) {
          setBakeryId(b.id);
          return loadOrders(b.id);
        }
      })
      .catch(() => setError(t('dashboard.orders.bakeryError')))
      .finally(() => setLoading(false));
  }, [loadOrders, t]);

  // Refresh when a new order arrives via WebSocket
  useEffect(() => {
    if (lastEvent?.type === 'new_order' && bakeryId) {
      loadOrders(bakeryId);
    }
  }, [lastEvent, bakeryId, loadOrders]);

  // --- Status update handler (from action buttons) ---
  const handleStatusChange = useCallback(
    async (orderId: string, newStatus: OrderStatus) => {
      if (!bakeryId) return;
      try {
        await updateOrderStatus(orderId, newStatus);
        await loadOrders(bakeryId);
      } catch {
        showToast(t('dashboard.orders.statusError'), true);
      }
    },
    [bakeryId, loadOrders, t]
  );

  // --- Drag-and-drop handlers ---
  function handleDragStart(e: React.DragEvent, orderId: string) {
    setDraggedOrderId(orderId);
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', orderId);
  }

  function handleDragEnd() {
    setDraggedOrderId(null);
    setDragOverColumn(null);
    dragCounterRef.current.clear();
  }

  function handleDragEnter(column: KanbanStatus) {
    const counter = (dragCounterRef.current.get(column) ?? 0) + 1;
    dragCounterRef.current.set(column, counter);
    setDragOverColumn(column);
  }

  function handleDragLeave(column: KanbanStatus) {
    const counter = (dragCounterRef.current.get(column) ?? 0) - 1;
    dragCounterRef.current.set(column, counter);
    if (counter <= 0) {
      dragCounterRef.current.set(column, 0);
      if (dragOverColumn === column) {
        setDragOverColumn(null);
      }
    }
  }

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  }

  async function handleDrop(e: React.DragEvent, targetColumn: KanbanStatus) {
    e.preventDefault();
    setDragOverColumn(null);
    dragCounterRef.current.clear();

    const orderId = e.dataTransfer.getData('text/plain') || draggedOrderId;
    setDraggedOrderId(null);

    if (!orderId || !bakeryId) return;

    // Find the source column for this order
    const order = orders.find((o) => o.id === orderId);
    if (!order) return;

    const sourceColumn = order.status as KanbanStatus;
    if (sourceColumn === targetColumn) return;

    // Validate adjacent transition
    if (!isAdjacentTransition(sourceColumn, targetColumn)) {
      showToast(t('dashboard.orders.invalidTransition'), true);
      return;
    }

    // Valid transition — update via API
    try {
      await updateOrderStatus(orderId, targetColumn);
      await loadOrders(bakeryId);
    } catch {
      showToast(t('dashboard.orders.dropFailed'), true);
    }
  }

  // --- Filtering ---
  const filteredOrders = orders.filter((order) => {
    if (filter === 'all') return true;
    if (filter === 'livraison') return order.type === 'order';
    if (filter === 'retrait') return order.type === 'reservation';
    return true;
  });

  const columns = groupOrdersByStatus(filteredOrders);

  // --- Render ---
  if (loading) {
    return <div className="kanban__loading">{t('dashboard.orders.loading')}</div>;
  }

  if (error && orders.length === 0) {
    return (
      <div className="kanban__error">
        <ErrorBanner
          message={error}
          onRetry={bakeryId ? () => loadOrders(bakeryId) : undefined}
        />
      </div>
    );
  }

  const dayLabel = formatDayName(new Date(selectedDate + 'T00:00:00'), 'fr-FR');

  return (
    <div className="kanban">
      {/* Header */}
      <div className="kanban__header">
        <h1 className="kanban__title">{t('dashboard.orders.title')} — {dayLabel}</h1>
        <input
          type="date"
          className="kanban__date-picker"
          value={selectedDate}
          onChange={(e) => setSelectedDate(e.target.value)}
          aria-label={t('dashboard.orders.selectDate')}
        />
      </div>

      {/* Filter chips */}
      <FilterChips
        options={FILTER_OPTIONS}
        selected={filter}
        onChange={setFilter}
      />

      {/* Kanban board */}
      <div className="kanban__board" role="region" aria-label={t('dashboard.orders.title')}>
        {COLUMN_ORDER.map((status) => {
          const columnOrders = columns.get(status) ?? [];
          const isDragOver = dragOverColumn === status;

          return (
            <div
              key={status}
              className={`kanban__column ${isDragOver ? 'kanban__column--drag-over' : ''}`}
              onDragOver={handleDragOver}
              onDragEnter={() => handleDragEnter(status)}
              onDragLeave={() => handleDragLeave(status)}
              onDrop={(e) => handleDrop(e, status)}
              role="group"
              aria-label={COLUMN_LABELS[status]}
            >
              {/* Column header */}
              <div className="kanban__column-header">
                <span className="kanban__column-label">{COLUMN_LABELS[status]}</span>
                <span className="kanban__column-count">({columnOrders.length})</span>
              </div>

              {/* Cards or empty state */}
              {columnOrders.length > 0 ? (
                <div className="kanban__cards">
                  {columnOrders.map((order) => (
                    <div
                      key={order.id}
                      className={`kanban__card-draggable ${
                        draggedOrderId === order.id ? 'kanban__card-draggable--dragging' : ''
                      }`}
                      draggable
                      onDragStart={(e) => handleDragStart(e, order.id)}
                      onDragEnd={handleDragEnd}
                    >
                      <OrderCard
                        orderId={order.id}
                        time={formatTime(order.scheduledTime)}
                        items={formatItems(order)}
                        type={getDeliveryType(order)}
                        customerName={undefined}
                        price={order.totalAmount}
                        status={order.status as OrderStatus}
                        onAction={(newStatus) => handleStatusChange(order.id, newStatus)}
                      />
                    </div>
                  ))}
                </div>
              ) : (
                <div className="kanban__empty">{t('dashboard.orders.emptyColumn')}</div>
              )}
            </div>
          );
        })}
      </div>

      {/* Toast notification */}
      {toast && (
        <div
          className={`kanban__toast ${toast.isError ? 'kanban__toast--error' : ''}`}
          role="alert"
        >
          {toast.message}
        </div>
      )}
    </div>
  );
}
