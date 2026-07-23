import './StockStepper.css';

interface StockStepperProps {
  value: number;
  min?: number;
  max?: number;
  onChange: (newValue: number) => void;
  danger?: boolean;
}

export function StockStepper({
  value,
  min = 0,
  max,
  onChange,
  danger = false,
}: StockStepperProps) {
  const atMin = value <= min;
  const atMax = max !== undefined && value >= max;

  function decrement() {
    if (!atMin) {
      onChange(Math.max(min, value - 1));
    }
  }

  function increment() {
    if (!atMax) {
      const next = value + 1;
      onChange(max !== undefined ? Math.min(max, next) : next);
    }
  }

  return (
    <div
      className={`stock-stepper ${danger ? 'stock-stepper--danger' : ''}`}
      role="group"
      aria-label="Contrôle de stock"
    >
      <button
        type="button"
        className="stock-stepper__btn"
        onClick={decrement}
        disabled={atMin}
        aria-label="Diminuer"
      >
        −
      </button>
      <span className="stock-stepper__value" aria-live="polite">
        {value}
      </span>
      <button
        type="button"
        className="stock-stepper__btn"
        onClick={increment}
        disabled={atMax}
        aria-label="Augmenter"
      >
        +
      </button>
    </div>
  );
}
