import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import fc from 'fast-check';
import AllergenMultiSelect from '../components/dashboard/AllergenMultiSelect';
import HealthScoreInput from '../components/dashboard/HealthScoreInput';

const ALL_ALLERGENS = [
  'gluten',
  'crustaceans',
  'eggs',
  'fish',
  'peanuts',
  'soy',
  'dairy',
  'nuts',
  'celery',
  'mustard',
  'sesame',
  'sulphites',
  'lupin',
  'molluscs',
] as const;

describe('AllergenMultiSelect', () => {
  it('renders all 14 allergen checkboxes', () => {
    const onChange = vi.fn();
    render(<AllergenMultiSelect selected={[]} onChange={onChange} />);

    const checkboxes = screen.getAllByRole('checkbox');
    expect(checkboxes).toHaveLength(14);

    // Verify each allergen label is present
    for (const allergen of ALL_ALLERGENS) {
      const capitalized = allergen.charAt(0).toUpperCase() + allergen.slice(1);
      expect(screen.getByLabelText(capitalized)).toBeInTheDocument();
    }
  });

  it('calls onChange with allergen added when checking a box', () => {
    const onChange = vi.fn();
    render(<AllergenMultiSelect selected={[]} onChange={onChange} />);

    const glutenCheckbox = screen.getByLabelText('Gluten');
    fireEvent.click(glutenCheckbox);

    expect(onChange).toHaveBeenCalledWith(['gluten']);
  });

  it('calls onChange with allergen removed when unchecking a box', () => {
    const onChange = vi.fn();
    render(<AllergenMultiSelect selected={['gluten', 'eggs']} onChange={onChange} />);

    const glutenCheckbox = screen.getByLabelText('Gluten');
    fireEvent.click(glutenCheckbox);

    expect(onChange).toHaveBeenCalledWith(['eggs']);
  });

  it('shows pre-selected values as checked', () => {
    const selected = ['dairy', 'nuts', 'sesame'];
    const onChange = vi.fn();
    render(<AllergenMultiSelect selected={selected} onChange={onChange} />);

    expect(screen.getByLabelText('Dairy')).toBeChecked();
    expect(screen.getByLabelText('Nuts')).toBeChecked();
    expect(screen.getByLabelText('Sesame')).toBeChecked();

    // Others should not be checked
    expect(screen.getByLabelText('Gluten')).not.toBeChecked();
    expect(screen.getByLabelText('Fish')).not.toBeChecked();
  });
});

describe('HealthScoreInput', () => {
  it('renders with null value (empty input)', () => {
    const onChange = vi.fn();
    render(<HealthScoreInput value={null} onChange={onChange} />);

    const input = screen.getByLabelText(/health score/i) as HTMLInputElement;
    expect(input.value).toBe('');
  });

  it('calls onChange with number for valid input (1-5)', () => {
    const onChange = vi.fn();
    render(<HealthScoreInput value={null} onChange={onChange} />);

    const input = screen.getByLabelText(/health score/i);
    fireEvent.change(input, { target: { value: '3' } });

    expect(onChange).toHaveBeenCalledWith(3);
  });

  it('shows error message for invalid value 0', () => {
    const onChange = vi.fn();
    render(<HealthScoreInput value={null} onChange={onChange} />);

    const input = screen.getByLabelText(/health score/i);
    fireEvent.change(input, { target: { value: '0' } });

    expect(screen.getByRole('alert')).toHaveTextContent(/between 1 and 5/i);
  });

  it('shows error message for invalid value 6', () => {
    const onChange = vi.fn();
    render(<HealthScoreInput value={null} onChange={onChange} />);

    const input = screen.getByLabelText(/health score/i);
    fireEvent.change(input, { target: { value: '6' } });

    expect(screen.getByRole('alert')).toHaveTextContent(/between 1 and 5/i);
  });

  it('clear button calls onChange with null', () => {
    const onChange = vi.fn();
    render(<HealthScoreInput value={3} onChange={onChange} />);

    const clearButton = screen.getByRole('button', { name: /clear/i });
    fireEvent.click(clearButton);

    expect(onChange).toHaveBeenCalledWith(null);
  });

  it('renders with pre-populated value', () => {
    const onChange = vi.fn();
    render(<HealthScoreInput value={4} onChange={onChange} />);

    const input = screen.getByLabelText(/health score/i) as HTMLInputElement;
    expect(input.value).toBe('4');
  });
});

