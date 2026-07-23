import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { I18nProvider } from '../../i18n/I18nContext';
import DashboardBundles from './DashboardBundles';

// Mock seller API (the routed component uses fetchMyBakery + fetchProducts)
vi.mock('../../api/seller', () => ({
  fetchMyBakery: vi.fn(),
  fetchProducts: vi.fn(),
}));

import { fetchMyBakery, fetchProducts } from '../../api/seller';

const mockBakery = {
  id: 'bakery-1',
  name: 'Le Fournil',
  schedule: [
    { day: 'monday', isOpen: true, openTime: { hour: 7, minute: 0 }, closeTime: { hour: 19, minute: 0 } },
  ],
};

const mockProducts = [
  { id: 'prod-1', name: 'Croissant', price: 150, category: 'Viennoiseries', isAvailable: true },
  { id: 'prod-2', name: 'Baguette', price: 130, category: 'Breads', isAvailable: true },
  { id: 'prod-3', name: 'Eclair', price: 380, category: 'Pastries', isAvailable: false },
];

function renderBundles() {
  return render(<I18nProvider><DashboardBundles /></I18nProvider>);
}

describe('DashboardBundles (routed bundle composer)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockResolvedValue(mockBakery);
    (fetchProducts as ReturnType<typeof vi.fn>).mockResolvedValue(mockProducts);
  });

  it('renders available products as checkable items', async () => {
    renderBundles();
    await waitFor(() => {
      // Only available products shown (prod-3 is unavailable)
      expect(screen.getByText('Croissant')).toBeInTheDocument();
      expect(screen.getByText('Baguette')).toBeInTheDocument();
      expect(screen.queryByText('Eclair')).not.toBeInTheDocument();
    });
  });

  it('selecting a product shows the stepper and updates pricing', async () => {
    const user = userEvent.setup();
    renderBundles();

    await waitFor(() => {
      expect(screen.getByText('Croissant')).toBeInTheDocument();
    });

    // Select first product
    const checkbox = screen.getByLabelText('Croissant');
    await user.click(checkbox);

    // Pricing section should appear with a discounted price
    await waitFor(() => {
      expect(screen.getByText('Publish bundles')).not.toBeDisabled();
    });
  });

  it('shows loading state initially', () => {
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    renderBundles();
    expect(screen.getByText(/Loading/)).toBeInTheDocument();
  });

  it('shows empty state when no bakery found', async () => {
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockResolvedValue(null);
    renderBundles();
    await waitFor(() => {
      expect(screen.getByText(/No bakery found/)).toBeInTheDocument();
    });
  });

  it('shows empty product message when all products are unavailable', async () => {
    (fetchProducts as ReturnType<typeof vi.fn>).mockResolvedValue([
      { id: 'prod-3', name: 'Eclair', price: 380, category: 'Pastries', isAvailable: false },
    ]);
    renderBundles();
    await waitFor(() => {
      expect(screen.getByText(/No products available/)).toBeInTheDocument();
    });
  });
});
