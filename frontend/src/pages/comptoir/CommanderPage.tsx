import { useState, useEffect } from 'react';
import { useI18n } from '../../i18n';
import { listApprovedBakeries, getConfig, checkout } from '../../api/b2b-client';
import { useB2BCart } from '../../hooks/useB2BCart';
import { useSiteContext } from '../../components/comptoir/SiteSwitcher';
import { CommandeRapide } from '../../components/comptoir/CommandeRapide';
import { B2BCartSummary } from '../../components/comptoir/B2BCartSummary';
import type { Bakery } from '../../types/bakery';
import type { B2BConfig } from '../../types/b2b';

export default function CommanderPage() {
  const { t } = useI18n();
  const [bakeries, setBakeries] = useState<Bakery[]>([]);
  const [selectedBakeryId, setSelectedBakeryId] = useState<string>('');
  const [config, setConfig] = useState<B2BConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [ordering, setOrdering] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const cart = useB2BCart();
  const { selectedSite } = useSiteContext();

  useEffect(() => {
    listApprovedBakeries()
      .then((list) => {
        setBakeries(list);
        if (list.length > 0) setSelectedBakeryId(list[0].id);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!selectedBakeryId) return;
    getConfig(selectedBakeryId)
      .then(setConfig)
      .catch(() => setConfig(null));
  }, [selectedBakeryId]);

  const isCutoffPassed = (): boolean => {
    if (!config?.cutoffTime) return false;
    const [h, m] = config.cutoffTime.split(':').map(Number);
    const now = new Date();
    const cutoff = h * 60 + m;
    const current = now.getHours() * 60 + now.getMinutes();
    return current >= cutoff;
  };

  const groupItems = cart.getGroupItems(selectedBakeryId);
  const groupSubtotal = groupItems.reduce((sum, i) => sum + i.quantity * i.unitPrice, 0);
  const belowMinimum = config ? groupSubtotal < config.orderMinimum : false;
  const cutoffPassed = isCutoffPassed();

  const handleCheckout = async () => {
    if (!selectedSite || !selectedBakeryId || groupItems.length === 0) return;
    setError('');
    setSuccess('');
    setOrdering(true);
    try {
      await checkout({
        bakeryId: selectedBakeryId,
        deliverySiteId: selectedSite.id,
        items: groupItems.map((i) => ({ productId: i.productId, quantity: i.quantity })),
      });
      cart.clearGroup(selectedBakeryId);
      setSuccess('Order placed');
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t('comptoir.error.generic');
      setError(msg);
    } finally {
      setOrdering(false);
    }
  };

  if (loading) return <p>{t('comptoir.common.loading')}</p>;

  return (
    <div className="commander-page">
      <h1>{t('comptoir.commander.title')}</h1>

      {bakeries.length === 0 ? (
        <p>{t('comptoir.commander.selectBakery')}</p>
      ) : (
        <>
          <div className="commander-page__bakery-select">
            <select
              value={selectedBakeryId}
              onChange={(e) => setSelectedBakeryId(e.target.value)}
              aria-label={t('comptoir.commander.selectBakery')}
            >
              {bakeries.map((b) => (
                <option key={b.id} value={b.id}>{b.name}</option>
              ))}
            </select>
          </div>

          {config && (
            <div className="commander-page__config-info">
              <span>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                {t('comptoir.commander.cutoff')}: {config.cutoffTime}
              </span>
              <span>
                {t('comptoir.commander.deliveryWindow')}: {config.deliveryWindowStart} - {config.deliveryWindowEnd}
              </span>
              <span>
                {t('comptoir.commander.orderMinimum')}: {(config.orderMinimum / 100).toFixed(2)} EUR
              </span>
            </div>
          )}

          {cutoffPassed && (
            <p className="commander-page__warning">{t('comptoir.commander.cutoffPassed')}</p>
          )}

          <CommandeRapide bakeryId={selectedBakeryId} bakeryName={bakeries.find((b) => b.id === selectedBakeryId)?.name ?? ''} cart={cart} />

          {groupItems.length > 0 && (
            <>
              <B2BCartSummary bakeryId={selectedBakeryId} items={groupItems} />

              {belowMinimum && (
                <p className="commander-page__warning">{t('comptoir.commander.belowMinimum')}</p>
              )}

              {error && <p className="commander-page__error">{error}</p>}
              {success && <p className="commander-page__success">{success}</p>}

              <button
                type="button"
                className="commander-page__order-btn"
                disabled={cutoffPassed || belowMinimum || ordering || !selectedSite}
                onClick={handleCheckout}
              >
                {ordering ? t('comptoir.common.loading') : t('comptoir.commander.placeOrder')}
              </button>
            </>
          )}
        </>
      )}
    </div>
  );
}
