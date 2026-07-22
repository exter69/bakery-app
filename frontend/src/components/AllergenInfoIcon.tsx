import { useState } from 'react';
import { useI18n } from '../i18n';
import AllergenInfoModal from './AllergenInfoModal';
import './AllergenInfoIcon.css';

export interface AllergenInfoIconProps {
  onOpenModal?: () => void;
}

export default function AllergenInfoIcon({ onOpenModal }: AllergenInfoIconProps) {
  const { t } = useI18n();
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleClick = () => {
    if (onOpenModal) {
      onOpenModal();
    } else {
      setIsModalOpen(true);
    }
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
  };

  return (
    <>
      <button
        type="button"
        className="allergen-info-icon"
        aria-label={t('allergenInfo.label')}
        onClick={handleClick}
      >
        <span className="allergen-info-icon__symbol" aria-hidden="true">
          🛡️
        </span>
      </button>

      {!onOpenModal && (
        <AllergenInfoModal isOpen={isModalOpen} onClose={handleCloseModal} />
      )}
    </>
  );
}
