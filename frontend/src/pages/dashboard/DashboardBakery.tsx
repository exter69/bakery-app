import { useState, useEffect } from 'react';
import { fetchMyBakery, updateBakery } from '../../api/seller';
import type { Bakery } from '../../types/bakery';
import './Dashboard.css';

export default function DashboardBakery() {
  const [bakery, setBakery] = useState<Bakery | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [address, setAddress] = useState('');
  const [photoUrl, setPhotoUrl] = useState('');

  useEffect(() => {
    fetchMyBakery()
      .then((b) => {
        if (b) {
          setBakery(b);
          setName(b.name);
          setDescription(b.description);
          setAddress(b.address);
          setPhotoUrl(b.photoUrl);
        }
      })
      .catch(() => setMsg({ type: 'error', text: 'Failed to load bakery information.' }))
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!bakery) return;
    setSaving(true);
    setMsg(null);
    try {
      const updated = await updateBakery(bakery.id, { name, description, address, photoUrl });
      setBakery(updated);
      setMsg({ type: 'success', text: 'Bakery updated successfully.' });
    } catch {
      setMsg({ type: 'error', text: 'Failed to save changes. Please try again.' });
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="dash-loading">Loading bakery info…</div>;

  if (!bakery) {
    return (
      <div className="dash-empty">
        <h1 className="dash-page__title">My Bakery</h1>
        <p style={{ marginTop: '1rem' }}>No bakery found. Please contact support to set up your bakery.</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="dash-page__title">My Bakery</h1>
      <p className="dash-page__subtitle">Update your bakery's public information.</p>

      {msg && <div className={`dash-msg dash-msg--${msg.type}`}>{msg.text}</div>}

      <div className="dash-card">
        <form className="dash-form" onSubmit={handleSave}>
          <div className="dash-form__field">
            <label className="dash-form__label" htmlFor="bakery-name">Name</label>
            <input
              id="bakery-name"
              className="dash-form__input"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="dash-form__field">
            <label className="dash-form__label" htmlFor="bakery-desc">Description</label>
            <textarea
              id="bakery-desc"
              className="dash-form__textarea"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>

          <div className="dash-form__field">
            <label className="dash-form__label" htmlFor="bakery-address">Address</label>
            <input
              id="bakery-address"
              className="dash-form__input"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
            />
          </div>

          <div className="dash-form__field">
            <label className="dash-form__label" htmlFor="bakery-photo">Photo URL</label>
            <input
              id="bakery-photo"
              className="dash-form__input"
              value={photoUrl}
              onChange={(e) => setPhotoUrl(e.target.value)}
              placeholder="https://..."
            />
          </div>

          <button type="submit" className="dash-btn dash-btn--primary" disabled={saving}>
            {saving ? 'Saving…' : 'Save Changes'}
          </button>
        </form>
      </div>
    </div>
  );
}
