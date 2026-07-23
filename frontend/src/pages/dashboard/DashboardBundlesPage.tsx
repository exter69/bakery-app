import { useState, useEffect, useCallback } from 'react';
import { listBundles, publishBundle } from '../../api/bundles';
import { BundleForm } from '../../components/dashboard/BundleForm';
import type { Bundle, BundleStatus } from '../../types/bundle';
import '../dashboard/Dashboard.css';
import './DashboardBundlesPage.css';

const STATUS_COLORS: Record<BundleStatus, string> = {
  draft: 'dash-badge--preparing',       // grey/amber
  published: 'dash-badge--ready',       // green
  expired: 'dash-badge--delivered',     // muted
  sold_out: 'dash-badge--cancelled',    // red
};

const STATUS_LABELS: Record<BundleStatus, string> = {
  draft: 'Brouillon',
  published: 'Publié',
  expired: 'Expiré',
  sold_out: 'Épuisé',
};

export default function DashboardBundlesPage() {
  const [bundles, setBundles] = useState<Bundle[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const loadBundles = useCallback(async () => {
    try {
      const res = await listBundles();
      setBundles(res.items);
    } catch {
      setMsg({ type: 'error', text: 'Failed to load bundles.' });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadBundles();
  }, [loadBundles]);

  const handlePublish = async (id: string) => {
    try {
      await publishBundle(id);
      setMsg({ type: 'success', text: 'Bundle published!' });
      await loadBundles();
    } catch {
      setMsg({ type: 'error', text: 'Failed to publish bundle.' });
    }
  };

  const handleFormSuccess = () => {
    setShowForm(false);
    setMsg({ type: 'success', text: 'Bundle created!' });
    loadBundles();
  };

  const formatPrice = (cents: number) => `€${(cents / 100).toFixed(2)}`;

  if (loading) return <div className="dash-loading">Loading bundles…</div>;

  return (
    <div>
      <div className="dash-bundles__header">
        <div className="dash-bundles__header-text">
          <h1 className="dash-page__title">Paniers du soir</h1>
          <p className="dash-page__subtitle">Create and manage your evening surplus bundles.</p>
        </div>
        {!showForm && (
          <button
            className="dash-btn dash-btn--primary"
            onClick={() => setShowForm(true)}
          >
            + Nouveau panier
          </button>
        )}
      </div>

      {msg && <div className={`dash-msg dash-msg--${msg.type}`}>{msg.text}</div>}

      {showForm && (
        <div className="dash-card" style={{ marginBottom: '1.5rem' }}>
          <BundleForm
            onSuccess={handleFormSuccess}
            onCancel={() => setShowForm(false)}
          />
        </div>
      )}

      {bundles.length === 0 && !showForm ? (
        <div className="dash-empty">
          <p>No bundles yet. Create your first evening bundle!</p>
        </div>
      ) : (
        <div className="dash-bundles__grid">
          {bundles.map((bundle) => (
            <div key={bundle.id} className="dash-bundles__card">
              <div className="dash-bundles__card-header">
                <div>
                  <div className="dash-bundles__card-name">{bundle.name}</div>
                  <div className="dash-bundles__card-type">
                    {bundle.type === 'compose' ? 'Composé' : 'Surprise'}
                  </div>
                </div>
                <span className={`dash-badge ${STATUS_COLORS[bundle.status]}`}>
                  {STATUS_LABELS[bundle.status]}
                </span>
              </div>

              <div className="dash-bundles__card-info">
                Pickup: {bundle.pickupStartTime}–{bundle.pickupEndTime} · Stock: {bundle.quantityRemaining}/{bundle.quantityTotal}
              </div>

              <div className="dash-bundles__card-pricing">
                <span className="original">{formatPrice(bundle.originalPrice)}</span>
                <span className="discounted">{formatPrice(bundle.discountedPrice)}</span>
              </div>

              <div className="dash-bundles__card-footer">
                {bundle.status === 'draft' && (
                  <button
                    className="dash-btn dash-btn--primary dash-btn--sm"
                    onClick={() => handlePublish(bundle.id)}
                  >
                    Publier
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
