import { useCallback, useEffect, useMemo, useState } from 'react';
import type { DaySchedule, Product } from '../types/bakery';
import './OrderSidePanel.css';

/** An item selected for the order */
export interface OrderItem {
  product: Product;
  quantity: number;
}

export interface OrderSidePanelProps {
  isOpen: boolean;
  onClose: () => void;
  bakerySchedule: DaySchedule[];
  onStartSelection: () => void;
  items: OrderItem[];
  onSubmit: (day: string, startTime: string, endTime: string) => void;
  submitting?: boolean;
  submitError?: string | null;
}

const DAY_LABELS: Record<string, string> = {
  Monday: 'Mon',
  Tuesday: 'Tue',
  Wednesday: 'Wed',
  Thursday: 'Thu',
  Friday: 'Fri',
  Saturday: 'Sat',
  Sunday: 'Sun',
};

const DAY_ORDER = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];

function formatTime(t: { hour: number; minute: number }): string {
  return `${String(t.hour).padStart(2, '0')}:${String(t.minute).padStart(2, '0')}`;
}

function parseTime(value: string): { hour: number; minute: number } | null {
  const parts = value.split(':');
  if (parts.length !== 2) return null;
  const hour = parseInt(parts[0], 10);
  const minute = parseInt(parts[1], 10);
  if (isNaN(hour) || isNaN(minute)) return null;
  if (hour < 0 || hour > 23 || minute < 0 || minute > 59) return null;
  return { hour, minute };
}

function timeToMinutes(t: { hour: number; minute: number }): number {
  return t.hour * 60 + t.minute;
}

