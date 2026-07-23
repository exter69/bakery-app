import { memo } from 'react';
import type { OrderStatus } from '../../types/order';
import './OrderCard.css';

export interface OrderCardProps {
  orderId: string;
  time: string;
  items: string;
  type: 'livraison' | 'retrait';
  customerName?: string;
  price?: number;
  status: OrderStatus;
  onAction?: (newStatus: OrderStatus) => void;
}

/** Maps a current status to its next valid transition and button label. */
const STATUS_ACTIONS: Partial<
  Record<OrderStatus, { label: string; nextStatus: OrderStatus }>
> = {
  confirmed: { label: 'Commencer', nextStatus: 'preparing' },
  preparing: { label: 'Prêt', nextStatus: 'ready' },
  ready: { label: 'Remis', nextStatus: 'delivered' },
};

export const OrderCard = memo(function OrderCard({
  orderId,
  time,
  items,
  type,
  customerName,
  price,
  status,
  onAction,
}: OrderCardProps) {
  const action = STATUS_ACTIONS[status];

  // Delivered cards show a simplified one-liner
  if (status === 'delivered') {
    return (
      <article className="order-card order-card--delivered">
        <span className="order-card__summary">
          #{orderId} · Livré {time}
        </span>
      </article>
    );
  }

  const isPreparing = status === 'preparing';

  return (
    <article
      className={`order-card ${isPreparing ? 'order-card--preparing' : ''}`}
    >
      <header className="order-card__header">
        <span className="order-card__id">#{orderId}</span>
        <span className="order-card__time">{time}</span>
        <span className={`order-card__badge order-card__badge--${type}`}>
          {type}
        </span>
      </header>

      <p className="order-card__items">{items}</p>

      {(customerName || price != null) && (
        <div className="order-card__details">
          {customerName && (
            <span className="order-card__customer">{customerName}</span>
          )}
          {price != null && (
            <span className="order-card__price">
              {(price / 100).toFixed(2)}&nbsp;€
            </span>
          )}
        </div>
      )}

      {action && onAction && (
        <button
          type="button"
          className="order-card__action"
          onClick={() => onAction(action.nextStatus)}
        >
          {action.label}
        </button>
      )}
    </article>
  );
});
