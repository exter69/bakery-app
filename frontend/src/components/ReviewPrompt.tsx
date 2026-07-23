import { useState } from 'react';
import { createReview } from '../api/reviews';
import { useI18n } from '../i18n';
import StarRating from './StarRating';
import './ReviewPrompt.css';

interface ReviewPromptProps {
  bakeryId: string;
  onClose: () => void;
  onSubmitted: () => void;
}

const DISMISS_KEY = 'review_prompt_dismissed';

export function isReviewDismissed(bakeryId: string): boolean {
  return sessionStorage.getItem(`${DISMISS_KEY}_${bakeryId}`) === 'true';
}

export function dismissReviewPrompt(bakeryId: string): void {
  sessionStorage.setItem(`${DISMISS_KEY}_${bakeryId}`, 'true');
}

export default function ReviewPrompt({ bakeryId, onClose, onSubmitted }: ReviewPromptProps) {
  const { t } = useI18n();
  const [rating, setRating] = useState(0);
  const [text, setText] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [showThankYou, setShowThankYou] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDismiss = () => {
    dismissReviewPrompt(bakeryId);
    onClose();
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      handleDismiss();
    }
  };

  const handleSubmit = async () => {
    if (rating === 0) return;
    setSubmitting(true);
    setError(null);
    try {
      await createReview(bakeryId, {
        rating,
        text: text.trim() || undefined,
      });
      setShowThankYou(true);
      setTimeout(() => {
        dismissReviewPrompt(bakeryId);
        onSubmitted();
      }, 1500);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit review');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="review-prompt-backdrop" onClick={handleBackdropClick}>
      <div
        className="review-prompt"
        role="dialog"
        aria-modal="true"
        aria-label={t('reviews.writeReview')}
      >
        <button
          type="button"
          className="review-prompt__close"
          onClick={handleDismiss}
          aria-label="Close"
        >
          <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
            <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
        </button>

        {showThankYou ? (
          <div className="review-prompt__thank-you">
            <p>{t('reviews.thankYou')}</p>
          </div>
        ) : (
          <>
            <h2 className="review-prompt__title">{t('reviews.writeReview')}</h2>

            <label className="review-prompt__label">{t('reviews.ratingLabel')}</label>
            <div className="review-prompt__stars">
              <StarRating rating={rating} onChange={setRating} size="lg" />
            </div>

            <textarea
              className="review-prompt__textarea"
              placeholder={t('reviews.textPlaceholder')}
              value={text}
              onChange={(e) => setText(e.target.value)}
              maxLength={1000}
              rows={4}
            />

            {error && (
              <p className="review-prompt__error" role="alert">{error}</p>
            )}

            <button
              type="button"
              className="review-prompt__submit"
              onClick={handleSubmit}
              disabled={rating === 0 || submitting}
            >
              {submitting ? '...' : t('reviews.submit')}
            </button>
          </>
        )}
      </div>
    </div>
  );
}
