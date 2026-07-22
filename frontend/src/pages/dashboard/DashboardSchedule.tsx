import { useState, useEffect } from 'react';
import { fetchMyBakery, updateSchedule } from '../../api/seller';
import type { DaySchedule } from '../../types/bakery';
import './Dashboard.css';

const DAYS = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];

function defaultSchedule(): DaySchedule[] {
  return DAYS.map((day) => ({
    day,
    openTime: { hour: 8, minute: 0 },
    closeTime: { hour: 18, minute: 0 },
    isOpen: false,
  }));
}

function timeToString(t: { hour: number; minute: number }): string {
  return `${String(t.hour).padStart(2, '0')}:${String(t.minute).padStart(2, '0')}`;
}

function parseTime(val: string): { hour: number; minute: number } {
  const [h, m] = val.split(':').map(Number);
  return { hour: h || 0, minute: m || 0 };
}

function timeInMinutes(t: { hour: number; minute: number }): number {
  return t.hour * 60 + t.minute;
}

export default function DashboardSchedule() {
  const [bakeryId, setBakeryId] = useState<string | null>(null);
  const [schedule, setSchedule] = useState<DaySchedule[]>(defaultSchedule());
  const [useGoogleMaps, setUseGoogleMaps] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    fetchMyBakery()
      .then((b) => {
        if (b) {
          setBakeryId(b.id);
          if (b.schedule && b.schedule.length > 0) {
            setSchedule(b.schedule);
          }
          // Auto-enable Google Maps if a place ID is linked
          const placeId = b.googlePlaceId;
          if (placeId) {
            setUseGoogleMaps(true);
          }
        }
      })
      .catch(() => setMsg({ type: 'error', text: 'Failed to load schedule.' }))
      .finally(() => setLoading(false));
  }, []);

  const updateDay = (idx: number, updates: Partial<DaySchedule>) => {
    setSchedule((prev) => prev.map((d, i) => (i === idx ? { ...d, ...updates } : d)));
  };

  const validate = (): string | null => {
    for (const day of schedule) {
      if (day.isOpen) {
        if (timeInMinutes(day.closeTime) <= timeInMinutes(day.openTime)) {
          return `${day.day}: Close time must be after open time.`;
        }
      }
    }
    return null;
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!bakeryId) return;
    const err = validate();
    if (err) {
      setMsg({ type: 'error', text: err });
      return;
    }
    setSaving(true);
    setMsg(null);
    try {
      await updateSchedule(bakeryId, schedule);
      setMsg({ type: 'success', text: 'Schedule saved.' });
    } catch {
      setMsg({ type: 'error', text: 'Failed to save schedule.' });
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="dash-loading">Loading schedule…</div>;

  if (!bakeryId) {
    return (
      <div className="dash-empty">
        <h1 className="dash-page__title">Schedule</h1>
        <p style={{ marginTop: '1rem' }}>No bakery found.</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="dash-page__title">Schedule</h1>
      <p className="dash-page__subtitle">Set your bakery's opening hours for each day of the week.</p>

      {msg && <div className={`dash-msg dash-msg--${msg.type}`}>{msg.text}</div>}

      {/* Google Maps data source toggle */}
      <div className="dash-card" style={{ marginBottom: '1.5rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <h3 style={{ margin: '0 0 0.25rem', fontSize: '1rem', fontWeight: 600 }}>
              Use Google Maps hours
            </h3>
            <p style={{ margin: 0, fontSize: '0.85rem', color: '#64748b' }}>
              Automatically sync opening hours from your Google Maps listing. No need to update multiple tools.
            </p>
          </div>
          <label className="dash-toggle">
            <input
              type="checkbox"
              checked={useGoogleMaps}
              onChange={(e) => setUseGoogleMaps(e.target.checked)}
            />
            <span className="dash-toggle__slider" />
          </label>
        </div>

        {useGoogleMaps && (
          <div style={{ marginTop: '1rem', padding: '1rem', background: '#f0fdf4', borderRadius: '10px', border: '1px solid #bbf7d0' }}>
            <p style={{ margin: '0 0 0.75rem', fontSize: '0.85rem', color: '#166534' }}>
              ✓ Google Maps sync is active. Hours fetched from your linked profile:
            </p>
            <div className="dash-table-wrap" style={{ opacity: 0.85 }}>
              <table className="dash-table">
                <thead>
                  <tr>
                    <th>Day</th>
                    <th>Status</th>
                    <th>Hours</th>
                  </tr>
                </thead>
                <tbody>
                  {schedule.map((day) => (
                    <tr key={day.day}>
                      <td style={{ fontWeight: 500 }}>{day.day}</td>
                      <td>
                        <span style={{ fontSize: '0.8rem', padding: '2px 8px', borderRadius: '6px', background: day.isOpen ? '#dcfce7' : '#fee2e2', color: day.isOpen ? '#166534' : '#991b1b' }}>
                          {day.isOpen ? 'Open' : 'Closed'}
                        </span>
                      </td>
                      <td style={{ color: '#475569', fontSize: '0.9rem' }}>
                        {day.isOpen ? `${timeToString(day.openTime)} – ${timeToString(day.closeTime)}` : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <p style={{ margin: '0.75rem 0 0', fontSize: '0.8rem', color: '#64748b' }}>
              To change these hours, update them on your Google Maps listing.
            </p>
          </div>
        )}
      </div>

      {/* Manual schedule */}
      <div
        className="dash-card"
        style={{
          opacity: useGoogleMaps ? 0.5 : 1,
          pointerEvents: useGoogleMaps ? 'none' : 'auto',
          transition: 'opacity 0.2s ease',
        }}
      >
        {useGoogleMaps && (
          <div style={{ marginBottom: '0.75rem', fontSize: '0.85rem', color: '#64748b', fontStyle: 'italic' }}>
            Manual schedule disabled — using Google Maps data
          </div>
        )}
        <form onSubmit={handleSave}>
          <div className="dash-table-wrap">
            <table className="dash-table">
              <thead>
                <tr>
                  <th>Day</th>
                  <th>Open</th>
                  <th>From</th>
                  <th>To</th>
                </tr>
              </thead>
              <tbody>
                {schedule.map((day, idx) => (
                  <tr key={day.day}>
                    <td style={{ fontWeight: 500 }}>{day.day}</td>
                    <td>
                      <label className="dash-toggle">
                        <input
                          type="checkbox"
                          checked={day.isOpen}
                          onChange={(e) => updateDay(idx, { isOpen: e.target.checked })}
                        />
                        <span className="dash-toggle__slider" />
                      </label>
                    </td>
                    <td>
                      <input
                        type="time"
                        className="dash-form__input"
                        value={timeToString(day.openTime)}
                        onChange={(e) => updateDay(idx, { openTime: parseTime(e.target.value) })}
                        disabled={!day.isOpen}
                        required={day.isOpen}
                        style={{ maxWidth: 140 }}
                      />
                    </td>
                    <td>
                      <input
                        type="time"
                        className="dash-form__input"
                        value={timeToString(day.closeTime)}
                        onChange={(e) => updateDay(idx, { closeTime: parseTime(e.target.value) })}
                        disabled={!day.isOpen}
                        required={day.isOpen}
                        style={{ maxWidth: 140 }}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div style={{ marginTop: '1.25rem' }}>
            <button type="submit" className="dash-btn dash-btn--primary" disabled={saving || useGoogleMaps}>
              {saving ? 'Saving…' : 'Save Schedule'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
