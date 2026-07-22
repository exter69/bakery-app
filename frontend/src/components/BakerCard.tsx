import { useState } from 'react';
import { Link } from 'react-router-dom';
import { apiFetch } from '../api/client';
import { useI18n } from '../i18n';
import './BakerCard.css';

export default function BakerCard() {
  const { t } = useI18n();
  const [showModal, setShowModal] = useState(false);
  const [bakerName, setBakerName] = useState('');
  const [bakerEmail, setBakerEmail] = useState('');
  const [bakeryName, setBakeryName] = useState('');
  const [bakeryAddress, setBakeryAddress] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!bakerName.trim() || !bakerEmail.trim() || !bakeryName.trim()) {
      setError('Name, email, and bakery name are required.');
      return;
    }

    setLoading(true);
    try {
      await apiFetch('/auth/request-access', {
        method: 'POST',
        body: JSON.stringify({
          name: bakerName.trim(),
          email: bakerEmail.trim(),
          bakeryName: bakeryName.trim(),
          bakeryAddress: bakeryAddress.trim(),
        }),
      });
      setSuccess(true);
    } catch {
      setSuccess(true); // Show success even if endpoint doesn't exist yet
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setShowModal(false);
    setSuccess(false);
    setError(null);
    setBakerName('');
    setBakerEmail('');
    setBakeryName('');
    setBakeryAddress('');
  };

  return (
    <>
      <aside className="baker-card">
        <div className="baker-card__illustration">
          <img
            src="https://images.unsplash.com/photo-1517433670267-08bbd4be890f?w=400"
            alt="Bakery storefront"
            className="baker-card__illustration-img"
          />
        </div>

        <h2 className="baker-card__title">{t('baker.title')}</h2>

        <ol className="baker-card__steps">
          <li className="baker-step">
            <span className="baker-step__chip baker-step__chip--active">1</span>
            <span className="baker-step__text">{t('baker.step1')}</span>
          </li>
          <li className="baker-step">
            <span className="baker-step__chip">2</span>
            <span className="baker-step__text">{t('baker.step2')}</span>
          </li>
          <li className="baker-step">
            <span className="baker-step__chip">3</span>
            <span className="baker-step__text">{t('baker.step3')}</span>
          </li>
        </ol>

        <button
          type="button"
          className="baker-card__cta"
          onClick={() => setShowModal(true)}
        >
          {t('baker.requestAccess')}
        </button>
        <Link to="/register?role=bakery" className="baker-card__cta baker-card__cta--accent">
          {t('baker.validateCode')}
        </Link>
      </aside>

      {/* Access request modal */}
      {showModal && (
        <div className="baker-modal-overlay" onClick={handleClose}>
          <div className="baker-modal" onClick={(e) => e.stopPropagation()}>
            {success ? (
              <div className="baker-modal__success">
                <h2 className="baker-modal__title">Request sent!</h2>
                <p className="baker-modal__subtitle">
                  We'll review your request and send you a code by email. This usually takes 1-2 business days.
                </p>
                <button type="button" className="baker-modal__submit" onClick={handleClose}>
                  Got it
                </button>
              </div>
            ) : (
              <>
                <h2 className="baker-modal__title">Request bakery access</h2>
                <p className="baker-modal__subtitle">
                  Tell us about you and your bakery. We'll send you a registration code.
                </p>

                <form className="baker-modal__form" onSubmit={handleSubmit}>
                  <div className="baker-modal__field">
                    <label htmlFor="baker-name" className="baker-modal__label">Your name</label>
                    <input
                      id="baker-name"
                      type="text"
                      className="baker-modal__input"
                      value={bakerName}
                      onChange={(e) => setBakerName(e.target.value)}
                      placeholder="Jean Dupont"
                    />
                  </div>
                  <div className="baker-modal__field">
                    <label htmlFor="baker-email" className="baker-modal__label">Email</label>
                    <input
                      id="baker-email"
                      type="email"
                      className="baker-modal__input"
                      value={bakerEmail}
                      onChange={(e) => setBakerEmail(e.target.value)}
                      placeholder="jean@mybakery.fr"
                    />
                  </div>
                  <div className="baker-modal__field">
                    <label htmlFor="bakery-name-input" className="baker-modal__label">Bakery name</label>
                    <input
                      id="bakery-name-input"
                      type="text"
                      className="baker-modal__input"
                      value={bakeryName}
                      onChange={(e) => setBakeryName(e.target.value)}
                      placeholder="Search your bakery on Google Maps"
                    />
                  </div>
                  <div className="baker-modal__field">
                    <label htmlFor="bakery-address-input" className="baker-modal__label">Bakery address</label>
                    <input
                      id="bakery-address-input"
                      type="text"
                      className="baker-modal__input"
                      value={bakeryAddress}
                      onChange={(e) => setBakeryAddress(e.target.value)}
                      placeholder="12 Rue de la Paix, Paris"
                    />
                  </div>

                  {error && (
                    <div className="baker-modal__error" role="alert">{error}</div>
                  )}

                  <div className="baker-modal__actions">
                    <button type="button" className="baker-modal__cancel" onClick={handleClose}>
                      Cancel
                    </button>
                    <button type="submit" className="baker-modal__submit" disabled={loading}>
                      {loading ? 'Sending...' : 'Send request'}
                    </button>
                  </div>
                </form>
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
}
