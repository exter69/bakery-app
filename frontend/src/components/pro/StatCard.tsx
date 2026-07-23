import './StatCard.css';

interface StatCardProps {
  label: string;
  value: string | number;
  subtitle: string;
  badge?: { text: string; variant: 'positive' | 'neutral' | 'negative' };
}

export function StatCard({ label, value, subtitle, badge }: StatCardProps) {
  return (
    <article className="stat-card">
      <span className="stat-card__label">{label}</span>
      <span className="stat-card__value">{value}</span>
      <span className="stat-card__subtitle">{subtitle}</span>
      {badge && (
        <span className={`stat-card__badge stat-card__badge--${badge.variant}`}>
          {badge.text}
        </span>
      )}
    </article>
  );
}
