import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { I18nProvider } from '../../i18n/I18nContext';
import DashboardLayout from './DashboardLayout';

// Mock seller API
vi.mock('../../api/seller', () => ({
  fetchMyBakery: vi.fn(),
  fetchBakeryOrders: vi.fn(),
}));

// Mock ThemeSwitcher to simplify tests
vi.mock('../../components/ThemeSwitcher', () => ({
  ThemeSwitcher: () => <div data-testid="theme-switcher">Theme</div>,
}));

import { fetchMyBakery, fetchBakeryOrders } from '../../api/seller';

const mockBakery = {
  id: 'bakery-1',
  name: 'Le Fournil',
  photoUrl: '',
  description: 'Test',
  address: '1 rue',
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

function renderLayout() {
  return render(
    <I18nProvider>
      <MemoryRouter initialEntries={['/dashboard']}>
        <DashboardLayout />
      </MemoryRouter>
    </I18nProvider>
  );
}

describe('DashboardLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    (fetchMyBakery as ReturnType<typeof vi.fn>).mockResolvedValue(mockBakery);
    (fetchBakeryOrders as ReturnType<typeof vi.fn>).mockResolvedValue({ items: [], page: 1, pageSize: 20, total: 3 });
  });

  it('renders the "Votre Boulangerie" brand in sidebar', async () => {
    renderLayout();
    await waitFor(() => {
      expect(screen.getByText('Votre Boulangerie')).toBeInTheDocument();
    });
  });

  it('renders navigation links with French labels', async () => {
    renderLayout();
    await waitFor(() => {
      expect(screen.getByText('Tableau de bord')).toBeInTheDocument();
      expect(screen.getByText('Commandes')).toBeInTheDocument();
      expect(screen.getByText('Menu & stock')).toBeInTheDocument();
      expect(screen.getByText('Paniers du soir')).toBeInTheDocument();
      expect(screen.getByText('Planning')).toBeInTheDocument();
      expect(screen.getByText('Boutique')).toBeInTheDocument();
    });
  });

  it('shows order count badge when orders exist', async () => {
    renderLayout();
    await waitFor(() => {
      expect(screen.getByText('3')).toBeInTheDocument();
    });
  });

  it('renders bakery info in footer when loaded', async () => {
    renderLayout();
    await waitFor(() => {
      expect(screen.getByText('Le Fournil')).toBeInTheDocument();
    });
  });

  it('collapses sidebar when toggle is clicked', async () => {
    const user = userEvent.setup();
    renderLayout();
    const toggle = screen.getByLabelText('Collapse sidebar');
    await user.click(toggle);
    expect(localStorage.getItem('dashboard_sidebar_collapsed')).toBe('true');
  });

  it('shows abbreviation when collapsed', async () => {
    localStorage.setItem('dashboard_sidebar_collapsed', 'true');
    renderLayout();
    // Collapsed state shows "VB" abbreviation instead of the full brand name
    expect(screen.getByText('VB')).toBeInTheDocument();
    expect(screen.getByLabelText('Expand sidebar')).toBeInTheDocument();
  });

  it('includes the ThemeSwitcher in footer', () => {
    renderLayout();
    expect(screen.getByTestId('theme-switcher')).toBeInTheDocument();
  });

  it('renders sign out button', () => {
    renderLayout();
    expect(screen.getByText('Sign Out')).toBeInTheDocument();
  });
});
