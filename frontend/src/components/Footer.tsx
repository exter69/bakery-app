import { Link } from 'react-router-dom';
import { useI18n } from '../i18n';
import './Footer.css';

export default function Footer() {
  const { t } = useI18n();

  return (
    <footer className="site-footer">
      <div className="site-footer__inner">
        <div className="site-footer__brand">
          <Link to="/" className="site-footer__logo">Mie &amp; Beurre</Link>
          <p className="site-footer__tagline">{t('footer.tagline')}</p>
        </div>

        <div className="site-footer__links">
          <div className="site-footer__col">
            <h4 className="site-footer__col-title">{t('footer.navigate')}</h4>
            <Link to="/" className="site-footer__link">{t('footer.home')}</Link>
            <Link to="/bakeries" className="site-footer__link">{t('nav.bakeries')}</Link>
            <Link to="/about" className="site-footer__link">{t('nav.about')}</Link>
          </div>
          <div className="site-footer__col">
            <h4 className="site-footer__col-title">{t('footer.forBakers')}</h4>
            <Link to="/login" className="site-footer__link">{t('nav.signIn')}</Link>
            <Link to="/register?role=bakery" className="site-footer__link">{t('footer.registerBakery')}</Link>
          </div>
          <div className="site-footer__col">
            <h4 className="site-footer__col-title">{t('footer.contact')}</h4>
            <a href="mailto:contact@mieetbeurre.com" className="site-footer__link">contact@mieetbeurre.com</a>
            <span className="site-footer__link">Paris, France</span>
          </div>
        </div>

        <div className="site-footer__bottom">
          <p>© {new Date().getFullYear()} Mie &amp; Beurre. All rights reserved.</p>
        </div>
      </div>
    </footer>
  );
}
