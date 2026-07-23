import { useId, useState } from 'react';
import './StarRating.css';

interface StarRatingProps {
  rating: number;
  onChange?: (rating: number) => void;
  size?: 'sm' | 'md' | 'lg';
}

const STAR_PATH =
  'M12 2l3.09 6.26L22 9.27l-5 4.87L18.18 21 12 17.77 5.82 21 7 14.14l-5-4.87 6.91-1.01L12 2z';

function StarIcon({ filled, half, clipId }: { filled: boolean; half?: boolean; clipId?: string }) {
  if (half) {
    return (
      <svg viewBox="0 0 24 24" className="star-rating__icon" aria-hidden="true">
        <defs>
          <clipPath id={clipId}>
            <rect x="0" y="0" width="12" height="24" />
          </clipPath>
        </defs>
        <path
          d={STAR_PATH}
          fill="#d4d4d4"
          stroke="none"
        />
        <path
          d={STAR_PATH}
          fill="var(--accent, #e8b04b)"
          stroke="none"
          clipPath={`url(#${clipId})`}
        />
      </svg>
    );
  }

  return (
    <svg viewBox="0 0 24 24" className="star-rating__icon" aria-hidden="true">
      <path
        d={STAR_PATH}
        fill={filled ? 'var(--accent, #e8b04b)' : '#d4d4d4'}
        stroke="none"
      />
    </svg>
  );
}

export default function StarRating({ rating, onChange, size = 'md' }: StarRatingProps) {
  const [hovered, setHovered] = useState<number | null>(null);
  const baseId = useId();
  const isInteractive = typeof onChange === 'function';

  // Round to nearest 0.5 for display mode
  const displayRating = isInteractive
    ? (hovered ?? rating)
    : Math.round(rating * 2) / 2;

  const ariaLabel = `Rating: ${Math.round(rating * 10) / 10} out of 5`;

  return (
    <div
      className={`star-rating star-rating--${size}`}
      role="group"
      aria-label={ariaLabel}
      onMouseLeave={isInteractive ? () => setHovered(null) : undefined}
    >
      {[1, 2, 3, 4, 5].map((star) => {
        const filled = star <= Math.floor(displayRating);
        const half = !filled && star === Math.floor(displayRating) + 1 && displayRating % 1 >= 0.5;
        const clipId = `${baseId}-half-${star}`;

        if (isInteractive) {
          return (
            <button
              key={star}
              type="button"
              className={`star-rating__star star-rating__star--interactive${star <= (hovered ?? rating) ? ' star-rating__star--filled' : ''}`}
              onClick={() => onChange(star)}
              onMouseEnter={() => setHovered(star)}
              aria-label={`Rate ${star} out of 5`}
            >
              <StarIcon filled={star <= (hovered ?? rating)} />
            </button>
          );
        }

        return (
          <span key={star} className="star-rating__star">
            <StarIcon filled={filled} half={half} clipId={clipId} />
          </span>
        );
      })}
    </div>
  );
}
