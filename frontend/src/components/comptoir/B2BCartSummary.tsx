import { useState, useEffect } from 'react';
import { useI18n } from '../../i18n';
import { computePricing } from '../../api/b2b-client';
import type { B2BCartItem, B2BOrderPricing } from '../../types/b2b';

interface Props {
  bakeryId: string;
  items: B2BCartItem[];
}

export function B2BCartSummary({ bakeryId, items }: Props) {
  const { t } = useI18n();
  const [pricing, setPricing] = useState<B2BOrderPricing | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (items.length === 0) {
      setPricing(null);
      return;
    }
    setLoading(true);
    computePricing({
      bakeryId,
      items: items.map((i) => ({ productId: i.productId, quantity: i.quantity })),
    })
      .then(setPricing)
      .catch(() => setPricing(null))
      .finally(() => setLoading(false));
  }, [bakeryId, items]);

  if (!pricing && !loading) return null;

  const fmt = (cents: number) => (cents / 100).toFixed(2);

  return (
    <div className="b2b-cart-summary">
      {loading ? (
        <p>{t('comptoir.common.loading')}</p>
      ) : pricing && (
        <table className="b2b-cart-summary__table">
          <tbody>
            <tr>
              <td>{t('comptoir.pricing.subtotalHt')}</td>
              <td>{fmt(pricing.subtotalHt)} EUR</td>
            </tr>
            {pricing.discountRate > 0 && (
              <tr>
                <td>{t('comptoir.pricing.remisePro')} ({pricing.discountRate}%)</td>
                <td>-{fmt(pricing.discountAmount)} EUR</td>
              </tr>
            )}
            <tr>
              <td>{t('comptoir.pricing.tva')}</td>
              <td>{fmt(pricing.tvaAmount)} EUR</td>
            </tr>
            <tr className="b2b-cart-summary__total">
              <td>{t('comptoir.pricing.totalTtc')}</td>
              <td>{fmt(pricing.totalTtc)} EUR</td>
            </tr>
          </tbody>
        </table>
      )}
    </div>
  );
}
