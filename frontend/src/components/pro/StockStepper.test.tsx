import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import * as fc from 'fast-check';
import { StockStepper } from './StockStepper';

afterEach(() => cleanup());

/**
 * Property 4: Stock stepper bounds
 * For any sequence of increment/decrement operations on the rendered component,
 * the value passed to onChange SHALL always remain within [min, max] bounds.
 *
 * **Validates: Requirements 6.4, 5.10**
 */
describe('StockStepper – Property: bounds invariant', () => {
  it('value never goes below min when decrementing via rendered component', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 50 }),   // min bound
        fc.integer({ min: 0, max: 100 }),  // initial value offset above min
        fc.integer({ min: 1, max: 50 }),   // number of decrements
        (min, offset, decrementCount) => {
          let currentValue = min + offset;
          const max = currentValue + 10;
          const onChange = vi.fn((v: number) => { currentValue = v; });

          const { rerender } = render(
            <StockStepper value={currentValue} min={min} max={max} onChange={onChange} />
          );

          for (let i = 0; i < decrementCount; i++) {
            fireEvent.click(screen.getByLabelText('Diminuer'));
            if (onChange.mock.calls.length > 0) {
              currentValue = onChange.mock.calls[onChange.mock.calls.length - 1][0];
            }
            rerender(
              <StockStepper value={currentValue} min={min} max={max} onChange={onChange} />
            );
          }

          expect(currentValue).toBeGreaterThanOrEqual(min);
          cleanup();
        }
      ),
      { numRuns: 50 }
    );
  });

  it('value never goes above max when incrementing via rendered component', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 50 }),   // min bound
        fc.integer({ min: 1, max: 100 }),  // max offset above min
        fc.integer({ min: 1, max: 50 }),   // number of increments
        (min, maxOffset, incrementCount) => {
          const max = min + maxOffset;
          let currentValue = min;
          const onChange = vi.fn((v: number) => { currentValue = v; });

          const { rerender } = render(
            <StockStepper value={currentValue} min={min} max={max} onChange={onChange} />
          );

          for (let i = 0; i < incrementCount; i++) {
            fireEvent.click(screen.getByLabelText('Augmenter'));
            if (onChange.mock.calls.length > 0) {
              currentValue = onChange.mock.calls[onChange.mock.calls.length - 1][0];
            }
            rerender(
              <StockStepper value={currentValue} min={min} max={max} onChange={onChange} />
            );
          }

          expect(currentValue).toBeLessThanOrEqual(max);
          cleanup();
        }
      ),
      { numRuns: 50 }
    );
  });

  it('value stays within [min, max] for random sequences of operations on rendered component', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 20 }),    // min
        fc.integer({ min: 1, max: 30 }),    // max offset
        fc.array(fc.oneof(fc.constant('inc'), fc.constant('dec')), { minLength: 1, maxLength: 20 }),
        (min, maxOffset, ops) => {
          const max = min + maxOffset;
          let currentValue = min + Math.floor(maxOffset / 2);
          const onChange = vi.fn((v: number) => { currentValue = v; });

          const { rerender } = render(
            <StockStepper value={currentValue} min={min} max={max} onChange={onChange} />
          );

          for (const op of ops) {
            const label = op === 'inc' ? 'Augmenter' : 'Diminuer';
            fireEvent.click(screen.getByLabelText(label));
            if (onChange.mock.calls.length > 0) {
              currentValue = onChange.mock.calls[onChange.mock.calls.length - 1][0];
            }
            rerender(
              <StockStepper value={currentValue} min={min} max={max} onChange={onChange} />
            );

            expect(currentValue).toBeGreaterThanOrEqual(min);
            expect(currentValue).toBeLessThanOrEqual(max);
          }

          cleanup();
        }
      ),
      { numRuns: 50 }
    );
  });

  it('component renders bounded value and buttons respect limits', () => {
    const onChange = vi.fn();
    const { rerender } = render(
      <StockStepper value={0} min={0} max={3} onChange={onChange} />
    );

    // Decrement at min should be disabled / not call onChange
    fireEvent.click(screen.getByLabelText('Diminuer'));
    expect(onChange).not.toHaveBeenCalled();

    // Increment from 0
    fireEvent.click(screen.getByLabelText('Augmenter'));
    expect(onChange).toHaveBeenCalledWith(1);

    // Rerender at max
    onChange.mockClear();
    rerender(<StockStepper value={3} min={0} max={3} onChange={onChange} />);

    // Increment at max should be disabled / not call onChange
    fireEvent.click(screen.getByLabelText('Augmenter'));
    expect(onChange).not.toHaveBeenCalled();

    // Decrement from max
    fireEvent.click(screen.getByLabelText('Diminuer'));
    expect(onChange).toHaveBeenCalledWith(2);
  });
});
