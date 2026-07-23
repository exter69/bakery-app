import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiFetch, clearToken, API_BASE, getToken } from '../api/client';
import { useI18n } from '../i18n';
import './AccountSettingsPage.css';

export default function AccountSettingsPage() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleExportData() {
    setError(null);
    setExporting(true);
    try {
      const token = getToken();
      const response = await fetch(`${API_BASE}/user/data-export`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Export failed');
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'my-data-export.json';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch {
      setError('Failed to export data. Please try again.');
    } finally {
      setExporting(false);
    }
  }

  async function handleDeleteAccount() {
    setError(null);
    setDeleting(true);
    try {
      await apiFetch('/user/account', { method: 'DELETE' });
      clearToken();
      navigate('/login', { replace: true });
    } catch {
      setError('Failed to delete account. Please try again.');
    } finally {
      setDeleting(false);
    }
  }

  return (
    <div className="account-settings">
      <h1 className="account-settings__title">{t('settings.title')}</h1>

      {error && (
        <div className="account-settings__error" role="alert">
          {error}
        </div>
      )}

      <section className="account-settings__section">
        <h2 className="account-settings__section-title">{t('settings.dataSection')}</h2>
        <p className="account-settings__description">{t('settings.exportDescription')}</p>
        <button
          className="account-settings__btn account-settings__btn--export"
          onClick={handleExportData}
          disabled={exporting}
        >
          {exporting ? '...' : t('settings.exportData')}
        </button>
      </section>

      <section className="account-settings__section account-settings__section--danger">
        <h2 className="account-settings__section-title">{t('settings.dangerZone')}</h2>
        <p className="account-settings__description">{t('settings.deleteDescription')}</p>

        {!showDeleteConfirm ? (
          <button
            className="account-settings__btn account-settings__btn--delete"
            onClick={() => setShowDeleteConfirm(true)}
          >
            {t('settings.deleteAccount')}
          </button>
        ) : (
          <div className="account-settings__confirm">
            <p className="account-settings__confirm-text">{t('settings.deleteConfirm')}</p>
            <div className="account-settings__confirm-actions">
              <button
                className="account-settings__btn account-settings__btn--cancel"
                onClick={() => setShowDeleteConfirm(false)}
              >
                {t('settings.cancel')}
              </button>
              <button
                className="account-settings__btn account-settings__btn--delete-confirm"
                onClick={handleDeleteAccount}
                disabled={deleting}
              >
                {deleting ? '...' : t('settings.confirmDelete')}
              </button>
            </div>
          </div>
        )}
      </section>
    </div>
  );
}
