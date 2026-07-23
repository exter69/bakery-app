import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { B2BCartSummary } from './B2BCartSummary';
import type { B2BPricingResult } from '../../types/b2b';

// Mock the i18n hook with interpolation support
vi.mock('../../i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, string>) => {
      let str = key;
      if (params) {
        for (const [k, v] of Object.entries(params)) {
          str = str.replace(new RegExp(`\\{${k}\\}`, 'g'), v);
        }
      }
      return str;
    },
  }),
}));

// Mock the API client
const mockComputePricing = vi.fn();
vi.mock('../../api/b2b-client', () => ({
  computePricing: (...args: unknown[]) => mockComputePricing(...args),
}));

describe('B2BCartSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('rendersNothingWhenNoItems', () => {
    const { container } = render(
      <B2BCartSummary bakeryId="b1" items={[]} />
    );
    expect(container.querySelector('.b2b-cart-summary')).toBeNull();
  });

  it('displaysLoadingWhileFetching', () => {
    mockComputePricing.mockReturnValue(new Promise(() => {}));

    render(
      <B2BCartSummary
        bakeryId="b1"
        items={[{ productId: 'p1', productName: 'Croissant', unitPrice: 200, quantity: 3 }]}
      />
    );

    expect(screen.getByText('comptoir.common.loading')).toBeInTheDocument();
  });

  it('displaysFullPricingBreakdownWithVolumeTier', async () => {
    const pricing: B2BPricingResult = {
      subtotalHt: 10000,
      proDiscountRate: 1,
      proDiscountAmt: 100,
      volDiscountRate: 8,
      volDiscountAmt: 792,
      tvaRate: 6,
      tvaAmount: 546,
      totalTtc: 9654,
      currentTier: { minMonthlySpend: 150000, discountPercent: 8 },
      nextTier: { minMonthlySpend: 200000, discountPercent: 10 },
      monthlySpend: 160000,
      spendToNextTier: 40000,
    };
    mockComputePricing.mockResolvedValue(pricing);

    render(
      <B2BCartSummary
        bakeryId="b1"
        items={[{ productId: 'p1', productName: 'Pain', unitPrice: 500, quantity: 20 }]}
      />
    );

    await waitFor(() => {
      // Subtotal
      expect(screen.getByText('100.00 EUR')).toBeInTheDocument();
      // Pro discount line
      expect(screen.getByText('-1.00 EUR')).toBeInTheDocument();
      // Volume discount line
      expect(screen.getByText('-7.92 EUR')).toBeInTheDocument();
      // Total
      expect(screen.getByText('96.54 EUR')).toBeInTheDocument();
    });
  });

  it('displaysMaxTierReachedWhenNoNextTier', async () => {
    const pricing: B2BPricingResult = {
      subtotalHt: 50000,
      proDiscountRate: 1,
      proDiscountAmt: 500,
      volDiscountRate: 10,
      volDiscountAmt: 4950,
      tvaRate: 6,
      tvaAmount: 2673,
      totalTtc: 47223,
      currentTier: { minMonthlySpend: 200000, discountPercent: 10 },
      monthlySpend: 250000,
      spendToNextTier: 0,
    };
    mockComputePricing.mockResolvedValue(pricing);

    render(
      <B2BCartSummary
        bakeryId="b1"
        items={[{ productId: 'p1', productName: 'Pain', unitPrice: 5000, quantity: 10 }]}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('comptoir.pricing.maxTier')).toBeInTheDocument();
    });
  });

  it('hidesVolumeDiscountLineWhenZero', async () => {
    const pricing: B2BPricingResult = {
      subtotalHt: 5000,
      proDiscountRate: 1,
      proDiscountAmt: 50,
      volDiscountRate: 0,
      volDiscountAmt: 0,
      tvaRate: 6,
      tvaAmount: 297,
      totalTtc: 5247,
      monthlySpend: 5000,
      spendToNextTier: 145000,
      nextTier: { minMonthlySpend: 150000, discountPercent: 8 },
    };
    mockComputePricing.mockResolvedValue(pricing);

    render(
      <B2BCartSummary
        bakeryId="b1"
        items={[{ productId: 'p1', productName: 'Pain', unitPrice: 500, quantity: 10 }]}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('52.47 EUR')).toBeInTheDocument();
    });
    // Volume discount line should not be present
    expect(screen.queryByText(/comptoir\.pricing\.volumeDiscount/)).not.toBeInTheDocument();
  });
});
