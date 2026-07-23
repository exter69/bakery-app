import { useState, useCallback, useEffect } from 'react';
import { getToken } from '../api/client';
import type { B2BCart, B2BCartItem, B2BCartGroup } from '../types/b2b';

const CART_STORAGE_PREFIX = 'b2b_cart_';

function getStorageKey(): string {
  const token = getToken();
  if (!token) return `${CART_STORAGE_PREFIX}anonymous`;
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return `${CART_STORAGE_PREFIX}${payload.sub || payload.userId || 'unknown'}`;
  } catch {
    return `${CART_STORAGE_PREFIX}unknown`;
  }
}

function loadCart(): B2BCart {
  try {
    const stored = localStorage.getItem(getStorageKey());
    if (stored) {
      const parsed = JSON.parse(stored) as B2BCart;
      // Filter out any empty groups that may have been persisted
      parsed.groups = parsed.groups.filter((g) => g.items.length > 0);
      parsed.totalHt = parsed.groups.reduce((sum, g) => sum + g.subtotalHt, 0);
      return parsed;
    }
  } catch {
    // Corrupted data — start fresh
  }
  return { groups: [], totalHt: 0 };
}

function persistCart(cart: B2BCart): void {
  try {
    localStorage.setItem(getStorageKey(), JSON.stringify(cart));
  } catch {
    // Storage full or unavailable — fail silently
  }
}

function computeGroupSubtotal(items: B2BCartItem[]): number {
  return items.reduce((sum, item) => sum + item.quantity * item.unitPrice, 0);
}

function recomputeTotals(cart: B2BCart): B2BCart {
  for (const group of cart.groups) {
    group.subtotalHt = computeGroupSubtotal(group.items);
  }
  cart.totalHt = cart.groups.reduce((sum, g) => sum + g.subtotalHt, 0);
  return cart;
}

export interface UseB2BCartReturn {
  cart: B2BCart;
  addItem: (bakeryId: string, bakeryName: string, item: Omit<B2BCartItem, 'quantity'>, quantity: number) => void;
  removeItem: (bakeryId: string, productId: string) => void;
  updateQuantity: (bakeryId: string, productId: string, quantity: number) => void;
  clearGroup: (bakeryId: string) => void;
  clearAll: () => void;
  getGroupItems: (bakeryId: string) => B2BCartItem[];
  getItemQuantity: (bakeryId: string, productId: string) => number;
}

export function useB2BCart(): UseB2BCartReturn {
  const [cart, setCart] = useState<B2BCart>(loadCart);

  // Re-load cart when token changes (login/logout)
  useEffect(() => {
    setCart(loadCart());
  }, []);

  const updateCart = useCallback((updater: (prev: B2BCart) => B2BCart) => {
    setCart((prev) => {
      const next = updater({ ...prev, groups: prev.groups.map((g) => ({ ...g, items: [...g.items] })) });
      // Remove empty groups
      next.groups = next.groups.filter((g) => g.items.length > 0);
      recomputeTotals(next);
      persistCart(next);
      return next;
    });
  }, []);

  const addItem = useCallback((bakeryId: string, bakeryName: string, item: Omit<B2BCartItem, 'quantity'>, quantity: number) => {
    if (quantity <= 0) return;
    updateCart((cart) => {
      let group = cart.groups.find((g) => g.bakeryId === bakeryId);
      if (!group) {
        group = { bakeryId, bakeryName, items: [], subtotalHt: 0 };
        cart.groups.push(group);
      }
      const existing = group.items.find((i) => i.productId === item.productId);
      if (existing) {
        existing.quantity += quantity;
      } else {
        group.items.push({ ...item, quantity });
      }
      return cart;
    });
  }, [updateCart]);

  const removeItem = useCallback((bakeryId: string, productId: string) => {
    updateCart((cart) => {
      const group = cart.groups.find((g) => g.bakeryId === bakeryId);
      if (group) {
        group.items = group.items.filter((i) => i.productId !== productId);
      }
      return cart;
    });
  }, [updateCart]);

  const updateQuantity = useCallback((bakeryId: string, productId: string, quantity: number) => {
    if (quantity <= 0) {
      removeItem(bakeryId, productId);
      return;
    }
    updateCart((cart) => {
      const group = cart.groups.find((g) => g.bakeryId === bakeryId);
      if (group) {
        const item = group.items.find((i) => i.productId === productId);
        if (item) {
          item.quantity = quantity;
        }
      }
      return cart;
    });
  }, [updateCart, removeItem]);

  const clearGroup = useCallback((bakeryId: string) => {
    updateCart((cart) => {
      cart.groups = cart.groups.filter((g) => g.bakeryId !== bakeryId);
      return cart;
    });
  }, [updateCart]);

  const clearAll = useCallback(() => {
    const empty: B2BCart = { groups: [], totalHt: 0 };
    persistCart(empty);
    setCart(empty);
  }, []);

  const getGroupItems = useCallback((bakeryId: string): B2BCartItem[] => {
    const group = cart.groups.find((g) => g.bakeryId === bakeryId);
    return group?.items ?? [];
  }, [cart]);

  const getItemQuantity = useCallback((bakeryId: string, productId: string): number => {
    const group = cart.groups.find((g) => g.bakeryId === bakeryId);
    if (!group) return 0;
    const item = group.items.find((i) => i.productId === productId);
    return item?.quantity ?? 0;
  }, [cart]);

  return { cart, addItem, removeItem, updateQuantity, clearGroup, clearAll, getGroupItems, getItemQuantity };
}
