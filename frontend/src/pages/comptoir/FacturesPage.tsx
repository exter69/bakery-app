import { useState, useEffect, useCallback } from 'react';
import { useI18n } from '../../i18n';
import { listInvoices, downloadInvoicePDF } from '../../api/b2b-client';
import type { B2BInvoice } from '../../types/b2b';

export default function FacturesPage() {
  const { t } = useI18n();
  const [invoices, setInvoices] = useState<B2BInvoice[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const result = await listInvoices(page);
      setInvoices(result.items);
      setTotal(result.total);
    } catch {
      setInvoices([]);
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const totalPages = Math.ceil(total / 20);

  const handleDownload = async (invoiceId: string, invoiceNumber: number) => {
    try {
      const resp = await downloadInvoicePDF(invoiceId);
      if (!resp.ok) return;
      const blob = await resp.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `facture-${invoiceNumber}.pdf`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch {
      // Silently fail
    }
  };

  const fmt = (cents: number) => (cents / 100).toFixed(2);

  const statusBadge = (status: B2BInvoice['paymentStatus']) => {
    const label = t(`comptoir.invoices.status.${status}`);
    return <span className={`factures-page__badge factures-page__badge--${status}`}>{label}</span>;
  };

  // Group invoices by month
  const grouped = invoices.reduce<Record<string, B2BInvoice[]>>((acc, inv) => {
    const date = new Date(inv.issuedAt);
    const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
    if (!acc[key]) acc[key] = [];
    acc[key].push(inv);
    return acc;
  }, {});

  const formatMonth = (key: string) => {
    const [year, month] = key.split('-');
    const date = new Date(Number(year), Number(month) - 1);
    return date.toLocaleDateString(undefined, { year: 'numeric', month: 'long' });
  };

  return (
    <div className="factures-page">
      <h1>{t('comptoir.invoices.title')}</h1>

      {loading ? (
        <p>{t('comptoir.common.loading')}</p>
      ) : invoices.length === 0 ? (
        <p className="factures-page__empty">{t('comptoir.invoices.empty')}</p>
      ) : (
        <>
          {Object.entries(grouped).map(([monthKey, monthInvoices]) => (
            <section key={monthKey}>
              <h2>{formatMonth(monthKey)}</h2>
              <table className="factures-page__table">
                <thead>
                  <tr>
                    <th>{t('comptoir.invoices.number')}</th>
                    <th>{t('comptoir.invoices.date')}</th>
                    <th>{t('comptoir.invoices.totalHt')}</th>
                    <th>{t('comptoir.invoices.tva')}</th>
                    <th>{t('comptoir.invoices.totalTtc')}</th>
                    <th>{t('comptoir.invoices.status')}</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {monthInvoices.map((inv) => (
                    <tr key={inv.id}>
                      <td>#{inv.invoiceNumber}</td>
                      <td>{new Date(inv.issuedAt).toLocaleDateString()}</td>
                      <td>{fmt(inv.subtotalHt)} EUR</td>
                      <td>{fmt(inv.tvaAmount)} EUR</td>
                      <td>{fmt(inv.totalTtc)} EUR</td>
                      <td>{statusBadge(inv.paymentStatus)}</td>
                      <td>
                        <button
                          type="button"
                          className="factures-page__download-btn"
                          onClick={() => handleDownload(inv.id, inv.invoiceNumber)}
                          aria-label={`${t('comptoir.invoices.download')} #${inv.invoiceNumber}`}
                        >
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                            <polyline points="7 10 12 15 17 10" />
                            <line x1="12" y1="15" x2="12" y2="3" />
                          </svg>
                          {t('comptoir.invoices.download')}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          ))}

          <div className="factures-page__pagination">
            <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              {t('comptoir.common.previous')}
            </button>
            <span>{page} / {totalPages || 1}</span>
            <button type="button" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              {t('comptoir.common.next')}
            </button>
          </div>
        </>
      )}
    </div>
  );
}
