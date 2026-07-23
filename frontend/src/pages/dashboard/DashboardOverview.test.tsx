import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { I18nProvider } from '../../i18n/I18nContext';
import DashboardOverview from './DashboardOverview';

// Mock seller API
vi.mock('../../api/seller', () => ({
  fetchMyBakery: vi.fn(),
  fetchBakeryOrders: vi.fn(),
  fetchBakeryReservations: vi.fn(),
  fetchProducts: vi.fn(),
}));

import { fetchMyBakery, fetchBakeryOrders, fetchBakeryReservations, fetchProducts } from '../../api/seller';

const mockBakery = {
  id: 'bakery-1',
  name: 'Le Fournil',
  photoUrl: '',
  description: 'A test bakery',
  address: '1 rue du Pain',
  schedule: [
    { day: 'monday', openTime: { hour: 7, minute: 0 }, closeTime: { hour: 19, minute: 0 }, isOpen: true },
    { day: 'tuesday', openTime: { hour: 7, minute: 0 }, closeTime: { hour: 19, minute: 0 }, isOpen: true },
    { day: 'wednesday', openTime: { hour: 7, minute: 0 }, closeTime: { hour: 19, minute: 0 }, isOpen: true },
    { day: 'thursday', openTime: { hour: 7, minute: 0 }, closeTime: { hour: 19, minute: 0 }, isOpen: true },
    { day: 'friday', openTime: { hour: 7, minute: 0 }, closeTime: { hour: 19, minute: 0 }, isOpen: true },
    { day: 'saturday', openTime: { hour: 8, minute: 0 }, closeTime: { hour: 14, minute: 0 }, isOpen: true },
    { day: 'sunday', openTime: { hour: 0, minute: 0 }, closeTime: { hour: 0, minute: 0 }, isOpen: false },
  ],
  createdAt: '2024-01-01T00:00:00Z',
};

const mockOrders = {
  items: [
    {
      id: 'order-abc123',
      type: 'order' as const,
      bakeryId: 'bakery-1',
      items: [{ productId: 'p1', productName: 'Croissant', quantity: 3, unitPrice: 150, subtotal: 450 }],
      scheduledDay: 'monday',
      scheduledTime: { startTime: '09:00', endTime: '09:30' },
      status: 'confirmed' as const,
      totalAmount: 450,
      createdAt: '2024-01-01T08:00:00Z',
    },
  ],
  page: 1,
  pageSize: 20,
  total: 1,
};

const mockReservations = {
  items: [
    {
      id: 'res-def456',
      type: 'reservation' as const,
      bakeryId: 'bakery-1',
      items: [{ productId: 'p2', productName: 'Baguette', quantity: 1, unitPrice: 120, subtotal: 120 }],
      scheduledDay: 'monday',
      scheduledTime: { startTime: '10:00', endTime: '10:30' },
      status: 'confirmed' as const,
      totalAmount: 120,
      createdAt: '2024-01-01T09:00:00Z',
    },
  ],
  page: 1,
  pageSize: 20,
  total: 1,
};

const mockProducts = [
  { id: 'p1', bakeryId: 'bakery-1', name: 'Croissant', description: '', price: 150, photoUrl: '', category: 'viennoiseries', isAvailable: true, allergens: [], healthScore: 3 },
  { id: 'p2', bakeryId: 'bakery-1', name: 'Baguette', description: '', price: 120, photoUrl: '', category: 'pains', isAvailable: false, allergens: ['gluten'], healthScore: 4 },
];

function renderOverview() {
  return render(
    <I18nProvider>
      <MemoryRouter>
        <DashboardOverview />
      </MemoryRouter>
    </I18nProvider>
  );
}

describe('DashboardOverview', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockResolvedValue(mockBakery);
    (fetchBakeryOrders as ReturnType<typeof vi.fn>).mockResolvedValue(mockOrders);
    (fetchBakeryReservations as ReturnType<typeof vi.fn>).mockResolvedValue(mockReservations);
    (fetchProducts as ReturnType<typeof vi.fn>).mockResolvedValue(mockProducts);
  });

  it('shows loading state initially', () => {
    // Delay resolution so loading is visible
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    renderOverview();
    expect(screen.getByText('Chargement…')).toBeInTheDocument();
  });

  it('renders greeting with bakery name', async () => {
    renderOverview();
    await waitFor(() => {
      expect(screen.getByText(/Bonjour Le/)).toBeInTheDocument();
    });
  });

  it('renders 3 KPI stat cards with values from mock data', async () => {
    renderOverview();
    await waitFor(() => {
      expect(screen.getByText('Commandes du jour')).toBeInTheDocument();
      expect(screen.getByText('Prochain retrait')).toBeInTheDocument();
      expect(screen.getByText('Recette du jour')).toBeInTheDocument();
    });
  });

  it('renders "À préparer maintenant" section with order items', async () => {
    renderOverview();
    await waitFor(() => {
      expect(screen.getByText('À préparer maintenant')).toBeInTheDocument();
      expect(screen.getByText(/3× Croissant/)).toBeInTheDocument();
      expect(screen.getByText(/1× Baguette/)).toBeInTheDocument();
    });
  });

  it('renders low stock products when some are unavailable', async () => {
    renderOverview();
    await waitFor(() => {
      expect(screen.getByText('Stock faible')).toBeInTheDocument();
      expect(screen.getByText('Baguette')).toBeInTheDocument();
    });
  });

  it('renders anti-gaspi card', async () => {
    renderOverview();
    await waitFor(() => {
      expect(screen.getByText(/Panier du soir/)).toBeInTheDocument();
      expect(screen.getByText('Composer →')).toBeInTheDocument();
    });
  });

  it('shows empty state when no bakery is found', async () => {
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockResolvedValue(null);
    renderOverview();
    await waitFor(() => {
      expect(screen.getByText('Aucune boulangerie trouvée.')).toBeInTheDocument();
    });
  });

  it('shows "tout voir" link pointing to orders page', async () => {
    renderOverview();
    await waitFor(() => {
      const link = screen.getByText('tout voir →');
      expect(link).toBeInTheDocument();
      expect(link.closest('a')).toHaveAttribute('href', '/dashboard/orders');
    });
  });
});
