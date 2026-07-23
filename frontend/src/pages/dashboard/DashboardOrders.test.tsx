import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { I18nProvider } from '../../i18n/I18nContext';
import DashboardOrders from './DashboardOrders';

// Mock seller API
vi.mock('../../api/seller', () => ({
  fetchMyBakery: vi.fn(),
  fetchBakeryOrders: vi.fn(),
  updateOrderStatus: vi.fn(),
}));

// Mock useWebSocket hook
vi.mock('../../hooks/useWebSocket', () => ({
  useWebSocket: () => ({ lastEvent: null, connected: false }),
}));

import { fetchMyBakery, fetchBakeryOrders, updateOrderStatus } from '../../api/seller';

const mockBakery = {
  id: 'bakery-1',
  name: 'Le Fournil',
  photoUrl: '',
  description: '',
  address: '1 rue du Pain',
  schedule: [],
  createdAt: '2024-01-01T00:00:00Z',
};

function makeOrder(id: string, status: string, type: 'order' | 'reservation' = 'order') {
  return {
    id,
    type,
    bakeryId: 'bakery-1',
    items: [{ productId: 'p1', productName: 'Croissant', quantity: 2, unitPrice: 150, subtotal: 300 }],
    scheduledDay: 'monday',
    scheduledTime: { startTime: '09:00', endTime: '09:30' },
    status,
    totalAmount: 300,
    createdAt: '2024-06-01T06:00:00Z',
  };
}

const mockOrders = {
  items: [
    makeOrder('o1', 'confirmed', 'order'),
    makeOrder('o2', 'preparing', 'order'),
    makeOrder('o3', 'ready', 'reservation'),
    makeOrder('o4', 'delivered', 'reservation'),
  ],
  page: 1,
  pageSize: 20,
  total: 4,
};

function renderOrders() {
  return render(
    <I18nProvider>
      <MemoryRouter>
        <DashboardOrders />
      </MemoryRouter>
    </I18nProvider>,
  );
}

describe('DashboardOrders', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockResolvedValue(mockBakery);
    (fetchBakeryOrders as ReturnType<typeof vi.fn>).mockResolvedValue(mockOrders);
    (updateOrderStatus as ReturnType<typeof vi.fn>).mockResolvedValue(makeOrder('o1', 'preparing'));
  });

  it('renders 4 kanban columns with correct labels', async () => {
    renderOrders();
    await waitFor(() => {
      expect(screen.getByText('TO PREPARE')).toBeInTheDocument();
      expect(screen.getByText('PREPARING')).toBeInTheDocument();
      expect(screen.getByText('READY')).toBeInTheDocument();
      expect(screen.getByText('DELIVERED')).toBeInTheDocument();
    });
  });

  it('filter chips show/hide orders by type (delivery vs pickup)', async () => {
    const user = userEvent.setup();
    renderOrders();

    await waitFor(() => {
      expect(screen.getByText('TO PREPARE')).toBeInTheDocument();
    });

    // Click "Delivery" filter to show only orders of type 'order'
    await user.click(screen.getByText('Delivery'));

    // Only 'order' type entries should be visible (o1 confirmed, o2 preparing)
    await waitFor(() => {
      const confirmedColumn = screen.getByRole('group', { name: 'TO PREPARE' });
      expect(confirmedColumn).toHaveTextContent('(1)');
    });

    // Click "Pickup" filter
    await user.click(screen.getByText('Pickup'));

    await waitFor(() => {
      const confirmedColumn = screen.getByRole('group', { name: 'TO PREPARE' });
      expect(confirmedColumn).toHaveTextContent('(0)');
    });
  });

  it('action button on a card triggers API call with correct status', async () => {
    const user = userEvent.setup();
    renderOrders();

    await waitFor(() => {
      expect(screen.getByText('TO PREPARE')).toBeInTheDocument();
    });

    // The confirmed order should have "Commencer" action button
    const actionButton = screen.getByText('Commencer');
    expect(actionButton).toBeInTheDocument();

    await user.click(actionButton);

    expect(updateOrderStatus).toHaveBeenCalledWith('o1', 'preparing');
  });

  it('shows loading state initially', () => {
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    renderOrders();
    expect(screen.getByText(/Loading orders/)).toBeInTheDocument();
  });

  it('shows error banner when API fails and no orders loaded', async () => {
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Network error'));
    renderOrders();
    await waitFor(() => {
      expect(screen.getByText(/Unable to load bakery/)).toBeInTheDocument();
    });
  });
});
