import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { I18nProvider } from '../../i18n/I18nContext';
import DashboardProducts from './DashboardProducts';

// Mock seller API
vi.mock('../../api/seller', () => ({
  fetchMyBakery: vi.fn(),
  fetchProducts: vi.fn(),
  createProduct: vi.fn(),
  updateProduct: vi.fn(),
}));

import { fetchMyBakery, fetchProducts, updateProduct } from '../../api/seller';

const mockBakery = {
  id: 'bakery-1',
  name: 'Le Fournil',
  photoUrl: '',
  description: '',
  address: '1 rue du Pain',
  schedule: [],
  createdAt: '2024-01-01T00:00:00Z',
};

const mockProducts = [
  {
    id: 'p1',
    bakeryId: 'bakery-1',
    name: 'Croissant',
    description: 'Butter croissant',
    price: 150,
    photoUrl: '',
    category: 'Viennoiseries',
    isAvailable: true,
    allergens: ['gluten'],
    healthScore: 3,
  },
  {
    id: 'p2',
    bakeryId: 'bakery-1',
    name: 'Baguette tradition',
    description: 'French baguette',
    price: 120,
    photoUrl: '',
    category: 'Pains',
    isAvailable: true,
    allergens: ['gluten'],
    healthScore: 4,
  },
  {
    id: 'p3',
    bakeryId: 'bakery-1',
    name: 'Tarte aux pommes',
    description: 'Apple tart',
    price: 450,
    photoUrl: '',
    category: 'Pâtisseries',
    isAvailable: false,
    allergens: ['gluten', 'eggs'],
    healthScore: 2,
  },
];

function renderProducts() {
  return render(<I18nProvider><DashboardProducts /></I18nProvider>);
}

describe('DashboardProducts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockResolvedValue(mockBakery);
    (fetchProducts as ReturnType<typeof vi.fn>).mockResolvedValue(mockProducts);
    (updateProduct as ReturnType<typeof vi.fn>).mockResolvedValue(mockProducts[0]);
  });

  it('renders products as cards with product names in the DOM', async () => {
    renderProducts();
    await waitFor(() => {
      expect(screen.getByText('Croissant')).toBeInTheDocument();
      expect(screen.getByText('Baguette tradition')).toBeInTheDocument();
      expect(screen.getByText('Tarte aux pommes')).toBeInTheDocument();
    });
  });

  it('category filter shows only matching products', async () => {
    const user = userEvent.setup();
    renderProducts();

    await waitFor(() => {
      expect(screen.getByText('Croissant')).toBeInTheDocument();
    });

    // Click the "Pains" category filter
    await user.click(screen.getByText('Pains'));

    await waitFor(() => {
      expect(screen.getByText('Baguette tradition')).toBeInTheDocument();
      expect(screen.queryByText('Croissant')).not.toBeInTheDocument();
      expect(screen.queryByText('Tarte aux pommes')).not.toBeInTheDocument();
    });
  });

  it('stock stepper calls API when +/- is clicked', async () => {
    const user = userEvent.setup();
    renderProducts();

    await waitFor(() => {
      expect(screen.getByText('Croissant')).toBeInTheDocument();
    });

    // Find the increment buttons (there is one per product card)
    const incrementButtons = screen.getAllByLabelText('Augmenter');
    expect(incrementButtons.length).toBeGreaterThan(0);

    // Click increment on first product
    await user.click(incrementButtons[0]);

    await waitFor(() => {
      expect(updateProduct).toHaveBeenCalled();
    });
  });

  it('visibility toggle updates product', async () => {
    const user = userEvent.setup();
    // Return updated product when toggling visibility
    (updateProduct as ReturnType<typeof vi.fn>).mockResolvedValue({
      ...mockProducts[0],
      isAvailable: false,
    });
    renderProducts();

    await waitFor(() => {
      expect(screen.getByText('Croissant')).toBeInTheDocument();
    });

    // Find visibility toggle buttons — the available ones show "en vente"
    const visibleButtons = screen.getAllByText('en vente');
    expect(visibleButtons.length).toBeGreaterThan(0);

    await user.click(visibleButtons[0]);

    await waitFor(() => {
      expect(updateProduct).toHaveBeenCalledWith('p1', { isAvailable: false });
    });
  });

  it('shows loading state initially', () => {
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    renderProducts();
    expect(screen.getByText(/Loading products/)).toBeInTheDocument();
  });
});
