import { useI18n } from '../i18n';
import './TermsPage.css';

export default function TermsPage() {
  const { t } = useI18n();

  return (
    <div className="terms-page">
      <div className="terms-page__container">
        <p className="terms-page__disclaimer">
          [PLACEHOLDER — to be reviewed by legal counsel before launch]
        </p>

        <h1 className="terms-page__title">{t('terms.title')}</h1>
        <p className="terms-page__updated">Last updated: 2025-07-28</p>

        <section className="terms-page__section">
          <h2>1. Acceptance of Terms</h2>
          <p>
            By creating an account or using our platform, you agree to be bound by these Terms of
            Service. If you do not agree, please do not use our services.
          </p>
        </section>

        <section className="terms-page__section">
          <h2>2. Description of Services</h2>
          <p>
            Mie &amp; Beurre is a platform connecting customers with local bakeries for ordering
            delivery and making pickup reservations. We facilitate the transaction but are not the
            seller of the baked goods.
          </p>
        </section>

        <section className="terms-page__section">
          <h2>3. User Responsibilities</h2>
          <ul>
            <li>You must provide accurate information when creating your account</li>
            <li>You are responsible for maintaining the security of your credentials</li>
            <li>You agree to use the platform only for lawful purposes</li>
            <li>You will not abuse the ordering or reservation system</li>
            <li>You acknowledge that allergen information is provided by bakeries and may not be exhaustive</li>
          </ul>
        </section>

        <section className="terms-page__section">
          <h2>4. Orders and Payments</h2>
          <ul>
            <li>Prices are set by individual bakeries and may change without notice</li>
            <li>Payment for delivery orders is processed online via Stripe</li>
            <li>Reservation (pickup) orders are paid at the counter</li>
            <li>Cancellation policies depend on order status and bakery rules</li>
          </ul>
        </section>

        <section className="terms-page__section">
          <h2>5. Limitation of Liability</h2>
          <p>
            Mie &amp; Beurre acts as an intermediary platform. We are not liable for the quality,
            safety, or allergen accuracy of products sold by bakeries. Each bakery is independently
            responsible for their products and food safety compliance.
          </p>
        </section>

        <section className="terms-page__section">
          <h2>6. Account Termination</h2>
          <p>
            You may delete your account at any time from your Account Settings. We reserve the right
            to suspend or terminate accounts that violate these terms.
          </p>
        </section>

        <section className="terms-page__section">
          <h2>7. Changes to Terms</h2>
          <p>
            We may update these terms from time to time. Continued use of the platform after changes
            constitutes acceptance of the updated terms.
          </p>
        </section>

        <section className="terms-page__section">
          <h2>8. Governing Law</h2>
          <p>
            These terms are governed by the laws of the European Union and the applicable national
            law of the user's country of residence.
          </p>
        </section>

        <section className="terms-page__section">
          <h2>9. Contact</h2>
          <p>
            For questions about these terms, contact us at:{' '}
            <a href="mailto:legal@mieetbeurre.com">legal@mieetbeurre.com</a>
          </p>
        </section>
      </div>
    </div>
  );
}
