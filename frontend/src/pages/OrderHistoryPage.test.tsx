import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import OrderHistoryPage from './OrderHistoryPage';

// Mock navigate
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock the API module
vi.mock('../api/orders', () => ({
  fetchOrderHistory: vi.fn(),
  storeReorderData: vi.fn(),
}));

// Mock i18n
vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const map: Record<string, string> = {
        'history.title': 'Order History',
        'history.empty': 'No orders yet',
        'history.reorder': 'Re-order',
        'history.unavailable': 'No longer available',
        'history.type.delivery': 'Delivery',
        'history.type.pickup': 'Pickup',
        'history.status.delivered': 'Delivered',
        'history.status.pickedUp': 'Picked up',
        'history.previous': 'Previous',
        'history.next': 'Next',
      };
      return map[key] ?? key;
    },
  }),
}));

import { fetchOrderHistory, storeReorderData } from '../api/orders';
const mockFetchHistory = fetchOrderHistory as ReturnType<typeof vi.fn>;
const mockStoreReorder = storeReorderData as ReturnType<typeof vi.fn>;

function renderPage() {
  return render(
    <MemoryRouter>
      <OrderHistoryPage />
    </MemoryRouter>,
  );
}

const sampleEntries = [
  {
    id: 'order-1',
    type: 'order' as const,
    bakeryId: 'bakery-abc',
    items: [
      { productId: 'p1', productName: 'Croissant', quantity: 2, unitPrice: 150, subtotal: 300 },
      { productId: 'p2', productName: 'Baguette', quantity: 1, unitPrice: 250, subtotal: 250 },
    ],
    scheduledTime: '2025-01-15T08:00:00Z',
    status: 'delivered' as const,
    totalAmount: 550,
    createdAt: '2025-01-14T10:00:00Z',
  },
  {
    id: 'res-2',
    type: 'reservation' as const,
    bakeryId: 'bakery-xyz',
    items: [
      { productId: 'p3', productName: 'Pain au chocolat', quantity: 3, unitPrice: 200, subtotal: 600 },
    ],
    scheduledTime: '2025-01-10T09:00:00Z',
    status: 'picked_up' as const,
    totalAmount: 600,
    createdAt: '2025-01-09T15:00:00Z',
  },
];

describe('OrderHistoryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders title and shows loading state initially', () => {
    mockFetchHistory.mockReturnValue(new Promise(() => {})); // never resolves
    renderPage();
    expect(screen.getByText('Order History')).toBeInTheDocument();
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('renders empty state when no orders exist', async () => {
    mockFetchHistory.mockResolvedValueOnce({ items: [], page: 1, pageSize: 20, total: 0 });
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('No orders yet')).toBeInTheDocument();
    });
  });

  it('renders order cards with items and status badges', async () => {
    mockFetchHistory.mockResolvedValueOnce({ items: sampleEntries, page: 1, pageSize: 20, total: 2 });
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('Delivered')).toBeInTheDocument();
    });
    expect(screen.getByText('Picked up')).toBeInTheDocument();
    expect(screen.getByText('2x Croissant, 1x Baguette')).toBeInTheDocument();
    expect(screen.getByText('3x Pain au chocolat')).toBeInTheDocument();
    expect(screen.getByText('Delivery')).toBeInTheDocument();
    expect(screen.getByText('Pickup')).toBeInTheDocument();
  });

  it('displays formatted total amounts', async () => {
    mockFetchHistory.mockResolvedValueOnce({ items: sampleEntries, page: 1, pageSize: 20, total: 2 });
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('€5.50')).toBeInTheDocument();
    });
    expect(screen.getByText('€6.00')).toBeInTheDocument();
  });

  it('renders error state with retry button', async () => {
    mockFetchHistory.mockRejectedValueOnce(new Error('Network error'));
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('Failed to load order history')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: 'Retry' })).toBeInTheDocument();
  });

  it('retries fetching on retry button click', async () => {
    mockFetchHistory.mockRejectedValueOnce(new Error('Network error'));
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('Failed to load order history')).toBeInTheDocument();
    });

    mockFetchHistory.mockResolvedValueOnce({ items: sampleEntries, page: 1, pageSize: 20, total: 2 });
    await userEvent.click(screen.getByRole('button', { name: 'Retry' }));
    await waitFor(() => {
      expect(screen.getByText('2x Croissant, 1x Baguette')).toBeInTheDocument();
    });
  });

  it('stores re-order data and navigates to bakery on re-order click', async () => {
    mockFetchHistory.mockResolvedValueOnce({ items: sampleEntries, page: 1, pageSize: 20, total: 2 });
    renderPage();
    await waitFor(() => {
      expect(screen.getAllByText('Re-order')).toHaveLength(2);
    });

    const reorderButtons = screen.getAllByText('Re-order');
    await userEvent.click(reorderButtons[0]);

    expect(mockStoreReorder).toHaveBeenCalledWith({
      bakeryId: 'bakery-abc',
      items: [
        { productId: 'p1', productName: 'Croissant', quantity: 2 },
        { productId: 'p2', productName: 'Baguette', quantity: 1 },
      ],
    });
    expect(mockNavigate).toHaveBeenCalledWith('/bakeries/bakery-abc');
  });

  it('handles pagination navigation', async () => {
    mockFetchHistory.mockResolvedValueOnce({ items: sampleEntries, page: 1, pageSize: 20, total: 40 });
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('1 / 2')).toBeInTheDocument();
    });

    const prevBtn = screen.getByRole('button', { name: 'Previous' });
    const nextBtn = screen.getByRole('button', { name: 'Next' });

    expect(prevBtn).toBeDisabled();
    expect(nextBtn).not.toBeDisabled();

    mockFetchHistory.mockResolvedValueOnce({ items: sampleEntries, page: 2, pageSize: 20, total: 40 });
    await userEvent.click(nextBtn);
    expect(mockFetchHistory).toHaveBeenCalledWith(2);
  });

  it('disables next button on last page', async () => {
    mockFetchHistory.mockResolvedValueOnce({ items: sampleEntries, page: 1, pageSize: 20, total: 2 });
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('1 / 1')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled();
  });
});
