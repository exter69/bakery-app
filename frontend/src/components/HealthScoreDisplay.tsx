import './HealthScoreDisplay.css';

interface HealthScoreDisplayProps {
  /** Health score value (1-5). Component should only be rendered when score is non-null. */
  score: number;
}

/**
 * Small pill/badge displaying a product's health score.
 * Uses the artisan theme styling with Patrick Hand font.
 * Parent component is responsible for conditional rendering (only when score is non-null).
 */
export default function HealthScoreDisplay({ score }: HealthScoreDisplayProps) {
  return (
    <span
      className="health-score-badge"
      aria-label={`Health score: ${score} out of 5`}
    >
      <span className="health-score-badge__emoji" aria-hidden="true">🌿</span>
      <span className="health-score-badge__value">{score}/5</span>
    </span>
  );
}
