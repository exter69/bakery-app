import { useCallback, useEffect, useMemo, useState } from 'react';
import type { DaySchedule, Product } from '../types/bakery';
import './ReservationSidePanel.css';

export interface ReservationItem {
  product: Product;
  quantity: number;
}

interface ReservationSidePanelProps {
  isOpen: boolean;
  onClose: () => void;
  schedule: DaySchedule[];
  items: ReservationItem[];
  onItemsChange: (items: ReservationItem[]) => void;
  onStartSelection?: () => void;
  onSubmit?: (day: string, startTime: string, endTime: string) => void;
  submitting?: boolean;
  submitError?: string | null;
}

const DAYS = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];

function generateTimeSlots(open: { hour: number; minute: number }, close: { hour: number; minute: number }): string[] {
  const slots: string[] = [];
  let h = open.hour;
  let m = open.minute;

  while (h < close.hour || (h === close.hour && m < close.minute)) {
    const startLabel = `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
    // Advance 30 minutes
    m += 30;
    if (m >= 60) {
      h += 1;
      m -= 60;
    }
    // Don't exceed closing time
    if (h < close.hour || (h === close.hour && m <= close.minute)) {
      const endLabel = `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
      slots.push(`${startLabel} - ${endLabel}`);
    }
  }
  return slots;
}

