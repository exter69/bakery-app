import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import SearchBar from './SearchBar';

// Mock the API module
vi.mock('../api/bakeries', () => ({
  searchProducts: vi.fn(),
}));

// Mock the i18n hook
vi.mock('../i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const map: Record<string, string> = {
        'search.placeholder': 'Search products...',
        'search.noResults': 'No products found',
        'search.filters.category': 'Category',
        'search.filters.allergens': 'Exclude allergens',
        'search.filters.healthScore': 'Min. health score',
        'allergen.gluten': 'Gluten',
        'allergen.dairy': 'Dairy',
        'allergen.eggs': 'Eggs',
        'allergen.nuts': 'Nuts',
        'allergen.peanuts': 'Peanuts',
        'allergen.soy': 'Soy',
        'allergen.fish': 'Fish',
        'allergen.crustaceans': 'Crustaceans',
        'allergen.celery': 'Celery',
        'allergen.mustard': 'Mustard',
        'allergen.sesame': 'Sesame',
        'allergen.sulphites': 'Sulphites',
        'allergen.lupin': 'Lupin',
        'allergen.molluscs': 'Molluscs',
      };
      return map[key] ?? key;
    },
  }),
}));

import { searchProducts } from '../api/bakeries';
const mockSearch = searchProducts as ReturnType<typeof vi.fn>;

function renderSearchBar() {
  return render(
    <MemoryRouter>
      <SearchBar />
    </MemoryRouter>,
  );
}

describe('SearchBar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  it('renders search input with placeholder', () => {
    renderSearchBar();
    expect(screen.getByPlaceholderText('Search products...')).toBeInTheDocument();
  });

  it('shows filter panel when toggle is clicked', async () => {
    vi.useRealTimers();
    const user = userEvent.setup();
    renderSearchBar();

    await user.click(screen.getByLabelText('Category'));
    expect(screen.getByText('Exclude allergens')).toBeInTheDocument();
    expect(screen.getByText('Min. health score')).toBeInTheDocument();
  });

  it('calls searchProducts after debounce when query is entered', async () => {
    mockSearch.mockResolvedValue({ items: [], page: 1, pageSize: 20, total: 0 });
    renderSearchBar();

    const input = screen.getByPlaceholderText('Search products...');
    await userEvent.type(input, 'croissant');

    // Advance past debounce
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      expect(mockSearch).toHaveBeenCalledWith(
        expect.objectContaining({ q: 'croissant' }),
      );
    });
  });

  it('displays no results message when search returns empty', async () => {
    mockSearch.mockResolvedValue({ items: [], page: 1, pageSize: 20, total: 0 });
    renderSearchBar();

    const input = screen.getByPlaceholderText('Search products...');
    await userEvent.type(input, 'nonexistent');
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      expect(screen.getByText('No products found')).toBeInTheDocument();
    });
  });

  it('renders product results with bakery name as links', async () => {
    mockSearch.mockResolvedValue({
      items: [
        {
          product: {
            id: 'p1',
            bakeryId: 'b1',
            name: 'Croissant',
            description: 'Buttery',
            price: 250,
            photoUrl: '',
            category: 'Viennoiseries',
            isAvailable: true,
            allergens: ['gluten', 'dairy'],
            healthScore: 3,
          },
          bakeryId: 'b1',
          bakeryName: 'Le Pain Doré',
        },
      ],
      page: 1,
      pageSize: 20,
      total: 1,
    });

    renderSearchBar();
    const input = screen.getByPlaceholderText('Search products...');
    await userEvent.type(input, 'croissant');
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      expect(screen.getByText('Croissant')).toBeInTheDocument();
      expect(screen.getByText('Le Pain Doré')).toBeInTheDocument();
    });

    // The result is a link to the bakery detail page
    const link = screen.getByRole('link', { name: /Croissant/i });
    expect(link).toHaveAttribute('href', '/bakeries/b1');
  });

  it('does not call search when input is cleared', async () => {
    mockSearch.mockResolvedValue({ items: [], page: 1, pageSize: 20, total: 0 });
    renderSearchBar();

    const input = screen.getByPlaceholderText('Search products...');
    await userEvent.type(input, 'test');
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      expect(mockSearch).toHaveBeenCalledTimes(1);
    });

    await userEvent.clear(input);
    vi.advanceTimersByTime(350);

    // Should not call again after clearing — reverts to no-search state
    expect(mockSearch).toHaveBeenCalledTimes(1);
  });
});
