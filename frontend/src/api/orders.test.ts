import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { storeReorderData, consumeReorderData } from './orders';
import type { ReorderData } from './orders';

describe('reorder sessionStorage helpers', () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  afterEach(() => {
    sessionStorage.clear();
  });

  const sampleData: ReorderData = {
    bakeryId: 'bakery-123',
    items: [
      { productId: 'p1', productName: 'Croissant', quantity: 2 },
      { productId: 'p2', productName: 'Baguette', quantity: 1 },
    ],
  };

  it('stores and consumes re-order data from sessionStorage', () => {
    storeReorderData(sampleData);
    const result = consumeReorderData();
    expect(result).toEqual(sampleData);
  });

  it('returns null when no re-order data exists', () => {
    const result = consumeReorderData();
    expect(result).toBeNull();
  });

  it('clears data after consuming', () => {
    storeReorderData(sampleData);
    consumeReorderData();
    const secondResult = consumeReorderData();
    expect(secondResult).toBeNull();
  });

  it('returns null when sessionStorage contains invalid JSON', () => {
    sessionStorage.setItem('reorder_items', 'not valid json');
    const result = consumeReorderData();
    expect(result).toBeNull();
  });

  it('overwrites previous re-order data with new data', () => {
    storeReorderData(sampleData);
    const newData: ReorderData = {
      bakeryId: 'bakery-456',
      items: [{ productId: 'p3', productName: 'Pain', quantity: 5 }],
    };
    storeReorderData(newData);
    const result = consumeReorderData();
    expect(result).toEqual(newData);
  });
});
