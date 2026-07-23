import { useState, useEffect } from 'react';
import { useI18n } from '../../i18n';
import {
  getProfile,
  updateProfile,
  listSites,
  createSite,
  updateSite,
  deleteSite,
} from '../../api/b2b-client';
import type { BusinessProfile, DeliverySite } from '../../types/b2b';

export default function ComptoirProfilePage() {
  const { t } = useI18n();
  const [profile, setProfile] = useState<BusinessProfile | null>(null);
  const [sites, setSites] = useState<DeliverySite[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [profileSaved, setProfileSaved] = useState(false);
  const [editingSiteId, setEditingSiteId] = useState<string | null>(null);
  const [showNewSiteForm, setShowNewSiteForm] = useState(false);

  // Form state
  const [form, setForm] = useState({
    companyName: '',
    iban: '',
    billingEmail: '',
    billingContactName: '',
  });
  const [siteForm, setSiteForm] = useState({
    name: '',
    streetAddress: '',
    city: '',
    postalCode: '',
    country: 'BE',
    deliveryInstructions: '',
  });

  useEffect(() => {
    Promise.all([getProfile(), listSites()])
      .then(([p, s]) => {
        setProfile(p);
        setForm({
          companyName: p.companyName,
          iban: p.iban,
          billingEmail: p.billingEmail,
          billingContactName: p.billingContactName,
        });
        setSites(s);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setProfileSaved(false);
    try {
      const updated = await updateProfile(form);
      setProfile(updated);
      setProfileSaved(true);
    } catch {
      // Error
    } finally {
      setSaving(false);
    }
  };

  const handleCreateSite = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const newSite = await createSite(siteForm);
      setSites((prev) => [...prev, newSite]);
      setShowNewSiteForm(false);
      setSiteForm({ name: '', streetAddress: '', city: '', postalCode: '', country: 'BE', deliveryInstructions: '' });
    } catch {
      // Error
    }
  };

  const handleUpdateSite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingSiteId) return;
    try {
      const updated = await updateSite(editingSiteId, siteForm);
      setSites((prev) => prev.map((s) => s.id === editingSiteId ? updated : s));
      setEditingSiteId(null);
      setSiteForm({ name: '', streetAddress: '', city: '', postalCode: '', country: 'BE', deliveryInstructions: '' });
    } catch {
      // Error
    }
  };

  const handleDeleteSite = async (id: string) => {
    try {
      await deleteSite(id);
      setSites((prev) => prev.filter((s) => s.id !== id));
    } catch {
      // Error — probably last site
    }
  };

  const startEditSite = (site: DeliverySite) => {
    setEditingSiteId(site.id);
    setSiteForm({
      name: site.name,
      streetAddress: site.streetAddress,
      city: site.city,
      postalCode: site.postalCode,
      country: site.country,
      deliveryInstructions: site.deliveryInstructions ?? '',
    });
  };

  if (loading) return <p>{t('comptoir.common.loading')}</p>;
  if (!profile) return <p>{t('comptoir.error.generic')}</p>;

  return (
    <div className="comptoir-profile-page">
      <h1>{t('comptoir.nav.profile')}</h1>

      {/* Profile Form */}
      <section className="comptoir-profile-page__section">
        <form onSubmit={handleSaveProfile}>
          <label>
            {t('comptoir.profile.vatSiret')}
            <input type="text" value={profile.vatSiret} disabled readOnly />
          </label>
          <label>
            {t('comptoir.profile.companyName')}
            <input
              type="text"
              value={form.companyName}
              onChange={(e) => setForm((f) => ({ ...f, companyName: e.target.value }))}
              required
            />
          </label>
          <label>
            {t('comptoir.profile.iban')}
            <input
              type="text"
              value={form.iban}
              onChange={(e) => setForm((f) => ({ ...f, iban: e.target.value }))}
              required
            />
          </label>
          <label>
            {t('comptoir.profile.billingEmail')}
            <input
              type="email"
              value={form.billingEmail}
              onChange={(e) => setForm((f) => ({ ...f, billingEmail: e.target.value }))}
              required
            />
          </label>
          <label>
            {t('comptoir.profile.billingContact')}
            <input
              type="text"
              value={form.billingContactName}
              onChange={(e) => setForm((f) => ({ ...f, billingContactName: e.target.value }))}
              required
            />
          </label>
          <button type="submit" disabled={saving}>
            {saving ? t('comptoir.common.loading') : t('comptoir.profile.save')}
          </button>
          {profileSaved && <p>{t('comptoir.profile.saved')}</p>}
        </form>
      </section>

      {/* Delivery Sites */}
      <section className="comptoir-profile-page__section">
        <h2>{t('comptoir.profile.sites')}</h2>
        <ul className="comptoir-profile-page__sites">
          {sites.map((site) => (
            <li key={site.id}>
              <div>
                <strong>{site.name}</strong>
                <span>{site.streetAddress}, {site.postalCode} {site.city}</span>
              </div>
              <div className="comptoir-profile-page__site-actions">
                <button type="button" onClick={() => startEditSite(site)}>
                  {t('comptoir.profile.editSite')}
                </button>
                <button type="button" onClick={() => handleDeleteSite(site.id)}>
                  {t('comptoir.profile.deleteSite')}
                </button>
              </div>
            </li>
          ))}
        </ul>

        {editingSiteId && (
          <form onSubmit={handleUpdateSite} className="comptoir-profile-page__site-form">
            <h3>{t('comptoir.profile.editSite')}</h3>
            <input type="text" value={siteForm.name} onChange={(e) => setSiteForm((f) => ({ ...f, name: e.target.value }))} placeholder="Name" required />
            <input type="text" value={siteForm.streetAddress} onChange={(e) => setSiteForm((f) => ({ ...f, streetAddress: e.target.value }))} placeholder="Street" required />
            <input type="text" value={siteForm.city} onChange={(e) => setSiteForm((f) => ({ ...f, city: e.target.value }))} placeholder="City" required />
            <input type="text" value={siteForm.postalCode} onChange={(e) => setSiteForm((f) => ({ ...f, postalCode: e.target.value }))} placeholder="Postal code" required />
            <input type="text" value={siteForm.country} onChange={(e) => setSiteForm((f) => ({ ...f, country: e.target.value }))} placeholder="Country" required />
            <textarea value={siteForm.deliveryInstructions} onChange={(e) => setSiteForm((f) => ({ ...f, deliveryInstructions: e.target.value }))} placeholder="Instructions" />
            <div>
              <button type="submit">{t('comptoir.common.save')}</button>
              <button type="button" onClick={() => setEditingSiteId(null)}>{t('comptoir.common.cancel')}</button>
            </div>
          </form>
        )}

        {!editingSiteId && (
          showNewSiteForm ? (
            <form onSubmit={handleCreateSite} className="comptoir-profile-page__site-form">
              <h3>{t('comptoir.profile.addSite')}</h3>
              <input type="text" value={siteForm.name} onChange={(e) => setSiteForm((f) => ({ ...f, name: e.target.value }))} placeholder="Name" required />
              <input type="text" value={siteForm.streetAddress} onChange={(e) => setSiteForm((f) => ({ ...f, streetAddress: e.target.value }))} placeholder="Street" required />
              <input type="text" value={siteForm.city} onChange={(e) => setSiteForm((f) => ({ ...f, city: e.target.value }))} placeholder="City" required />
              <input type="text" value={siteForm.postalCode} onChange={(e) => setSiteForm((f) => ({ ...f, postalCode: e.target.value }))} placeholder="Postal code" required />
              <input type="text" value={siteForm.country} onChange={(e) => setSiteForm((f) => ({ ...f, country: e.target.value }))} placeholder="Country" required />
              <textarea value={siteForm.deliveryInstructions} onChange={(e) => setSiteForm((f) => ({ ...f, deliveryInstructions: e.target.value }))} placeholder="Instructions" />
              <div>
                <button type="submit">{t('comptoir.common.save')}</button>
                <button type="button" onClick={() => setShowNewSiteForm(false)}>{t('comptoir.common.cancel')}</button>
              </div>
            </form>
          ) : (
            <button type="button" onClick={() => setShowNewSiteForm(true)}>
              {t('comptoir.profile.addSite')}
            </button>
          )
        )}
      </section>
    </div>
  );
}
