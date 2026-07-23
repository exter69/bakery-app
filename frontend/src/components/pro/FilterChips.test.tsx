import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { FilterChips } from './FilterChips';

const options = [
  { value: 'all', label: 'Toutes' },
  { value: 'livraison', label: 'Livraison' },
  { value: 'retrait', label: 'Retrait' },
];

describe('FilterChips', () => {
  it('renders all options as buttons', () => {
    render(<FilterChips options={options} selected="all" onChange={() => {}} />);

    expect(screen.getByText('Toutes')).toBeInTheDocument();
    expect(screen.getByText('Livraison')).toBeInTheDocument();
    expect(screen.getByText('Retrait')).toBeInTheDocument();
  });

  it('active chip has aria-checked true', () => {
    render(<FilterChips options={options} selected="livraison" onChange={() => {}} />);

    const livraison = screen.getByText('Livraison');
    expect(livraison).toHaveAttribute('aria-checked', 'true');

    const toutes = screen.getByText('Toutes');
    expect(toutes).toHaveAttribute('aria-checked', 'false');
  });

  it('clicking a chip calls onChange with that value', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<FilterChips options={options} selected="all" onChange={onChange} />);

    await user.click(screen.getByText('Retrait'));

    expect(onChange).toHaveBeenCalledWith('retrait');
    expect(onChange).toHaveBeenCalledTimes(1);
  });
});
