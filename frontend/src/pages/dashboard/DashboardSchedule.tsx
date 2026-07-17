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

      <div className="dash-card">
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
            <button type="submit" className="dash-btn dash-btn--primary" disabled={saving}>
              {saving ? 'Saving…' : 'Save Schedule'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
