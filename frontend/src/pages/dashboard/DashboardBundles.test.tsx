import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import DashboardBundlesPage from './DashboardBundlesPage';

// Mock bundles API
vi.mock('../../api/bundles', () => ({
  listBundles: vi.fn(),
  publishBundle: vi.fn(),
}));

import { listBundles, publishBundle } from '../../api/bundles';

const mockBundles = [
  {
    id: 'bundle-1',
    bakeryId: 'bakery-1',
    bakeryName: 'Le Fournil',
    bakeryLatitude: 48.85,
    bakeryLongitude: 2.35,
    name: 'Panier surprise',
    type: 'surprise' as const,
    photoUrl: '',
    description: 'Bundle of leftover pastries',
    estimatedValue: 1200,
    originalPrice: 1200,
    discountedPrice: 540,
    quantityTotal: 5,
    quantityRemaining: 3,
    pickupStartTime: '18:00',
    pickupEndTime: '19:30',
    publishedDate: '2024-06-01',
    expiresAt: '2024-06-01T20:00:00Z',
    status: 'draft' as const,
    items: [],
    createdAt: '2024-06-01T10:00:00Z',
  },
  {
    id: 'bundle-2',
    bakeryId: 'bakery-1',
    bakeryName: 'Le Fournil',
    bakeryLatitude: 48.85,
    bakeryLongitude: 2.35,
    name: 'Panier viennoiseries',
    type: 'compose' as const,
    photoUrl: '',
    description: 'Leftover croissants and pains au chocolat',
    estimatedValue: 800,
    originalPrice: 800,
    discountedPrice: 360,
    quantityTotal: 3,
    quantityRemaining: 1,
    pickupStartTime: '17:30',
    pickupEndTime: '19:00',
    publishedDate: '2024-06-01',
    expiresAt: '2024-06-01T20:00:00Z',
    status: 'published' as const,
    items: [{ description: 'Croissant', quantity: 3 }],
    createdAt: '2024-06-01T09:00:00Z',
  },
];

function renderBundles() {
  return render(<DashboardBundlesPage />);
}

describe('DashboardBundlesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (listBundles as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: mockBundles,
      page: 1,
      pageSize: 20,
      total: 2,
    });
    (publishBundle as ReturnType<typeof vi.fn>).mockResolvedValue(mockBundles[0]);
  });

  it('renders product list with remaining stock labels', async () => {
    renderBundles();
    await waitFor(() => {
      expect(screen.getByText('Panier surprise')).toBeInTheDocument();
      expect(screen.getByText('Panier viennoiseries')).toBeInTheDocument();
      // Remaining stock is shown in the card info
      expect(screen.getByText(/Stock: 3\/5/)).toBeInTheDocument();
      expect(screen.getByText(/Stock: 1\/3/)).toBeInTheDocument();
    });
  });

  it('"Publier" button calls publishBundle API', async () => {
    const user = userEvent.setup();
    renderBundles();

    await waitFor(() => {
      expect(screen.getByText('Panier surprise')).toBeInTheDocument();
    });

    // Only draft bundles show "Publier" button
    const publishButton = screen.getByText('Publier');
    expect(publishButton).toBeInTheDocument();

    await user.click(publishButton);

    expect(publishBundle).toHaveBeenCalledWith('bundle-1');
  });

  it('bundle creation form appears when "Nouveau panier" is clicked', async () => {
    const user = userEvent.setup();
    renderBundles();

    await waitFor(() => {
      expect(screen.getByText('Panier surprise')).toBeInTheDocument();
    });

    const newButton = screen.getByText('+ Nouveau panier');
    await user.click(newButton);

    // The BundleForm component should be rendered
    // The "Nouveau panier" button should disappear
    expect(screen.queryByText('+ Nouveau panier')).not.toBeInTheDocument();
  });

  it('shows loading state initially', () => {
    (listBundles as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    renderBundles();
    expect(screen.getByText(/Loading bundles/)).toBeInTheDocument();
  });

  it('shows empty state when no bundles exist', async () => {
    (listBundles as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: [],
      page: 1,
      pageSize: 20,
      total: 0,
    });
    renderBundles();
    await waitFor(() => {
      expect(screen.getByText(/No bundles yet/)).toBeInTheDocument();
    });
  });
});
