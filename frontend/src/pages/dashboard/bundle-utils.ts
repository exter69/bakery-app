/**
 * Bundle pricing and quantity utilities for the anti-gaspi bundle composer.
 * Pure functions with no side effects — all prices are in cents (integers).
 */

export interface BundleItem {
  productId: string;
  name: string;
  remaining: number;
  selected: boolean;
  quantity: number;
  price: number; // cents per unit
}

export interface BundlePricing {
  originalPrice: number;   // cents
  discountedPrice: number; // cents
}

/** Default anti-gaspi discount: 55% off original price */
const DEFAULT_DISCOUNT_FACTOR = 0.55;

/**
 * Computes bundle pricing from selected items.
 *
 * - originalPrice = sum of (quantity × price) for all selected items
 * - discountedPrice = originalPrice × (1 - discountFactor), rounded to whole cents
 * - Returns { originalPrice: 0, discountedPrice: 0 } when no items are selected
 *
 * @param items - list of bundle items (only selected items count)
 * @param discountFactor - fraction to subtract from original (default 0.55 = 55% off)
 */
export function calculateBundlePrice(
  items: BundleItem[],
  discountFactor: number = DEFAULT_DISCOUNT_FACTOR
): BundlePricing {
  const selectedItems = items.filter((item) => item.selected);

  if (selectedItems.length === 0) {
    return { originalPrice: 0, discountedPrice: 0 };
  }

  const originalPrice = Math.round(
    selectedItems.reduce((sum, item) => sum + item.quantity * item.price, 0)
  );

  const discountedPrice = Math.round(originalPrice * (1 - discountFactor));

  return { originalPrice, discountedPrice };
}

/**
 * Caps a requested quantity at remaining stock, floored at 0.
 * Handles edge cases: negative requested values and zero remaining.
 */
export function capQuantity(requested: number, remaining: number): number {
  return Math.max(0, Math.min(requested, remaining));
}
