import { useState, useEffect, useCallback } from 'react';
import { fetchPayouts, fetchConnectStatus, startOnboarding } from '../../api/payouts';
import { useI18n } from '../../i18n';
import type { Payout, ConnectStatus } from '../../api/payouts';
import './DashboardPayouts.css';
import './Dashboard.css';

/** Format cents as euro currency */
function formatAmount(cents: number): string {
  return new Intl.NumberFormat('fr-BE', {
    style: 'currency',
    currency: 'EUR',
  }).format(cents / 100);
}

/** Format ISO date string */
function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('fr-BE', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/** Map status to a CSS class for styling */
function statusClass(status: Payout['status']): string {
  switch (status) {
    case 'transferred': return 'payout-status--transferred';
    case 'pending': return 'payout-status--pending';
    case 'failed': return 'payout-status--failed';
    case 'refunded': return 'payout-status--refunded';
    default: return '';
  }
}

export default function DashboardPayouts() {
  const { t } = useI18n();
  const [connectStatus, setConnectStatus] = useState<ConnectStatus | null>(null);
  const [payouts, setPayouts] = useState<Payout[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [onboarding, setOnboarding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [status, payoutData] = await Promise.all([
        fetchConnectStatus(),
        fetchPayouts(page),
      ]);
      setConnectStatus(status);
      setPayouts(payoutData.items);
      setTotal(payoutData.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load payout data');
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleOnboard = async () => {
    setOnboarding(true);
    setError(null);
    try {
      const currentUrl = window.location.href;
      const result = await startOnboarding(currentUrl, currentUrl);
      window.location.href = result.url;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start onboarding');
      setOnboarding(false);
    }
  };

  const totalPages = Math.ceil(total / 20);

  /** Map status to a translated label */
  function statusLabel(status: Payout['status']): string {
    switch (status) {
      case 'transferred': return t('dashboard.payouts.status.transferred');
      case 'pending': return t('dashboard.payouts.status.pending');
      case 'failed': return t('dashboard.payouts.status.failed');
      case 'refunded': return t('dashboard.payouts.status.refunded');
      default: return status;
    }
  }

  if (loading) {
    return (
      <div className="dashboard-page">
        <h1 className="dashboard-page__title">{t('dashboard.payouts.title')}</h1>
        <p>{t('dashboard.payouts.loading')}</p>
      </div>
    );
  }

  return (
    <div className="dashboard-page">
      <h1 className="dashboard-page__title">{t('dashboard.payouts.title')}</h1>

      {error && (
        <div className="dashboard-alert dashboard-alert--error" role="alert">
          {error}
        </div>
      )}

      {/* Connect Status Section */}
      <section className="dashboard-card" style={{ marginBottom: '1.5rem' }}>
        <h2 className="dashboard-card__title">{t('dashboard.payouts.stripeConnect')}</h2>
        {connectStatus?.connected ? (
          <div className="connect-status">
            <span className="connect-status__badge connect-status__badge--connected">
              {t('dashboard.payouts.connected')}
            </span>
            {connectStatus.chargesEnabled && connectStatus.payoutsEnabled ? (
              <p className="connect-status__detail">
                {t('dashboard.payouts.fullySetUp')}
              </p>
            ) : (
              <div>
                <p className="connect-status__detail">
                  {t('dashboard.payouts.needsSetup')}
                </p>
                <button
                  className="btn btn--primary"
                  onClick={handleOnboard}
                  disabled={onboarding}
                >
                  {onboarding ? t('dashboard.payouts.redirecting') : t('dashboard.payouts.completeSetup')}
                </button>
              </div>
            )}
          </div>
        ) : (
          <div className="connect-status">
            <span className="connect-status__badge connect-status__badge--disconnected">
              {t('dashboard.payouts.notConnected')}
            </span>
            <p className="connect-status__detail">
              {t('dashboard.payouts.connectDesc')}
            </p>
            <button
              className="btn btn--primary"
              onClick={handleOnboard}
              disabled={onboarding}
            >
              {onboarding ? t('dashboard.payouts.redirecting') : t('dashboard.payouts.connectToStripe')}
            </button>
          </div>
        )}
      </section>

      {/* Payout History */}
      <section className="dashboard-card">
        <h2 className="dashboard-card__title">{t('dashboard.payouts.history')}</h2>
        {payouts.length === 0 ? (
          <p className="dashboard-empty">{t('dashboard.payouts.empty')}</p>
        ) : (
          <>
            <div className="dashboard-table-wrapper">
              <table className="dashboard-table">
                <thead>
                  <tr>
                    <th>{t('dashboard.payouts.date')}</th>
                    <th>{t('dashboard.payouts.order')}</th>
                    <th>{t('dashboard.payouts.amount')}</th>
                    <th>{t('dashboard.payouts.commission')}</th>
                    <th>{t('dashboard.payouts.status')}</th>
                  </tr>
                </thead>
                <tbody>
                  {payouts.map((payout) => (
                    <tr key={payout.id}>
                      <td>{formatDate(payout.createdAt)}</td>
                      <td className="payout-order-id">{payout.orderId.slice(0, 8)}...</td>
                      <td>{formatAmount(payout.amount)}</td>
                      <td>{formatAmount(payout.commission)}</td>
                      <td>
                        <span className={`payout-status ${statusClass(payout.status)}`}>
                          {statusLabel(payout.status)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="dashboard-pagination">
                <button
                  className="btn btn--secondary btn--sm"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  {t('dashboard.payouts.previous')}
                </button>
                <span className="dashboard-pagination__info">
                  {t('dashboard.payouts.pageInfo').replace('{page}', String(page)).replace('{total}', String(totalPages))}
                </span>
                <button
                  className="btn btn--secondary btn--sm"
                  disabled={page >= totalPages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  {t('dashboard.payouts.next')}
                </button>
              </div>
            )}
          </>
        )}
      </section>
    </div>
  );
}
