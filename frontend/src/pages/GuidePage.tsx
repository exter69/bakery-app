import { useState } from 'react';
import { useI18n } from '../i18n';
import './GuidePage.css';

interface FlowStep {
  label: string;
}

function FlowDiagram({ steps }: { steps: FlowStep[] }) {
  return (
    <div className="guide-flow">
      {steps.map((step, i) => (
        <div className="guide-flow__step" key={i}>
          {i > 0 && <span className="guide-flow__connector" />}
          <span className="guide-flow__circle">{i + 1}</span>
          <span className="guide-flow__label">{step.label}</span>
        </div>
      ))}
    </div>
  );
}

function Section({
  icon,
  title,
  children,
}: {
  icon: string;
  title: string;
  children: React.ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <div className="guide-section">
      <div
        className="guide-section__header"
        onClick={() => setOpen(!open)}
        role="button"
        aria-expanded={open}
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            setOpen(!open);
          }
        }}
      >
        <span className="guide-section__icon">{icon}</span>
        <h3 className="guide-section__title">{title}</h3>
        <span className={`guide-section__chevron${open ? ' guide-section__chevron--open' : ''}`}>
          ▶
        </span>
      </div>
      {open && <div className="guide-section__body">{children}</div>}
    </div>
  );
}

export default function GuidePage() {
  const { t } = useI18n();
  const [tab, setTab] = useState<'customer' | 'baker'>('customer');

  return (
    <div className="guide-page">
      <h1 className="guide-page__title">{t('guide.title')}</h1>
      <p className="guide-page__subtitle">{t('guide.subtitle')}</p>

      <div className="guide-tabs">
        <button
          className={`guide-tabs__btn${tab === 'customer' ? ' guide-tabs__btn--active' : ''}`}
          onClick={() => setTab('customer')}
        >
          🛒 {t('guide.customerTab')}
        </button>
        <button
          className={`guide-tabs__btn${tab === 'baker' ? ' guide-tabs__btn--active' : ''}`}
          onClick={() => setTab('baker')}
        >
          🏪 {t('guide.bakerTab')}
        </button>
      </div>

      {tab === 'customer' && (
        <div>
          <Section icon="🗺️" title={t('guide.customer.browse')}>
            <p>{t('guide.customer.browseDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.customer.browseStep1') },
                { label: t('guide.customer.browseStep2') },
                { label: t('guide.customer.browseStep3') },
              ]}
            />
          </Section>

          <Section icon="🥐" title={t('guide.customer.viewBakery')}>
            <p>{t('guide.customer.viewBakeryDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.customer.viewStep1') },
                { label: t('guide.customer.viewStep2') },
                { label: t('guide.customer.viewStep3') },
              ]}
            />
          </Section>

          <Section icon="🚚" title={t('guide.customer.delivery')}>
            <p>{t('guide.customer.deliveryDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.customer.deliveryStep1') },
                { label: t('guide.customer.deliveryStep2') },
                { label: t('guide.customer.deliveryStep3') },
                { label: t('guide.customer.deliveryStep4') },
                { label: t('guide.customer.deliveryStep5') },
              ]}
            />
          </Section>

          <Section icon="📦" title={t('guide.customer.reservation')}>
            <p>{t('guide.customer.reservationDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.customer.reservationStep1') },
                { label: t('guide.customer.reservationStep2') },
                { label: t('guide.customer.reservationStep3') },
                { label: t('guide.customer.reservationStep4') },
              ]}
            />
          </Section>

          <Section icon="📋" title={t('guide.customer.manage')}>
            <p>{t('guide.customer.manageDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.customer.manageStep1') },
                { label: t('guide.customer.manageStep2') },
                { label: t('guide.customer.manageStep3') },
              ]}
            />
          </Section>

          <Section icon="🔄" title={t('guide.customer.recurring')}>
            <p>{t('guide.customer.recurringDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.customer.recurringStep1') },
                { label: t('guide.customer.recurringStep2') },
                { label: t('guide.customer.recurringStep3') },
              ]}
            />
          </Section>

          <Section icon="⚡" title={t('guide.customer.quickReserve')}>
            <p>{t('guide.customer.quickReserveDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.customer.quickStep1') },
                { label: t('guide.customer.quickStep2') },
                { label: t('guide.customer.quickStep3') },
              ]}
            />
          </Section>
        </div>
      )}

      {tab === 'baker' && (
        <div>
          <Section icon="📝" title={t('guide.baker.register')}>
            <p>{t('guide.baker.registerDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.baker.registerStep1') },
                { label: t('guide.baker.registerStep2') },
                { label: t('guide.baker.registerStep3') },
              ]}
            />
          </Section>

          <Section icon="🏠" title={t('guide.baker.setup')}>
            <p>{t('guide.baker.setupDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.baker.setupStep1') },
                { label: t('guide.baker.setupStep2') },
                { label: t('guide.baker.setupStep3') },
              ]}
            />
          </Section>

          <Section icon="🕐" title={t('guide.baker.schedule')}>
            <p>{t('guide.baker.scheduleDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.baker.scheduleStep1') },
                { label: t('guide.baker.scheduleStep2') },
                { label: t('guide.baker.scheduleStep3') },
              ]}
            />
          </Section>

          <Section icon="🍞" title={t('guide.baker.products')}>
            <p>{t('guide.baker.productsDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.baker.productsStep1') },
                { label: t('guide.baker.productsStep2') },
                { label: t('guide.baker.productsStep3') },
              ]}
            />
          </Section>

          <Section icon="📦" title={t('guide.baker.orders')}>
            <p>{t('guide.baker.ordersDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.baker.ordersStep1') },
                { label: t('guide.baker.ordersStep2') },
                { label: t('guide.baker.ordersStep3') },
                { label: t('guide.baker.ordersStep4') },
              ]}
            />
          </Section>

          <Section icon="🎫" title={t('guide.baker.reservations')}>
            <p>{t('guide.baker.reservationsDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.baker.reservationsStep1') },
                { label: t('guide.baker.reservationsStep2') },
                { label: t('guide.baker.reservationsStep3') },
              ]}
            />
          </Section>

          <Section icon="📊" title={t('guide.baker.analytics')}>
            <p>{t('guide.baker.analyticsDesc')}</p>
            <FlowDiagram
              steps={[
                { label: t('guide.baker.analyticsStep1') },
                { label: t('guide.baker.analyticsStep2') },
                { label: t('guide.baker.analyticsStep3') },
              ]}
            />
          </Section>
        </div>
      )}
    </div>
  );
}
