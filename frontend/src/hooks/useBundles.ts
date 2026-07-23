import { useEffect, useState, useCallback, useRef } from 'react';
import {
  listBundles,
  getBundle,
  getBundleImpact,
  reserveBundle,
  confirmReservation,
  cancelBundleReservation,
} from '../api/bundles';
import { useWebSocket } from './useWebSocket';
import type { ListResponse } from '../types/bakery';
import type { Bundle, BundleFilters, BundleImpact, BundleReservation } from '../types/bundle';

// --- Query hooks ---

/** Fetch published bundles with optional filters. Refetches when filters change. */
export function useBundles(filters: BundleFilters) {
  const [data, setData] = useState<ListResponse<Bundle> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const mountedRef = useRef(true);

  const refetch = useCallback(() => {
    setLoading(true);
    setError(null);
    listBundles({
      type: filters.type,
      pickupBefore: filters.pickupBefore,
    })
      .then((res) => {
        if (mountedRef.current) setData(res);
      })
      .catch((err: Error) => {
        if (mountedRef.current) setError(err.message);
      })
      .finally(() => {
        if (mountedRef.current) setLoading(false);
      });
  }, [filters.type, filters.pickupBefore]);

  useEffect(() => {
    mountedRef.current = true;
    refetch();
    return () => {
      mountedRef.current = false;
    };
  }, [refetch]);

  return { data, loading, error, refetch };
}

/** Fetch a single bundle by ID */
export function useBundle(id: string) {
  const [data, setData] = useState<Bundle | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const mountedRef = useRef(true);

  const refetch = useCallback(() => {
    if (!id) return;
    setLoading(true);
    setError(null);
    getBundle(id)
      .then((bundle) => {
        if (mountedRef.current) setData(bundle);
      })
      .catch((err: Error) => {
        if (mountedRef.current) setError(err.message);
      })
      .finally(() => {
        if (mountedRef.current) setLoading(false);
      });
  }, [id]);

  useEffect(() => {
    mountedRef.current = true;
    refetch();
    return () => {
      mountedRef.current = false;
    };
  }, [refetch]);

  return { data, loading, error, refetch };
}

/** Fetch community impact metrics */
export function useBundleImpact() {
  const [data, setData] = useState<BundleImpact | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    getBundleImpact()
      .then((impact) => {
        if (mountedRef.current) setData(impact);
      })
      .catch((err: Error) => {
        if (mountedRef.current) setError(err.message);
      })
      .finally(() => {
        if (mountedRef.current) setLoading(false);
      });
    return () => {
      mountedRef.current = false;
    };
  }, []);

  return { data, loading, error };
}

// --- Mutation hooks ---

/** Mutation to reserve a bundle. Returns mutate fn + state. */
export function useReserveBundle() {
  const [data, setData] = useState<BundleReservation | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mutate = useCallback(async (bundleId: string) => {
    setLoading(true);
    setError(null);
    try {
      const reservation = await reserveBundle(bundleId);
      setData(reservation);
      return reservation;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Reserve failed';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return { mutate, data, loading, error };
}

/** Mutation to confirm a reservation */
export function useConfirmReservation() {
  const [data, setData] = useState<BundleReservation | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mutate = useCallback(async (bundleId: string) => {
    setLoading(true);
    setError(null);
    try {
      const reservation = await confirmReservation(bundleId);
      setData(reservation);
      return reservation;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Confirm failed';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return { mutate, data, loading, error };
}

/** Mutation to cancel a bundle reservation */
export function useCancelBundleReservation() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mutate = useCallback(async (reservationId: string) => {
    setLoading(true);
    setError(null);
    try {
      await cancelBundleReservation(reservationId);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Cancel failed';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return { mutate, loading, error };
}

// --- WebSocket hook ---

interface BundleStockUpdate {
  bundleId: string;
  quantityRemaining: string;
  status: string;
}

/**
 * Listen for real-time bundle WebSocket events.
 * - `bundle_stock_update`: calls onStockUpdate with the updated bundle info.
 * - `bundle_expired`: calls onExpired with the bundleId.
 */
export function useBundleWebSocket(callbacks: {
  onStockUpdate?: (update: BundleStockUpdate) => void;
  onExpired?: (bundleId: string) => void;
}) {
  const { lastEvent } = useWebSocket();
  const callbacksRef = useRef(callbacks);
  callbacksRef.current = callbacks;

  useEffect(() => {
    if (!lastEvent) return;

    if (lastEvent.type === 'bundle_stock_update' && callbacksRef.current.onStockUpdate) {
      callbacksRef.current.onStockUpdate(lastEvent.payload as unknown as BundleStockUpdate);
    }

    if (lastEvent.type === 'bundle_expired' && callbacksRef.current.onExpired) {
      callbacksRef.current.onExpired(lastEvent.payload.bundleId);
    }
  }, [lastEvent]);
}
