import { useState } from 'react';
import { useI18n } from '../../i18n';

interface RecurrenceTemplate {
  id: string;
  bakeryName: string;
  frequency: 'daily' | 'weekly' | 'custom';
  itemCount: number;
  estimatedTotal: number;
  active: boolean;
}

export default function RecurrencesPage() {
  const { t } = useI18n();
  const [templates, setTemplates] = useState<RecurrenceTemplate[]>([]);
  const [showForm, setShowForm] = useState(false);

  // Placeholder — actual API integration would go here when the backend
  // recurring template endpoint is implemented (follow-up scope per design).

  const handleToggleActive = (id: string) => {
    setTemplates((prev) =>
      prev.map((tpl) => tpl.id === id ? { ...tpl, active: !tpl.active } : tpl)
    );
  };

  return (
    <div className="recurrences-page">
      <div className="recurrences-page__header">
        <h1>{t('comptoir.recurrences.title')}</h1>
        <button
          type="button"
          className="recurrences-page__new-btn"
          onClick={() => setShowForm(!showForm)}
        >
          {t('comptoir.recurrences.new')}
        </button>
      </div>

      {showForm && (
        <div className="recurrences-page__form">
          <p>{t('comptoir.recurrences.frequency')}</p>
          <div className="recurrences-page__form-actions">
            <button type="button" onClick={() => setShowForm(false)}>
              {t('comptoir.common.cancel')}
            </button>
          </div>
        </div>
      )}

      {templates.length === 0 ? (
        <p className="recurrences-page__empty">{t('comptoir.recurrences.empty')}</p>
      ) : (
        <table className="recurrences-page__table">
          <thead>
            <tr>
              <th>{t('comptoir.invoices.bakery')}</th>
              <th>{t('comptoir.recurrences.frequency')}</th>
              <th>{t('comptoir.commander.quantity')}</th>
              <th>{t('comptoir.pricing.totalTtc')}</th>
              <th>{t('comptoir.invoices.status')}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {templates.map((tpl) => (
              <tr key={tpl.id}>
                <td>{tpl.bakeryName}</td>
                <td>{t(`comptoir.recurrences.${tpl.frequency}`)}</td>
                <td>{tpl.itemCount}</td>
                <td>{(tpl.estimatedTotal / 100).toFixed(2)} EUR</td>
                <td>
                  <span className={`recurrences-page__badge ${tpl.active ? 'recurrences-page__badge--active' : ''}`}>
                    {tpl.active ? t('comptoir.recurrences.active') : t('comptoir.recurrences.inactive')}
                  </span>
                </td>
                <td>
                  <button type="button" onClick={() => handleToggleActive(tpl.id)}>
                    {tpl.active ? t('comptoir.recurrences.deactivate') : t('comptoir.recurrences.activate')}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
