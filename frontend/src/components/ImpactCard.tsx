import { useI18n } from '../i18n';
import { useBundleImpact } from '../hooks/useBundles';
import './ImpactCard.css';

/**
 * Community impact metrics card.
 * Shows total bundles saved and estimated weight of food waste avoided this month.
 * Does not render if no impact data is available or totalSaved is 0.
 */
export function ImpactCard() {
  const { t } = useI18n();
  const { data, loading } = useBundleImpact();

  if (loading || !data || data.totalSaved === 0) {
    return null;
  }

  const savedText = t('bundles.impact.saved').replace('{count}', String(data.totalSaved));
  const weightText = t('bundles.impact.weight').replace('{kg}', String(data.weightAvoided));

  return (
    <div className="impact-card" aria-label="Community impact">
      <p className="impact-card__saved">
        <svg className="impact-card__icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M6 3v12"/><path d="M18 9a6 6 0 0 1-6 6H6"/><path d="M18 3a6 6 0 0 0-6 6"/></svg>
        {savedText}
      </p>
      <p className="impact-card__weight">{weightText}</p>
    </div>
  );
}
