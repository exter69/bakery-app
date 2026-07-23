import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { OrderCard } from './OrderCard';

describe('OrderCard', () => {
  it('renders time, items text, and type badge', () => {
    render(
      <OrderCard
        orderId="abc123"
        time="14:30"
        items="2x Croissant, 1x Baguette"
        type="livraison"
        status="confirmed"
      />
    );

    expect(screen.getByText('14:30')).toBeInTheDocument();
    expect(screen.getByText('2x Croissant, 1x Baguette')).toBeInTheDocument();
    expect(screen.getByText('livraison')).toBeInTheDocument();
  });

  it('renders retrait badge for pickup orders', () => {
    render(
      <OrderCard
        orderId="def456"
        time="09:00"
        items="1x Pain"
        type="retrait"
        status="confirmed"
      />
    );

    expect(screen.getByText('retrait')).toBeInTheDocument();
  });

  it('action button text is "Commencer" for confirmed status', () => {
    const onAction = vi.fn();
    render(
      <OrderCard
        orderId="abc123"
        time="14:30"
        items="2x Croissant"
        type="livraison"
        status="confirmed"
        onAction={onAction}
      />
    );

    expect(screen.getByText('Commencer')).toBeInTheDocument();
  });

  it('action button text is "Prêt" for preparing status', () => {
    const onAction = vi.fn();
    render(
      <OrderCard
        orderId="abc123"
        time="14:30"
        items="2x Croissant"
        type="livraison"
        status="preparing"
        onAction={onAction}
      />
    );

    expect(screen.getByText('Prêt')).toBeInTheDocument();
  });

  it('action button text is "Remis" for ready status', () => {
    const onAction = vi.fn();
    render(
      <OrderCard
        orderId="abc123"
        time="14:30"
        items="2x Croissant"
        type="retrait"
        status="ready"
        onAction={onAction}
      />
    );

    expect(screen.getByText('Remis')).toBeInTheDocument();
  });

  it('clicking action button calls onAction with next status', async () => {
    const user = userEvent.setup();
    const onAction = vi.fn();
    render(
      <OrderCard
        orderId="abc123"
        time="14:30"
        items="2x Croissant"
        type="livraison"
        status="confirmed"
        onAction={onAction}
      />
    );

    await user.click(screen.getByText('Commencer'));

    expect(onAction).toHaveBeenCalledWith('preparing');
    expect(onAction).toHaveBeenCalledTimes(1);
  });

  it('clicking action button in preparing state calls onAction with ready', async () => {
    const user = userEvent.setup();
    const onAction = vi.fn();
    render(
      <OrderCard
        orderId="abc123"
        time="14:30"
        items="2x Croissant"
        type="livraison"
        status="preparing"
        onAction={onAction}
      />
    );

    await user.click(screen.getByText('Prêt'));

    expect(onAction).toHaveBeenCalledWith('ready');
  });

  it('clicking action button in ready state calls onAction with delivered', async () => {
    const user = userEvent.setup();
    const onAction = vi.fn();
    render(
      <OrderCard
        orderId="abc123"
        time="14:30"
        items="2x Croissant"
        type="retrait"
        status="ready"
        onAction={onAction}
      />
    );

    await user.click(screen.getByText('Remis'));

    expect(onAction).toHaveBeenCalledWith('delivered');
  });
});
