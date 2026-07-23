import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { StatCard } from './StatCard';

describe('StatCard', () => {
  it('renders label, value, and subtitle', () => {
    render(<StatCard label="Commandes du jour" value={12} subtitle="dont 3 à préparer" />);

    expect(screen.getByText('Commandes du jour')).toBeInTheDocument();
    expect(screen.getByText('12')).toBeInTheDocument();
    expect(screen.getByText('dont 3 à préparer')).toBeInTheDocument();
  });

  it('renders badge when provided', () => {
    render(
      <StatCard
        label="Recette du jour"
        value="€150"
        subtitle="encaissée"
        badge={{ text: '+12%', variant: 'positive' }}
      />
    );

    expect(screen.getByText('+12%')).toBeInTheDocument();
  });

  it('does not render badge when omitted', () => {
    const { container } = render(
      <StatCard label="Prochain retrait" value="14:30" subtitle="retrait / livraison" />
    );

    expect(container.querySelector('.stat-card__badge')).toBeNull();
  });
});
