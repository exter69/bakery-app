import { useI18n } from '../i18n';
import './AllergenIndicator.css';

interface AllergenIndicatorProps {
  allergens: string[];
  productName: string;
  onOpenModal: () => void;
}

/**
 * A small icon on the product card that communicates allergen presence.
 * Only render when product has allergens (length > 0).
 */
export function AllergenIndicator({ allergens, productName: _productName, onOpenModal }: AllergenIndicatorProps) {
  const { t } = useI18n();

  if (allergens.length === 0) {
    return null;
  }

  const translatedNames = allergens
    .map((a) => t(`allergen.${a}`))
    .join(', ');

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onOpenModal();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      e.stopPropagation();
      onOpenModal();
    }
  };

  return (
    <button
      type="button"
      className="allergen-indicator"
      aria-label={t('allergenInfo.containsAllergens')}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
    >
      <span aria-hidden="true">⚠️</span>
      <span className="allergen-indicator__tooltip" role="tooltip">
        {translatedNames}
      </span>
    </button>
  );
}
