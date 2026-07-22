import { useState } from 'react';

interface HealthScoreInputProps {
  value: number | null;
  onChange: (score: number | null) => void;
  error?: string;
}

export default function HealthScoreInput({ value, onChange, error }: HealthScoreInputProps) {
  const [localError, setLocalError] = useState<string | null>(null);

  const displayError = error || localError;

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value;

    // Allow clearing via empty input
    if (raw === '') {
      setLocalError(null);
      onChange(null);
      return;
    }

    const num = parseInt(raw, 10);

    if (isNaN(num)) {
      setLocalError('Please enter a valid number.');
      return;
    }

    if (num < 1 || num > 5) {
      setLocalError('Health score must be between 1 and 5.');
      onChange(null);
      return;
    }

    setLocalError(null);
    onChange(num);
  };

  const handleClear = () => {
    setLocalError(null);
    onChange(null);
  };

  return (
    <div className="dash-form__field">
      <label className="dash-form__label" htmlFor="health-score-input">
        Health score (1 = least healthy, 5 = healthiest)
      </label>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <input
          id="health-score-input"
          className="dash-form__input"
          type="number"
          min={1}
          max={5}
          step={1}
          value={value ?? ''}
          onChange={handleChange}
          aria-describedby={displayError ? 'health-score-error' : undefined}
          aria-invalid={displayError ? true : undefined}
          style={{ maxWidth: '120px' }}
        />
        {value !== null && (
          <button
            type="button"
            onClick={handleClear}
            style={{
              background: 'none',
              border: 'none',
              color: '#6366f1',
              fontSize: '0.8rem',
              fontWeight: 600,
              cursor: 'pointer',
              padding: '0.25rem 0.5rem',
              fontFamily: 'inherit',
            }}
          >
            Clear
          </button>
        )}
      </div>
      {displayError && (
        <span
          id="health-score-error"
          role="alert"
          style={{
            color: '#dc2626',
            fontSize: '0.8rem',
            marginTop: '0.25rem',
          }}
        >
          {displayError}
        </span>
      )}
    </div>
  );
}
