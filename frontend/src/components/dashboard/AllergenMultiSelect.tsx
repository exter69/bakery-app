import type { ChangeEvent } from 'react';

const ALLERGENS = [
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

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

interface AllergenMultiSelectProps {
  selected: string[];
  onChange: (allergens: string[]) => void;
}

export default function AllergenMultiSelect({ selected, onChange }: AllergenMultiSelectProps) {
  const handleChange = (allergen: string) => (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.checked) {
      onChange([...selected, allergen]);
    } else {
      onChange(selected.filter((a) => a !== allergen));
    }
  };

  return (
    <fieldset style={{ border: 'none', margin: 0, padding: 0 }}>
      <legend
        style={{
          fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
          fontSize: '0.85rem',
          fontWeight: 600,
          color: '#475569',
          marginBottom: '0.5rem',
        }}
      >
        Allergens
      </legend>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr',
          gap: '0.5rem 1.5rem',
        }}
      >
        {ALLERGENS.map((allergen) => (
          <label
            key={allergen}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
              fontSize: '0.875rem',
              color: '#334155',
              cursor: 'pointer',
            }}
          >
            <input
              type="checkbox"
              checked={selected.includes(allergen)}
              onChange={handleChange(allergen)}
              style={{ accentColor: '#6366f1' }}
            />
            {capitalize(allergen)}
          </label>
        ))}
      </div>
    </fieldset>
  );
}
