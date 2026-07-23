import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Product } from '../../types/bakery';
import { ProductCard } from './ProductCard';

function makeProduct(overrides: Partial<Product> = {}): Product {
  return {
    id: 'prod-1',
    bakeryId: 'bk-1',
    name: 'Croissant',
    description: 'Beurre AOP',
    price: 140,
    photoUrl: 'https://example.com/croissant.jpg',
    category: 'viennoiseries',
    isAvailable: true,
    allergens: ['gluten', 'lait'],
    healthScore: null,
    ...overrides,
  };
}

describe('ProductCard', () => {
  it('rendersProductNameAndDescription', () => {
    render(
      <ProductCard
        product={makeProduct()}
        stock={5}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    expect(screen.getByRole('heading', { name: 'Croissant' })).toBeInTheDocument();
    expect(screen.getByText('Beurre AOP')).toBeInTheDocument();
  });

  it('rendersPriceInEuros', () => {
    render(
      <ProductCard
        product={makeProduct({ price: 250 })}
        stock={3}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    expect(screen.getByText(/2\.50/)).toBeInTheDocument();
  });

  it('rendersAllergensList', () => {
    render(
      <ProductCard
        product={makeProduct({ allergens: ['gluten', 'lait'] })}
        stock={2}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    expect(screen.getByText(/allergènes/)).toHaveTextContent('allergènes : gluten, lait');
  });

  it('doesNotRenderAllergensWhenEmpty', () => {
    render(
      <ProductCard
        product={makeProduct({ allergens: [] })}
        stock={2}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    expect(screen.queryByText(/allergènes/)).not.toBeInTheDocument();
  });

  it('lazyLoadsProductPhoto', () => {
    render(
      <ProductCard
        product={makeProduct()}
        stock={4}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    const img = screen.getByRole('img', { name: 'Croissant' });
    expect(img).toHaveAttribute('loading', 'lazy');
    expect(img).toHaveAttribute('src', 'https://example.com/croissant.jpg');
  });

  it('showsPlaceholderWhenNoPhoto', () => {
    render(
      <ProductCard
        product={makeProduct({ photoUrl: '' })}
        stock={1}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    expect(screen.queryByRole('img')).not.toBeInTheDocument();
    expect(screen.getByText('photo')).toBeInTheDocument();
  });

  it('showsEnVenteBadgeWhenAvailable', () => {
    render(
      <ProductCard
        product={makeProduct({ isAvailable: true })}
        stock={10}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    expect(screen.getByRole('button', { name: /masquer/i })).toHaveTextContent('en vente');
  });

  it('showsMasqueBadgeAndDimsCardWhenHidden', () => {
    const { container } = render(
      <ProductCard
        product={makeProduct({ isAvailable: false })}
        stock={0}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    expect(screen.getByRole('button', { name: /rendre visible/i })).toHaveTextContent('masqué');
    expect(container.querySelector('.product-card--hidden')).toBeInTheDocument();
  });

  it('callsOnToggleVisibilityWhenBadgeClicked', async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();

    render(
      <ProductCard
        product={makeProduct({ id: 'prod-42' })}
        stock={5}
        onStockChange={vi.fn()}
        onToggleVisibility={onToggle}
      />,
    );

    await user.click(screen.getByRole('button', { name: /masquer/i }));
    expect(onToggle).toHaveBeenCalledWith('prod-42');
  });

  it('rendersStockStepper', () => {
    render(
      <ProductCard
        product={makeProduct()}
        stock={24}
        onStockChange={vi.fn()}
        onToggleVisibility={vi.fn()}
      />,
    );

    expect(screen.getByRole('group', { name: /contrôle de stock/i })).toBeInTheDocument();
  });

  it('callsOnStockChangeWithDeltaOnIncrement', async () => {
    const user = userEvent.setup();
    const onStockChange = vi.fn();

    render(
      <ProductCard
        product={makeProduct({ id: 'prod-7' })}
        stock={10}
        onStockChange={onStockChange}
        onToggleVisibility={vi.fn()}
      />,
    );

    await user.click(screen.getByRole('button', { name: 'Augmenter' }));
    expect(onStockChange).toHaveBeenCalledWith('prod-7', 1);
  });
});