export default function OrderSidePanel({
  isOpen,
  onClose,
  bakerySchedule,
  onStartSelection,
  items,
  onSubmit,
  submitting = false,
  submitError = null,
}: OrderSidePanelProps) {
  const [selectedDay, setSelectedDay] = useState<string | null>(null);
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  // Reset state when panel closes
  useEffect(() => {
    if (!isOpen) {
      setSelectedDay(null);
      setStartTime('');
      setEndTime('');
      setValidationErrors([]);
    }
  }, [isOpen]);

  // Build schedule map for quick lookup
  const scheduleMap = useMemo(() => {
    const map = new Map<string, DaySchedule>();
    for (const ds of bakerySchedule) {
      map.set(ds.day, ds);
    }
    return map;
  }, [bakerySchedule]);

  // Get schedule for selected day
  const selectedDaySchedule = selectedDay ? scheduleMap.get(selectedDay) : null;

  // Validate time slot when inputs change
  const validateTimeSlot = useCallback(
    (day: string | null, start: string, end: string): string[] => {
      const errors: string[] = [];

      if (!day) return errors;

      const daySchedule = scheduleMap.get(day);
      if (!daySchedule) return errors;

      if (!daySchedule.isOpen) {
        errors.push(`The bakery is closed on ${day}. Please select an available day.`);
        return errors;
      }

      if (!start && !end) return errors;

      const parsedStart = parseTime(start);
      const parsedEnd = parseTime(end);

      if (start && !parsedStart) {
        errors.push('Invalid start time format. Use HH:MM.');
        return errors;
      }

      if (end && !parsedEnd) {
        errors.push('Invalid end time format. Use HH:MM.');
        return errors;
      }

      if (parsedStart && parsedEnd) {
        const startMinutes = timeToMinutes(parsedStart);
        const endMinutes = timeToMinutes(parsedEnd);
        const openMinutes = timeToMinutes(daySchedule.openTime);
        const closeMinutes = timeToMinutes(daySchedule.closeTime);

        if (startMinutes >= endMinutes) {
          errors.push('Start time must be before end time.');
        }

        if (startMinutes < openMinutes || endMinutes > closeMinutes) {
          errors.push(
            `Time must be within operating hours: ${formatTime(daySchedule.openTime)} – ${formatTime(daySchedule.closeTime)}.`
          );
        }
      }

      return errors;
    },
    [scheduleMap]
  );

  const handleDaySelect = (day: string) => {
    const daySchedule = scheduleMap.get(day);
    if (!daySchedule || !daySchedule.isOpen) {
      setSelectedDay(day);
      setValidationErrors([`The bakery is closed on ${day}. Please select an available day.`]);
      return;
    }
    setSelectedDay(day);
    setValidationErrors(validateTimeSlot(day, startTime, endTime));
  };

  const handleStartTimeChange = (value: string) => {
    setStartTime(value);
    setValidationErrors(validateTimeSlot(selectedDay, value, endTime));
  };

  const handleEndTimeChange = (value: string) => {
    setEndTime(value);
    setValidationErrors(validateTimeSlot(selectedDay, startTime, value));
  };

  const handleSubmit = () => {
    // Final validation before submit
    const errors = validateTimeSlot(selectedDay, startTime, endTime);

    if (!selectedDay) {
      errors.unshift('Please select a day.');
    }
    if (!startTime || !endTime) {
      errors.push('Please specify both start and end times.');
    }
    if (items.length === 0) {
      errors.push('Please add at least one product to your order.');
    }

    if (errors.length > 0) {
      setValidationErrors(errors);
      return;
    }

    onSubmit(selectedDay!, startTime, endTime);
  };

  const handleClose = () => {
    onClose();
  };

  // Calculate total
  const total = items.reduce((sum, item) => sum + item.quantity * item.product.price, 0);

  const isSubmitDisabled = items.length === 0 || submitting;

  return (
    <>
      {/* Backdrop */}
      <div
        className={`order-side-panel__backdrop ${isOpen ? 'order-side-panel__backdrop--open' : ''}`}
        onClick={handleClose}
        aria-hidden="true"
      />

      {/* Panel */}
      <aside
        className={`order-side-panel ${isOpen ? 'order-side-panel--open' : ''}`}
        role="dialog"
        aria-modal="true"
        aria-label="Place Order"
        aria-hidden={!isOpen}
      >
        {/* Header */}
        <div className="order-side-panel__header">
          <h2 className="order-side-panel__title">Place Order</h2>
          <button
            type="button"
            className="order-side-panel__close-btn"
            onClick={handleClose}
            aria-label="Close panel"
          >
            ×
          </button>
        </div>

        {/* Content */}
        <div className="order-side-panel__content">
          {/* Day Selector */}
          <div className="order-side-panel__section">
            <h3 className="order-side-panel__section-title">Select Day</h3>
            <div className="order-side-panel__days">
              {DAY_ORDER.map((day) => {
                const schedule = scheduleMap.get(day);
                const isClosed = !schedule || !schedule.isOpen;
                const isSelected = selectedDay === day;
                return (
                  <button
                    key={day}
                    type="button"
                    className={`order-side-panel__day-btn ${isSelected ? 'order-side-panel__day-btn--selected' : ''}`}
                    disabled={isClosed}
                    onClick={() => handleDaySelect(day)}
                    aria-label={`${day}${isClosed ? ' (closed)' : ''}`}
                    aria-pressed={isSelected}
                  >
                    <span className="order-side-panel__day-label">{DAY_LABELS[day]}</span>
                    <span className="order-side-panel__day-status">
                      {isClosed ? 'Closed' : 'Open'}
                    </span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Time Slot Selector */}
          <div className="order-side-panel__section">
            <h3 className="order-side-panel__section-title">Select Time Slot</h3>
            <div className="order-side-panel__time-inputs">
              <div className="order-side-panel__time-field">
                <label className="order-side-panel__time-label" htmlFor="order-start-time">
                  Start
                </label>
                <input
                  id="order-start-time"
                  type="time"
                  className="order-side-panel__time-input"
                  value={startTime}
                  onChange={(e) => handleStartTimeChange(e.target.value)}
                  aria-label="Order start time"
                />
              </div>
              <span className="order-side-panel__time-separator">–</span>
              <div className="order-side-panel__time-field">
                <label className="order-side-panel__time-label" htmlFor="order-end-time">
                  End
                </label>
                <input
                  id="order-end-time"
                  type="time"
                  className="order-side-panel__time-input"
                  value={endTime}
                  onChange={(e) => handleEndTimeChange(e.target.value)}
                  aria-label="Order end time"
                />
              </div>
            </div>
            {selectedDaySchedule && selectedDaySchedule.isOpen && (
              <p className="order-side-panel__hours-hint">
                Operating hours: {formatTime(selectedDaySchedule.openTime)} –{' '}
                {formatTime(selectedDaySchedule.closeTime)}
              </p>
            )}
          </div>

          {/* Validation Errors */}
          {validationErrors.length > 0 && (
            <div className="order-side-panel__section">
              {validationErrors.map((error, index) => (
                <div key={index} className="order-side-panel__error" role="alert">
                  <span className="order-side-panel__error-icon" aria-hidden="true">⚠️</span>
                  <span>{error}</span>
                </div>
              ))}
            </div>
          )}

          {/* Items Summary */}
          <div className="order-side-panel__section">
            <h3 className="order-side-panel__section-title">Order Items</h3>
            {items.length === 0 ? (
              <div className="order-side-panel__items-empty">
                No products selected yet. Click "Start Selecting Products" below to add items.
              </div>
            ) : (
              <div className="order-side-panel__items">
                {items.map((item) => (
                  <div key={item.product.id} className="order-side-panel__item">
                    <span className="order-side-panel__item-name">{item.product.name}</span>
                    <div className="order-side-panel__item-details">
                      <span className="order-side-panel__item-qty">×{item.quantity}</span>
                      <span className="order-side-panel__item-subtotal">
                        €{((item.quantity * item.product.price) / 100).toFixed(2)}
                      </span>
                    </div>
                  </div>
                ))}
                <div className="order-side-panel__total">
                  <span className="order-side-panel__total-label">Total</span>
                  <span className="order-side-panel__total-amount">
                    €{(total / 100).toFixed(2)}
                  </span>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="order-side-panel__footer">
          <button
            type="button"
            className="order-side-panel__select-btn"
            onClick={onStartSelection}
          >
            Start Selecting Products
          </button>
          {submitError && (
            <div className="order-side-panel__submit-error" role="alert">
              {submitError}
            </div>
          )}
          <button
            type="button"
            className="order-side-panel__submit-btn"
            disabled={isSubmitDisabled}
            onClick={handleSubmit}
            aria-disabled={isSubmitDisabled}
          >
            {submitting ? 'Submitting…' : isSubmitDisabled && !submitting ? 'Add products to submit' : 'Submit Order'}
          </button>
        </div>
      </aside>
    </>
  );
}
