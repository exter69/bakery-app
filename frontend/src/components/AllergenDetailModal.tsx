import { useEffect, useRef, useCallback } from 'react';
import ReactDOM from 'react-dom';
import { useI18n } from '../i18n';
import './AllergenDetailModal.css';

export interface AllergenDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  productName: string;
  allergens: string[];
}

export default function AllergenDetailModal({
  isOpen,
  onClose,
  productName,
  allergens,
}: AllergenDetailModalProps) {
  const { t } = useI18n();
  const modalRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  // Store the trigger element and manage focus on open/close
  useEffect(() => {
    if (isOpen) {
      previousFocusRef.current = document.activeElement as HTMLElement;
      document.body.style.overflow = 'hidden';
      // Focus the modal after render
      setTimeout(() => modalRef.current?.focus(), 50);
    } else {
      document.body.style.overflow = '';
      // Return focus to trigger element
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
  const sortedAllergens = [...allergens].sort((a, b) => {
    const nameA = t(`allergen.${a}`);
    const nameB = t(`allergen.${b}`);
    return nameA.localeCompare(nameB);
  });

  const modal = (
    <div
      className="adm-backdrop"
      onClick={handleBackdropClick}
      aria-hidden="true"
    >
      <div
        ref={modalRef}
        className="adm"
        role="dialog"
        aria-modal="true"
        aria-label={productName}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Header */}
        <div className="adm__header">
          <h2 className="adm__title">{productName}</h2>
          <button
            type="button"
            className="adm__close-btn"
            onClick={onClose}
            aria-label="Close modal"
          >
            ×
          </button>
        </div>

        {/* Body */}
        <div className="adm__body">
          <ul className="adm__allergen-list">
            {sortedAllergens.map((allergen) => (
              <li key={allergen} className="adm__allergen-item">
                <p className="adm__allergen-name">{t(`allergen.${allergen}`)}</p>
                <p className="adm__allergen-desc">
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