export default function ReservationSidePanel({
  isOpen,
  onClose,
  schedule,
  items,
  onItemsChange,
  onStartSelection,
  onSubmit,
  submitting = false,
  submitError = null,
}: ReservationSidePanelProps) {
  const [selectedDay, setSelectedDay] = useState<string | null>(null);
  const [selectedTimeSlot, setSelectedTimeSlot] = useState<string | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);

  // Reset all state when panel closes
  const resetState = useCallback(() => {
    setSelectedDay(null);
    setSelectedTimeSlot(null);
    setValidationError(null);
    setSubmitted(false);
    onItemsChange([]);
  }, [onItemsChange]);

  useEffect(() => {
    if (!isOpen) {
      resetState();
    }
  }, [isOpen, resetState]);

  // Get schedule for the selected day
  const selectedDaySchedule = useMemo(() => {
    if (!selectedDay) return null;
    return schedule.find((s) => s.day === selectedDay) ?? null;
  }, [selectedDay, schedule]);

  // Generate available time slots based on selected day's schedule
  const timeSlots = useMemo(() => {
    if (!selectedDaySchedule || !selectedDaySchedule.isOpen) return [];
    return generateTimeSlots(selectedDaySchedule.openTime, selectedDaySchedule.closeTime);
  }, [selectedDaySchedule]);

  // Check if a day is available (bakery is open)
  const isDayAvailable = useCallback(
    (day: string) => {
      const daySchedule = schedule.find((s) => s.day === day);
      return daySchedule?.isOpen ?? false;
    },
    [schedule]
  );

  // Handle day selection with validation
  const handleDaySelect = (day: string) => {
    setValidationError(null);
    if (!isDayAvailable(day)) {
      setValidationError(`Bakery is closed on ${day}. Please select an available day.`);
      return;
    }
    setSelectedDay(day);
    setSelectedTimeSlot(null); // Reset time when day changes
  };

  // Handle time slot selection
  const handleTimeSlotSelect = (slot: string) => {
    setValidationError(null);
    setSelectedTimeSlot(slot);
  };

  // Calculate estimated total
  const estimatedTotal = useMemo(() => {
    return items.reduce((sum, item) => sum + item.product.price * item.quantity, 0);
  }, [items]);

  // Check if submission is valid
  const canSubmit = items.length > 0 && selectedDay !== null && selectedTimeSlot !== null && !submitting;

  // Handle reservation submission
  const handleSubmit = () => {
    if (!canSubmit || !selectedDay || !selectedTimeSlot) return;

    if (onSubmit) {
      // Parse time slot "HH:MM - HH:MM" into start/end
      const [startTime, endTime] = selectedTimeSlot.split(' - ');
      onSubmit(selectedDay, startTime, endTime);
    } else {
      // Fallback: local-only confirmation (no API call)
      setSubmitted(true);
    }
  };

  // Handle panel close
  const handleClose = () => {
    onClose();
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className={`reservation-panel-backdrop${isOpen ? ' reservation-panel-backdrop--visible' : ''}`}
        onClick={handleClose}
        aria-hidden="true"
      />

      {/* Panel */}
      <aside
        className={`reservation-panel${isOpen ? ' reservation-panel--open' : ''}`}
        role="dialog"
        aria-modal="true"
        aria-label="Make Reservation"
      >
        {/* Header */}
        <div className="reservation-panel__header">
          <h2 className="reservation-panel__title">Make Reservation</h2>
          <button
            type="button"
            className="reservation-panel__close-btn"
            onClick={handleClose}
            aria-label="Close reservation panel"
          >
            ✕
          </button>
        </div>

        {submitted ? (
          /* Success State */
          <div className="reservation-panel__success">
            <span className="reservation-panel__success-icon" aria-hidden="true">✓</span>
            <h3 className="reservation-panel__success-title">Reservation Confirmed!</h3>
            <p className="reservation-panel__success-message">
              Your reservation has been placed successfully.
            </p>
            <span className="reservation-panel__success-badge">Pay on arrival</span>
          </div>
        ) : (
          <>
            {/* Body */}
            <div className="reservation-panel__body">
              {/* Day Selector */}
              <div className="reservation-panel__section">
                <span className="reservation-panel__section-label">Select Day</span>
                <div className="reservation-panel__days">
                  {DAYS.map((day) => (
                    <button
                      key={day}
                      type="button"
                      className={`reservation-panel__day-btn${selectedDay === day ? ' reservation-panel__day-btn--selected' : ''}`}
                      onClick={() => handleDaySelect(day)}
                      disabled={!isDayAvailable(day)}
                      aria-pressed={selectedDay === day}
                      aria-label={`${day}${!isDayAvailable(day) ? ' (closed)' : ''}`}
                    >
                      {day.slice(0, 3)}
                    </button>
                  ))}
                </div>
              </div>

              {/* Time Slot Selector */}
              {selectedDay && selectedDaySchedule?.isOpen && (
                <div className="reservation-panel__section">
                  <span className="reservation-panel__section-label">Select Time</span>
                  <div className="reservation-panel__time-slots">
                    {timeSlots.map((slot) => (
                      <button
                        key={slot}
                        type="button"
                        className={`reservation-panel__time-btn${selectedTimeSlot === slot ? ' reservation-panel__time-btn--selected' : ''}`}
                        onClick={() => handleTimeSlotSelect(slot)}
                        aria-pressed={selectedTimeSlot === slot}
                      >
                        {slot}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Validation Error */}
              {validationError && (
                <div className="reservation-panel__error" role="alert">
                  {validationError}
                </div>
              )}

              {/* Order Summary */}
              <div className="reservation-panel__section">
                <span className="reservation-panel__section-label">Reservation Summary</span>
                <div className="reservation-panel__summary">
                  {items.length === 0 ? (
                    <p className="reservation-panel__summary-empty">
                      No products selected yet.
                    </p>
                  ) : (
                    <>
                      {items.map((item) => (
                        <div key={item.product.id} className="reservation-panel__summary-item">
                          <div>
                            <span className="reservation-panel__summary-item-name">
                              {item.product.name}
                            </span>
                            <span className="reservation-panel__summary-item-qty">
                              ×{item.quantity}
                            </span>
                          </div>
                          <span className="reservation-panel__summary-item-subtotal">
                            €{((item.product.price * item.quantity) / 100).toFixed(2)}
                          </span>
                        </div>
                      ))}
                      <div className="reservation-panel__summary-total">
                        <span className="reservation-panel__summary-total-label">
                          Estimated Total
                        </span>
                        <span className="reservation-panel__summary-total-amount">
                          €{(estimatedTotal / 100).toFixed(2)}
                        </span>
                      </div>
                    </>
                  )}
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className="reservation-panel__footer">
              {onStartSelection && (
                <button
                  type="button"
                  className="reservation-panel__select-btn"
                  onClick={onStartSelection}
                >
                  Start Selecting Products
                </button>
              )}
              {submitError && (
                <div className="reservation-panel__submit-error" role="alert">
                  {submitError}
                </div>
              )}
              <button
                type="button"
                className="reservation-panel__submit-btn"
                disabled={!canSubmit}
                onClick={handleSubmit}
              >
                {submitting ? 'Submitting…' : 'Confirm Reservation'}
              </button>
              {!canSubmit && items.length === 0 && (
                <p className="reservation-panel__validation-msg">
                  Select at least one product to confirm your reservation.
                </p>
              )}
            </div>
          </>
        )}
      </aside>
    </>
  );
}
