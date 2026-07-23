import { useI18n } from '../i18n';
import './PrivacyPage.css';

export default function PrivacyPage() {
  const { t } = useI18n();

  return (
    <div className="privacy-page">
      <div className="privacy-page__container">
        <p className="privacy-page__disclaimer">
          [PLACEHOLDER — to be reviewed by legal counsel before launch]
        </p>

        <h1 className="privacy-page__title">{t('privacy.title')}</h1>
        <p className="privacy-page__updated">Last updated: 2025-07-28</p>

        <section className="privacy-page__section">
          <h2>1. Data We Collect</h2>
          <ul>
            <li>Account information: username, email address, password (hashed), locale preference</li>
            <li>Order data: bakery, items, amounts, delivery/pickup timestamps</li>
            <li>Reservations: bakery, items, scheduled times</li>
            <li>Reviews: bakery, rating, text content</li>
            <li>Geolocation: used client-side only for sorting bakeries by proximity; not stored on our servers</li>
            <li>Social login data: provider name, provider user ID, associated email</li>
            <li>B2B profiles (business accounts): company name, VAT/SIRET, IBAN, billing details, delivery sites</li>
          </ul>
        </section>

        <section className="privacy-page__section">
          <h2>2. Purpose of Data Processing</h2>
          <ul>
            <li>To provide our bakery ordering and reservation service</li>
            <li>To authenticate your identity and secure your account</li>
            <li>To process payments and generate invoices (B2B)</li>
            <li>To send transactional notifications (order status updates)</li>
            <li>To improve our service through anonymized analytics</li>
          </ul>
        </section>

        <section className="privacy-page__section">
          <h2>3. Data Retention</h2>
          <p>
            We retain your personal data for as long as your account is active. If you delete your
            account, we anonymize your personal information while keeping order records for bakery
            accounting purposes.
          </p>
        </section>

        <section className="privacy-page__section">
          <h2>4. Your Rights (GDPR)</h2>
          <ul>
            <li><strong>Right of access:</strong> Export all your data from Account Settings</li>
            <li><strong>Right to rectification:</strong> Update your profile at any time</li>
            <li><strong>Right to erasure:</strong> Delete your account from Account Settings</li>
            <li><strong>Right to data portability:</strong> Download your data in JSON format</li>
            <li><strong>Right to object:</strong> Contact us to object to specific processing</li>
          </ul>
        </section>

        <section className="privacy-page__section">
          <h2>5. Data Processors</h2>
          <ul>
            <li>Stripe — payment processing</li>
            <li>Railway — application hosting</li>
            <li>SMTP provider — transactional email delivery</li>
            <li>Sentry — error tracking (PII scrubbed)</li>
          </ul>
        </section>

        <section className="privacy-page__section">
          <h2>6. Contact</h2>
          <p>
            For any privacy-related requests, contact us at:{' '}
            <a href="mailto:privacy@mieetbeurre.com">privacy@mieetbeurre.com</a>
          </p>
        </section>
      </div>
    </div>
  );
}
