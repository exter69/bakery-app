import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';
import { calculateBundlePrice, capQuantity } from './bundle-utils';

/**
 * Property 5: Bundle price discount invariant
 * For any non-empty list of items with positive prices, discountedPrice < originalPrice.
 *
 * **Validates: Requirements 5.4, 5.6**
 */
describe('bundle-utils - Property 5: Bundle price discount invariant', () => {
  const selectedItemArb = fc.record({
    productId: fc.uuid(),
    name: fc.string({ minLength: 1, maxLength: 20 }),
    remaining: fc.integer({ min: 1, max: 100 }),
    selected: fc.constant(true),
    quantity: fc.integer({ min: 1, max: 20 }),
    price: fc.integer({ min: 1, max: 10000 }), // positive price in cents
  });

  it('discountedPrice is strictly less than originalPrice for non-empty selected items', () => {
    fc.assert(
      fc.property(
        fc.array(selectedItemArb, { minLength: 1, maxLength: 20 }),
        fc.double({ min: 0.01, max: 0.99, noNaN: true }),
        (items, discountFactor) => {
          const { originalPrice, discountedPrice } = calculateBundlePrice(items, discountFactor);

          // With at least one selected item with positive price and quantity,
          // originalPrice > 0 and discountedPrice < originalPrice
          expect(originalPrice).toBeGreaterThan(0);
          expect(discountedPrice).toBeLessThan(originalPrice);
          expect(discountedPrice).toBeGreaterThanOrEqual(0);
        },
      ),
      { numRuns: 200 },
    );
  });

  it('returns zero prices when no items are selected', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.record({
            productId: fc.uuid(),
            name: fc.string({ minLength: 1, maxLength: 20 }),
            remaining: fc.integer({ min: 0, max: 100 }),
            selected: fc.constant(false),
            quantity: fc.integer({ min: 1, max: 20 }),
            price: fc.integer({ min: 1, max: 10000 }),
          }),
          { minLength: 0, maxLength: 10 },
        ),
        (items) => {
          const { originalPrice, discountedPrice } = calculateBundlePrice(items);
          expect(originalPrice).toBe(0);
          expect(discountedPrice).toBe(0);
        },
      ),
      { numRuns: 100 },
    );
  });
});

/**
 * Property 6: Bundle item quantity bounds
 * capQuantity(requested, remaining) always returns min(requested, remaining) and is >= 0.
 *
 * **Validates: Requirements 5.4, 5.6**
 */
describe('bundle-utils - Property 6: Bundle item quantity bounds', () => {
  it('capQuantity returns min(requested, remaining) and is always >= 0', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: -50, max: 200 }),  // requested (can be negative)
        fc.integer({ min: 0, max: 200 }),    // remaining (non-negative stock)
        (requested, remaining) => {
          const result = capQuantity(requested, remaining);

          // Result is always >= 0
          expect(result).toBeGreaterThanOrEqual(0);

          // Result is always <= remaining
          expect(result).toBeLessThanOrEqual(remaining);

          // Result is always <= requested (when requested is positive)
          if (requested >= 0) {
            expect(result).toBeLessThanOrEqual(requested);
          }

          // Result equals Math.max(0, Math.min(requested, remaining))
          expect(result).toBe(Math.max(0, Math.min(requested, remaining)));
        },
      ),
      { numRuns: 200 },
    );
  });
});