describe('Form pre-population with existing product data', () => {
  it('AllergenMultiSelect pre-populates with existing allergens', () => {
    const existingAllergens = ['gluten', 'dairy', 'eggs', 'peanuts'];
    const onChange = vi.fn();
    render(<AllergenMultiSelect selected={existingAllergens} onChange={onChange} />);

    expect(screen.getByLabelText('Gluten')).toBeChecked();
    expect(screen.getByLabelText('Dairy')).toBeChecked();
    expect(screen.getByLabelText('Eggs')).toBeChecked();
    expect(screen.getByLabelText('Peanuts')).toBeChecked();

    // Others not selected
    expect(screen.getByLabelText('Fish')).not.toBeChecked();
    expect(screen.getByLabelText('Soy')).not.toBeChecked();
    expect(screen.getByLabelText('Nuts')).not.toBeChecked();
  });

  it('HealthScoreInput pre-populates with existing score', () => {
    const onChange = vi.fn();
    render(<HealthScoreInput value={2} onChange={onChange} />);

    const input = screen.getByLabelText(/health score/i) as HTMLInputElement;
    expect(input.value).toBe('2');
  });
});


/**
 * Property 12: Baker form accepts all valid allergen/score combinations
 *
 * For any combination of 0–14 allergens selected from the valid set and
 * for any health score value (null or 1–5), the baker product form SHALL
 * accept the input without validation errors, and SHALL correctly
 * pre-populate when editing a product with those values.
 *
 * **Validates: Requirements 2.1, 2.2, 2.3, 2.8, 2.9**
 */
describe('Property 12: Baker form accepts all valid allergen/score combinations', () => {
  const allergenArbitrary = fc.subarray([...ALL_ALLERGENS], { minLength: 0, maxLength: 14 });
  const healthScoreArbitrary = fc.oneof(
    fc.constant(null),
    fc.integer({ min: 1, max: 5 })
  );

  it('AllergenMultiSelect accepts any valid subset of allergens without errors', () => {
    fc.assert(
      fc.property(allergenArbitrary, (allergens) => {
        const onChange = vi.fn();
        const { container, unmount } = render(
          <AllergenMultiSelect selected={allergens} onChange={onChange} />
        );

        // All selected allergens should be checked
        for (const allergen of allergens) {
          const capitalized = allergen.charAt(0).toUpperCase() + allergen.slice(1);
          const label = screen.getByLabelText(capitalized) as HTMLInputElement;
          expect(label.checked).toBe(true);
        }

        // Non-selected allergens should not be checked
        const unselected = ALL_ALLERGENS.filter((a) => !allergens.includes(a));
        for (const allergen of unselected) {
          const capitalized = allergen.charAt(0).toUpperCase() + allergen.slice(1);
          const label = screen.getByLabelText(capitalized) as HTMLInputElement;
          expect(label.checked).toBe(false);
        }

        // Total checkboxes always 14
        const checkboxes = container.querySelectorAll('input[type="checkbox"]');
        expect(checkboxes.length).toBe(14);

        unmount();
      }),
      { numRuns: 100 }
    );
  });

  it('HealthScoreInput accepts any valid score (null or 1-5) without errors', () => {
    fc.assert(
      fc.property(healthScoreArbitrary, (score) => {
        const onChange = vi.fn();
        const { container, unmount } = render(
          <HealthScoreInput value={score} onChange={onChange} />
        );

        const input = container.querySelector('input[type="number"]') as HTMLInputElement;

        // Value should be correctly rendered
        if (score === null) {
          expect(input.value).toBe('');
        } else {
          expect(input.value).toBe(String(score));
        }

        // No error message should be visible for valid values
        const alert = container.querySelector('[role="alert"]');
        expect(alert).toBeNull();

        unmount();
      }),
      { numRuns: 100 }
    );
  });

  it('combined allergen and score renders without validation errors', () => {
    fc.assert(
      fc.property(allergenArbitrary, healthScoreArbitrary, (allergens, score) => {
        const onAllergenChange = vi.fn();
        const onScoreChange = vi.fn();
        const { container, unmount } = render(
          <div>
            <AllergenMultiSelect selected={allergens} onChange={onAllergenChange} />
            <HealthScoreInput value={score} onChange={onScoreChange} />
          </div>
        );

        // No error alerts should be present
        const alerts = container.querySelectorAll('[role="alert"]');
        expect(alerts.length).toBe(0);

        // 14 checkboxes for allergens
        const checkboxes = container.querySelectorAll('input[type="checkbox"]');
        expect(checkboxes.length).toBe(14);

        // Health score input present
        const numberInput = container.querySelector('input[type="number"]') as HTMLInputElement;
        expect(numberInput).not.toBeNull();

        // Correct selected count
        const checked = container.querySelectorAll('input[type="checkbox"]:checked');
        expect(checked.length).toBe(allergens.length);

        unmount();
      }),
      { numRuns: 100 }
    );
  });
});
