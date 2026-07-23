import { useState, useEffect } from 'react';
import { useI18n } from '../../i18n';
import {
  listAccessRequests,
  approveAccess,
  rejectAccess,
  revokeAccess,
  getBakerConfig,
  saveBakerConfig,
} from '../../api/b2b-client';
import type { B2BAccess } from '../../types/b2b';

export default function DashboardB2BPage() {
  const { t } = useI18n();
  const [requests, setRequests] = useState<B2BAccess[]>([]);
  const [config, setConfig] = useState({
    cutoffTime: '18:00',
    deliveryWindowStart: '06:00',
    deliveryWindowEnd: '09:00',
    orderMinimum: 2000,
    proDiscount: 0,
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    Promise.all([
      listAccessRequests().catch(() => [] as B2BAccess[]),
      getBakerConfig().catch(() => null),
    ]).then(([reqs, cfg]) => {
      setRequests(reqs);
      if (cfg) {
        setConfig({
          cutoffTime: cfg.cutoffTime,
          deliveryWindowStart: cfg.deliveryWindowStart,
          deliveryWindowEnd: cfg.deliveryWindowEnd,
          orderMinimum: cfg.orderMinimum,
          proDiscount: cfg.proDiscount,
        });
      }
    }).finally(() => setLoading(false));
  }, []);

  const handleApprove = async (id: string) => {
    await approveAccess(id);
    setRequests((prev) => prev.map((r) => r.id === id ? { ...r, status: 'approved' as const } : r));
  };

  const handleReject = async (id: string) => {
    await rejectAccess(id);
    setRequests((prev) => prev.map((r) => r.id === id ? { ...r, status: 'rejected' as const } : r));
  };

  const handleRevoke = async (id: string) => {
    await revokeAccess(id);
    setRequests((prev) => prev.map((r) => r.id === id ? { ...r, status: 'revoked' as const } : r));
  };

  const handleSaveConfig = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setSaved(false);
    try {
      await saveBakerConfig(config);
      setSaved(true);
    } catch {
      // Error handling
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <p>{t('comptoir.common.loading')}</p>;

  const pendingRequests = requests.filter((r) => r.status === 'pending');
  const approvedRequests = requests.filter((r) => r.status === 'approved');

  return (
    <div className="dashboard-b2b-page">
      <h1>{t('comptoir.config.title')}</h1>

      {/* Access Requests */}
      <section className="dashboard-b2b-page__section">
        <h2>{t('comptoir.config.accessRequests')}</h2>
        {pendingRequests.length === 0 && approvedRequests.length === 0 ? (
          <p>{t('comptoir.config.noRequests')}</p>
        ) : (
          <table className="dashboard-b2b-page__table">
            <thead>
              <tr>
                <th>ID</th>
                <th>{t('comptoir.invoices.status')}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {pendingRequests.map((req) => (
                <tr key={req.id}>
                  <td>{req.businessUserId.slice(0, 8)}...</td>
                  <td>{req.status}</td>
                  <td>
                    <button type="button" onClick={() => handleApprove(req.id)}>
                      {t('comptoir.config.approve')}
                    </button>
                    <button type="button" onClick={() => handleReject(req.id)}>
                      {t('comptoir.config.reject')}
                    </button>
                  </td>
                </tr>
              ))}
              {approvedRequests.map((req) => (
                <tr key={req.id}>
                  <td>{req.businessUserId.slice(0, 8)}...</td>
                  <td>{req.status}</td>
                  <td>
                    <button type="button" onClick={() => handleRevoke(req.id)}>
                      {t('comptoir.config.revoke')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {/* B2B Config Form */}
      <section className="dashboard-b2b-page__section">
        <h2>{t('comptoir.config.title')}</h2>
        <form onSubmit={handleSaveConfig} className="dashboard-b2b-page__form">
          <label>
            {t('comptoir.config.cutoffTime')}
            <input
              type="time"
              value={config.cutoffTime}
              onChange={(e) => setConfig((c) => ({ ...c, cutoffTime: e.target.value }))}
            />
          </label>
          <label>
            {t('comptoir.config.deliveryStart')}
            <input
              type="time"
              value={config.deliveryWindowStart}
              onChange={(e) => setConfig((c) => ({ ...c, deliveryWindowStart: e.target.value }))}
            />
          </label>
          <label>
            {t('comptoir.config.deliveryEnd')}
            <input
              type="time"
              value={config.deliveryWindowEnd}
              onChange={(e) => setConfig((c) => ({ ...c, deliveryWindowEnd: e.target.value }))}
            />
          </label>
          <label>
            {t('comptoir.config.orderMinimum')}
            <input
              type="number"
              min="0"
              step="100"
              value={config.orderMinimum}
              onChange={(e) => setConfig((c) => ({ ...c, orderMinimum: parseInt(e.target.value) || 0 }))}
            />
            <span className="dashboard-b2b-page__hint">(cents)</span>
          </label>
          <label>
            {t('comptoir.config.proDiscount')}
            <input
              type="number"
              min="0"
              max="100"
              value={config.proDiscount}
              onChange={(e) => setConfig((c) => ({ ...c, proDiscount: parseInt(e.target.value) || 0 }))}
            />
          </label>
          <button type="submit" disabled={saving}>
            {saving ? t('comptoir.common.loading') : t('comptoir.config.save')}
          </button>
          {saved && <p className="dashboard-b2b-page__success">{t('comptoir.config.saved')}</p>}
        </form>
      </section>
    </div>
  );
}
