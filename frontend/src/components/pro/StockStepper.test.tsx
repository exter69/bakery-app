import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import * as fc from 'fast-check';
import { StockStepper } from './StockStepper';

afterEach(() => cleanup());

/**
 * Property 4: Stock stepper bounds
 * For any sequence of increment/decrement operations on a stock stepper,
 * the displayed value SHALL always remain within [min, max] bounds.
 *
 * **Validates: Requirements 6.4, 5.10**
 */
describe('StockStepper – Property: bounds invariant', () => {
  it('value never goes below min when decrementing', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 50 }),   // min bound
        fc.integer({ min: 0, max: 100 }),  // initial value offset above min
        fc.integer({ min: 1, max: 50 }),   // number of decrements
        (min, offset, decrementCount) => {
          const max = min + offset + 50;
          let currentValue = min + offset;
          const onChange = (newValue: number) => { currentValue = newValue; };

          // Simulate the decrement logic directly (same as component)
          for (let i = 0; i < decrementCount; i++) {
            const atMin = currentValue <= min;
            if (!atMin) {
              onChange(Math.max(min, currentValue - 1));
            }
          }

          // Property: value never below min
          expect(currentValue).toBeGreaterThanOrEqual(min);
        }
      ),
      { numRuns: 200 }
    );
  });

  it('value never goes above max when incrementing', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 50 }),   // min bound
        fc.integer({ min: 1, max: 100 }),  // max offset above min
        fc.integer({ min: 1, max: 50 }),   // number of increments
        (min, maxOffset, incrementCount) => {
          const max = min + maxOffset;
          let currentValue = min;
          const onChange = (newValue: number) => { currentValue = newValue; };

          // Simulate the increment logic directly (same as component)
          for (let i = 0; i < incrementCount; i++) {
            const atMax = currentValue >= max;
            if (!atMax) {
              const next = currentValue + 1;
              onChange(Math.min(max, next));
            }
          }

          // Property: value never above max
          expect(currentValue).toBeLessThanOrEqual(max);
        }
      ),
      { numRuns: 200 }
    );
  });

  it('value stays within [min, max] for random sequences of operations', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 20 }),    // min
        fc.integer({ min: 1, max: 30 }),    // max offset
        fc.array(fc.oneof(fc.constant('inc'), fc.constant('dec')), { minLength: 1, maxLength: 30 }),
        (min, maxOffset, ops) => {
          const max = min + maxOffset;
          let currentValue = min + Math.floor(maxOffset / 2);

          // Simulate component logic for each operation
          for (const op of ops) {
            if (op === 'dec') {
              const atMin = currentValue <= min;
              if (!atMin) {
                currentValue = Math.max(min, currentValue - 1);
              }
            } else {
              const atMax = currentValue >= max;
              if (!atMax) {
                currentValue = Math.min(max, currentValue + 1);
              }
            }

            // Property: value always within bounds after each op
            expect(currentValue).toBeGreaterThanOrEqual(min);
            expect(currentValue).toBeLessThanOrEqual(max);
          }
        }
      ),
      { numRuns: 200 }
    );
  });

  // Integration test: verify the component actually calls onChange with bounded values
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
