import { useI18n } from '../i18n';

export default function AboutPage() {
  const { t } = useI18n();

  return (
    <div className="about-page">
      <h1 className="about-page__title">{t('about.title')}</h1>
      <p className="about-page__text">
        {t('about.text1')}
      </p>
      <p className="about-page__text">
        {t('about.text2')}
      </p>
    </div>
  );
}
