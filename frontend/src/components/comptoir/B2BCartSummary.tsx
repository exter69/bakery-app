import { useState, useEffect } from 'react';
import { useI18n } from '../../i18n';
import { computePricing } from '../../api/b2b-client';
import type { B2BCartItem, B2BPricingResult } from '../../types/b2b';

interface Props {
  bakeryId: string;
  items: B2BCartItem[];
}

export function B2BCartSummary({ bakeryId, items }: Props) {
  const { t } = useI18n();
  const [pricing, setPricing] = useState<B2BPricingResult | null>(null);
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
        <>
          <table className="b2b-cart-summary__table">
            <tbody>
              <tr>
                <td>{t('comptoir.pricing.subtotalHt')}</td>
                <td>{fmt(pricing.subtotalHt)} EUR</td>
              </tr>
              {pricing.proDiscountRate > 0 && (
                <tr>
                  <td>{t('comptoir.pricing.remisePro')} ({pricing.proDiscountRate}%)</td>
                  <td>-{fmt(pricing.proDiscountAmt)} EUR</td>
                </tr>
              )}
              {pricing.volDiscountRate > 0 && (
                <tr>
                  <td>{t('comptoir.pricing.volumeDiscount')} ({pricing.volDiscountRate}%)</td>
                  <td>-{fmt(pricing.volDiscountAmt)} EUR</td>
                </tr>
              )}
              <tr>
                <td>{t('comptoir.pricing.tva')} ({pricing.tvaRate}%)</td>
                <td>{fmt(pricing.tvaAmount)} EUR</td>
              </tr>
              <tr className="b2b-cart-summary__total">
                <td>{t('comptoir.pricing.totalTtc')}</td>
                <td>{fmt(pricing.totalTtc)} EUR</td>
              </tr>
            </tbody>
          </table>
          <div className="b2b-cart-summary__tier-nudge">
            {pricing.nextTier ? (
              <p className="b2b-cart-summary__next-tier">
                {t('comptoir.pricing.currentTier', {
                  rate: String(pricing.volDiscountRate),
                })}
                {' — '}
                {t('comptoir.pricing.nextTier', {
                  rate: String(pricing.nextTier.discountPercent),
                  amount: fmt(pricing.nextTier.minMonthlySpend),
                  remaining: fmt(pricing.spendToNextTier),
                })}
              </p>
            ) : pricing.currentTier ? (
              <p className="b2b-cart-summary__max-tier">
                {t('comptoir.pricing.maxTier')}
              </p>
            ) : null}
          </div>
        </>
      )}
    </div>
  );
}
