import { useEffect, useRef, useCallback } from 'react';
import ReactDOM from 'react-dom';
import { useI18n } from '../i18n';
import './AllergenInfoModal.css';

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

export interface AllergenInfoModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function AllergenInfoModal({ isOpen, onClose }: AllergenInfoModalProps) {
  const { t } = useI18n();
  const modalRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  // Manage focus on open/close
  useEffect(() => {
    if (isOpen) {
      previousFocusRef.current = document.activeElement as HTMLElement;
      document.body.style.overflow = 'hidden';
      setTimeout(() => modalRef.current?.focus(), 50);
    } else {
      document.body.style.overflow = '';
      previousFocusRef.current?.focus();
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  // Focus trap: Tab/Shift+Tab cycle within modal
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
        return;
      }

      if (e.key === 'Tab') {
        const modal = modalRef.current;
        if (!modal) return;

        const focusableElements = modal.querySelectorAll<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        const focusable = Array.from(focusableElements);
        if (focusable.length === 0) return;

        const first = focusable[0];
        const last = focusable[focusable.length - 1];

        if (e.shiftKey) {
          if (document.activeElement === first) {
            e.preventDefault();
            last.focus();
          }
        } else {
          if (document.activeElement === last) {
            e.preventDefault();
            first.focus();
          }
        }
      }
    },
    [onClose]
  );

  // Handle backdrop click
  const handleBackdropClick = useCallback(() => {
    onClose();
  }, [onClose]);

  if (!isOpen) return null;

  // Sort allergens alphabetically by translated name in active language
  const sortedAllergens = [...ALL_ALLERGENS].sort((a, b) => {
    const nameA = t(`allergen.${a}`);
    const nameB = t(`allergen.${b}`);
    return nameA.localeCompare(nameB);
  });

  const modal = (
    <div
      className="aim-backdrop"
      onClick={handleBackdropClick}
      aria-hidden="true"
    >
      <div
        ref={modalRef}
        className="aim"
        role="dialog"
        aria-modal="true"
        aria-label={t('allergenInfo.title')}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Header */}
        <div className="aim__header">
          <h2 className="aim__title">{t('allergenInfo.title')}</h2>
          <button
            type="button"
            className="aim__close-btn"
            onClick={onClose}
            aria-label="Close modal"
          >
            ×
          </button>
        </div>

        {/* Body */}
        <div className="aim__body">
          <p className="aim__intro">{t('allergenInfo.intro')}</p>
          <ul className="aim__allergen-list">
            {sortedAllergens.map((allergen) => (
              <li key={allergen} className="aim__allergen-item">
                <p className="aim__allergen-name">{t(`allergen.${allergen}`)}</p>
                <p className="aim__allergen-desc">
                  {t(`allergen.${allergen}.description`)}
                </p>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );

  return ReactDOM.createPortal(modal, document.body);
}
