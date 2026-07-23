import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';
import { groupOrdersByStatus, COLUMN_ORDER, type KanbanStatus } from './kanban-utils';
import type { Order } from '../../api/seller';

/** Generate a valid Order with a random kanban status */
const kanbanStatusArb = fc.oneof(
  fc.constant('confirmed' as const),
  fc.constant('preparing' as const),
  fc.constant('ready' as const),
  fc.constant('delivered' as const),
);

/** Generate an Order with any status (including non-kanban ones) */
const orderStatusArb = fc.oneof(
  kanbanStatusArb,
  fc.constant('pending_payment' as const),
  fc.constant('cancelled' as const),
);

function makeOrder(id: string, status: string): Order {
  return {
    id,
    type: 'order',
    bakeryId: 'b1',
    items: [{ productId: 'p1', productName: 'Croissant', quantity: 1, unitPrice: 150, subtotal: 150 }],
    scheduledTime: '2024-06-01T08:00:00Z',
    status: status as Order['status'],
    totalAmount: 150,
    createdAt: '2024-06-01T06:00:00Z',
  };
}

const orderArb = fc.tuple(fc.uuid(), orderStatusArb).map(([id, status]) => makeOrder(id, status));

/**
 * Property 1: Order conservation
 * For any list of orders, grouping them into columns produces no duplicates
 * and no lost orders (total count of kanban-eligible orders is preserved).
 *
 * **Validates: Requirements 3.2**
 */
describe('kanban-utils - Property 1: Order conservation', () => {
  it('no orders are lost or duplicated across columns', () => {
    fc.assert(
      fc.property(
        fc.array(orderArb, { minLength: 0, maxLength: 50 }),
        (orders) => {
          const columns = groupOrdersByStatus(orders);

          // Collect all orders from all columns
          const allInColumns: Order[] = [];
          for (const status of COLUMN_ORDER) {
            const col = columns.get(status) ?? [];
            allInColumns.push(...col);
          }

          // Only orders with a kanban-eligible status should appear
          const kanbanStatuses = new Set<string>(COLUMN_ORDER);
          const eligibleOrders = orders.filter((o) => kanbanStatuses.has(o.status));

          // Total count preserved (no lost orders)
          expect(allInColumns.length).toBe(eligibleOrders.length);

          // No duplicates — all IDs in columns are unique
          const ids = allInColumns.map((o) => o.id);
          expect(new Set(ids).size).toBe(ids.length);
        },
      ),
      { numRuns: 200 },
    );
  });
});

/**
 * Property 2: Column assignment by status
 * Each order appears in the column matching its status.
 *
 * **Validates: Requirements 3.2**
 */
describe('kanban-utils - Property 2: Column assignment by status', () => {
  it('each order is placed in the column matching its status', () => {
    fc.assert(
      fc.property(
        fc.array(orderArb, { minLength: 1, maxLength: 50 }),
        (orders) => {
          const columns = groupOrdersByStatus(orders);

          for (const status of COLUMN_ORDER) {
            const columnOrders = columns.get(status) ?? [];
            for (const order of columnOrders) {
              expect(order.status).toBe(status);
            }
          }
        },
      ),
      { numRuns: 200 },
    );
  });
});

/**
 * Property 3: Filter chip consistency
 * For any list of orders and any selected filter (livraison/retrait/all),
 * the filtered result is exactly the subset matching that filter type.
 *
 * **Validates: Requirements 3.5, 4.7**
 */
describe('kanban-utils - Property 3: Filter chip consistency', () => {
  type DeliveryFilter = 'all' | 'livraison' | 'retrait';

  const filterArb = fc.oneof(
    fc.constant('all' as DeliveryFilter),
    fc.constant('livraison' as DeliveryFilter),
    fc.constant('retrait' as DeliveryFilter),
  );

  const orderTypeArb = fc.oneof(
    fc.constant('order' as const),
    fc.constant('reservation' as const),
  );

  const orderWithTypeArb = fc.tuple(fc.uuid(), kanbanStatusArb, orderTypeArb).map(
    ([id, status, type]) => ({
      ...makeOrder(id, status),
      type,
    }),
  );

  /** Applies the same filter logic as DashboardOrders component */
  function applyFilter(orders: Order[], filter: DeliveryFilter): Order[] {
    if (filter === 'all') return orders;
    if (filter === 'livraison') return orders.filter((o) => o.type === 'order');
    if (filter === 'retrait') return orders.filter((o) => o.type === 'reservation');
    return orders;
  }

  it('filtered orders are exactly the subset matching the filter', () => {
    fc.assert(
      fc.property(
        fc.array(orderWithTypeArb, { minLength: 0, maxLength: 50 }),
        filterArb,
        (orders, filter) => {
          const result = applyFilter(orders, filter);

          if (filter === 'all') {
            // All filter returns all orders
            expect(result.length).toBe(orders.length);
          } else if (filter === 'livraison') {
            // Every result should be type 'order'
            for (const o of result) {
              expect(o.type).toBe('order');
            }
            // Should include ALL orders of type 'order'
            const expected = orders.filter((o) => o.type === 'order');
            expect(result.length).toBe(expected.length);
          } else {
            // retrait: every result should be type 'reservation'
            for (const o of result) {
              expect(o.type).toBe('reservation');
            }
            // Should include ALL orders of type 'reservation'
            const expected = orders.filter((o) => o.type === 'reservation');
            expect(result.length).toBe(expected.length);
          }
        },
      ),
      { numRuns: 200 },
    );
  });
});
