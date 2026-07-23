import type { Order } from '../../api/seller';

/**
 * The 4 kanban statuses used for order management columns.
 * Subset of the full OrderStatus type — only actionable statuses appear on the board.
 */
export type KanbanStatus = 'confirmed' | 'preparing' | 'ready' | 'delivered';

/**
 * Ordered sequence of kanban columns, left to right.
 * Used for rendering and adjacency checks.
 */
export const COLUMN_ORDER: readonly KanbanStatus[] = [
  'confirmed',
  'preparing',
  'ready',
  'delivered',
] as const;

/**
 * Partitions a list of orders into kanban columns by status.
 * Orders whose status doesn't match a kanban column are excluded
 * (e.g. pending_payment, cancelled).
 */
export function groupOrdersByStatus(orders: Order[]): Map<KanbanStatus, Order[]> {
  const columns = new Map<KanbanStatus, Order[]>(
    COLUMN_ORDER.map((status) => [status, []])
  );

  for (const order of orders) {
    const bucket = columns.get(order.status as KanbanStatus);
    if (bucket) {
      bucket.push(order);
    }
  }

  return columns;
}

/**
 * Returns true when `to` is the immediate next column after `from`.
 * Only forward single-step transitions are valid for drag-and-drop:
 *   confirmed → preparing → ready → delivered
 */
export function isAdjacentTransition(from: KanbanStatus, to: KanbanStatus): boolean {
  const fromIndex = COLUMN_ORDER.indexOf(from);
  const toIndex = COLUMN_ORDER.indexOf(to);

  // Unknown statuses or same column are not valid transitions
  if (fromIndex === -1 || toIndex === -1) return false;

  return toIndex === fromIndex + 1;
}
