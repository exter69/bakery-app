import './FilterChips.css';

interface ChipOption<T extends string> {
  value: T;
  label: string;
}

interface FilterChipsProps<T extends string> {
  options: ChipOption<T>[];
  selected: T;
  onChange: (value: T) => void;
  variant?: 'default' | 'category';
}

export function FilterChips<T extends string>({
  options,
  selected,
  onChange,
  variant = 'default',
}: FilterChipsProps<T>) {
  return (
    <div className={`filter-chips filter-chips--${variant}`} role="radiogroup">
      {options.map((option) => {
        const isActive = option.value === selected;
        return (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={isActive}
            className={`filter-chip ${isActive ? 'filter-chip--active' : ''}`}
            onClick={() => onChange(option.value)}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
