import { useState, useCallback } from 'react';
import { createBundle } from '../../api/bundles';
import type { BundleType, BundleItem } from '../../types/bundle';
import './BundleForm.css';

export interface BundleFormProps {
  onSuccess: () => void;
  onCancel: () => void;
}

interface FormErrors {
  name?: string;
  type?: string;
  items?: string;
  description?: string;
  estimatedValue?: string;
  originalPrice?: string;
  discountedPrice?: string;
  quantityTotal?: string;
  pickupStartTime?: string;
  pickupEndTime?: string;
  pickupWindow?: string;
}

interface ItemRow {
  description: string;
  quantity: string;
}

export function BundleForm({ onSuccess, onCancel }: BundleFormProps) {
  const [name, setName] = useState('');
  const [type, setType] = useState<BundleType>('compose');
  const [items, setItems] = useState<ItemRow[]>([{ description: '', quantity: '1' }]);
  const [description, setDescription] = useState('');
  const [estimatedValue, setEstimatedValue] = useState('');
  const [photoUrl, setPhotoUrl] = useState('');
  const [originalPrice, setOriginalPrice] = useState('');
  const [discountedPrice, setDiscountedPrice] = useState('');
  const [quantityTotal, setQuantityTotal] = useState('1');
  const [pickupStartTime, setPickupStartTime] = useState('');
  const [pickupEndTime, setPickupEndTime] = useState('');

  const [errors, setErrors] = useState<FormErrors>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const validate = useCallback((): FormErrors => {
    const errs: FormErrors = {};

    if (!name.trim()) {
      errs.name = 'Name is required';
    } else if (name.trim().length > 100) {
      errs.name = 'Name must be 100 characters or less';
    }

    if (type === 'compose') {
      const validItems = items.filter((i) => i.description.trim());
      if (validItems.length === 0) {
        errs.items = 'At least one item is required';
      }
    }

    if (type === 'surprise') {
      if (!description.trim()) {
        errs.description = 'Description is required for surprise bundles';
      } else if (description.trim().length > 200) {
        errs.description = 'Description must be 200 characters or less';
      }
      if (!estimatedValue.trim()) {
        errs.estimatedValue = 'Estimated value is required';
      } else if (isNaN(Number(estimatedValue)) || Number(estimatedValue) <= 0) {
        errs.estimatedValue = 'Must be a positive number (cents)';
      }
    }

    const origP = Number(originalPrice);
    const discP = Number(discountedPrice);

    if (!originalPrice.trim() || isNaN(origP) || origP <= 0) {
      errs.originalPrice = 'Original price must be positive (cents)';
    }
    if (!discountedPrice.trim() || isNaN(discP) || discP <= 0) {
      errs.discountedPrice = 'Discounted price must be positive (cents)';
    }
    if (!errs.originalPrice && !errs.discountedPrice && discP >= origP) {
      errs.discountedPrice = 'Discounted price must be less than original price';
    }

    const qtyTotal = Number(quantityTotal);
    if (!quantityTotal.trim() || isNaN(qtyTotal) || qtyTotal < 1 || !Number.isInteger(qtyTotal)) {
      errs.quantityTotal = 'Quantity must be at least 1';
    }

    if (!pickupStartTime.trim()) {
      errs.pickupStartTime = 'Start time is required';
    }
    if (!pickupEndTime.trim()) {
      errs.pickupEndTime = 'End time is required';
    }
    if (pickupStartTime && pickupEndTime && pickupStartTime >= pickupEndTime) {
      errs.pickupWindow = 'Start time must be before end time';
    }

    return errs;
  }, [name, type, items, description, estimatedValue, originalPrice, discountedPrice, quantityTotal, pickupStartTime, pickupEndTime]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);

    const formErrors = validate();
    setErrors(formErrors);

    if (Object.keys(formErrors).length > 0) return;

    setSubmitting(true);
    try {
      const bundleItems: BundleItem[] | undefined =
        type === 'compose'
          ? items
              .filter((i) => i.description.trim())
              .map((i) => ({
                description: i.description.trim(),
                quantity: Number(i.quantity) || 1,
              }))
          : undefined;

      await createBundle({
        name: name.trim(),
        type,
        photoUrl: photoUrl.trim() || undefined,
        description: type === 'surprise' ? description.trim() : undefined,
        estimatedValue: type === 'surprise' ? Number(estimatedValue) : undefined,
        originalPrice: Number(originalPrice),
        discountedPrice: Number(discountedPrice),
        quantityTotal: Number(quantityTotal),
        pickupStartTime,
        pickupEndTime,
        items: bundleItems,
      });

      onSuccess();
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Failed to create bundle');
    } finally {
      setSubmitting(false);
    }
  };

  const addItem = () => {
    setItems((prev) => [...prev, { description: '', quantity: '1' }]);
  };

  const removeItem = (index: number) => {
    setItems((prev) => prev.filter((_, i) => i !== index));
  };

  const updateItem = (index: number, field: keyof ItemRow, value: string) => {
    setItems((prev) =>
      prev.map((item, i) => (i === index ? { ...item, [field]: value } : item))
    );
  };

  return (
    <form className="bundle-form" onSubmit={handleSubmit} noValidate>
      {/* Name */}
      <div className="dash-form__field">
        <label className="dash-form__label" htmlFor="bundle-name">Name *</label>
        <input
          id="bundle-name"
          className="dash-form__input"
          type="text"
          maxLength={100}
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. Panier surprise du soir"
        />
        {errors.name && <span className="bundle-form__error">{errors.name}</span>}
      </div>

      {/* Type selector */}
      <div className="dash-form__field">
        <span className="dash-form__label">Type *</span>
        <div className="bundle-form__radio-group">
          <label className="bundle-form__radio-label">
            <input
              type="radio"
              name="bundle-type"
              value="compose"
              checked={type === 'compose'}
              onChange={() => setType('compose')}
            />
            Composé
          </label>
          <label className="bundle-form__radio-label">
            <input
              type="radio"
              name="bundle-type"
              value="surprise"
              checked={type === 'surprise'}
              onChange={() => setType('surprise')}
            />
            Surprise
          </label>
        </div>
      </div>

      {/* Composé items */}
      {type === 'compose' && (
        <div className="dash-form__field">
          <span className="dash-form__label">Items *</span>
          <div className="bundle-form__items">
            {items.map((item, idx) => (
              <div key={idx} className="bundle-form__item-row">
                <div className="dash-form__field">
                  <input
                    className="dash-form__input"
                    type="text"
                    placeholder="Description"
                    value={item.description}
                    onChange={(e) => updateItem(idx, 'description', e.target.value)}
                  />
                </div>
                <div className="dash-form__field dash-form__field--qty">
                  <input
                    className="dash-form__input"
                    type="number"
                    min={1}
                    placeholder="Qty"
                    value={item.quantity}
                    onChange={(e) => updateItem(idx, 'quantity', e.target.value)}
                  />
                </div>
                {items.length > 1 && (
                  <button
                    type="button"
                    className="bundle-form__item-remove"
                    onClick={() => removeItem(idx)}
                    aria-label={`Remove item ${idx + 1}`}
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                )}
              </div>
            ))}
          </div>
          <button
            type="button"
            className="dash-btn dash-btn--secondary dash-btn--sm bundle-form__add-item"
            onClick={addItem}
          >
            + Add item
          </button>
          {errors.items && <span className="bundle-form__error">{errors.items}</span>}
        </div>
      )}

      {/* Surprise fields */}
      {type === 'surprise' && (
        <>
          <div className="dash-form__field">
            <label className="dash-form__label" htmlFor="bundle-description">Description *</label>
            <textarea
              id="bundle-description"
              className="dash-form__textarea"
              maxLength={200}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="e.g. Assortiment de viennoiseries et pains du jour"
            />
            {errors.description && <span className="bundle-form__error">{errors.description}</span>}
          </div>
          <div className="dash-form__field">
            <label className="dash-form__label" htmlFor="bundle-estimated-value">Estimated value (cents) *</label>
            <input
              id="bundle-estimated-value"
              className="dash-form__input"
              type="number"
              min={1}
              value={estimatedValue}
              onChange={(e) => setEstimatedValue(e.target.value)}
              placeholder="e.g. 1200"
            />
            {errors.estimatedValue && <span className="bundle-form__error">{errors.estimatedValue}</span>}
          </div>
        </>
      )}

      {/* Photo URL */}
      <div className="dash-form__field">
        <label className="dash-form__label" htmlFor="bundle-photo">Photo URL (optional)</label>
        <input
          id="bundle-photo"
          className="dash-form__input"
          type="url"
          value={photoUrl}
          onChange={(e) => setPhotoUrl(e.target.value)}
          placeholder="https://..."
        />
      </div>

      {/* Pricing */}
      <div className="bundle-form__row">
        <div className="dash-form__field">
          <label className="dash-form__label" htmlFor="bundle-original-price">Original price (cents) *</label>
          <input
            id="bundle-original-price"
            className="dash-form__input"
            type="number"
            min={1}
            value={originalPrice}
            onChange={(e) => setOriginalPrice(e.target.value)}
            placeholder="e.g. 1000"
          />
          {errors.originalPrice && <span className="bundle-form__error">{errors.originalPrice}</span>}
        </div>
        <div className="dash-form__field">
          <label className="dash-form__label" htmlFor="bundle-discounted-price">Discounted price (cents) *</label>
          <input
            id="bundle-discounted-price"
            className="dash-form__input"
            type="number"
            min={1}
            value={discountedPrice}
            onChange={(e) => setDiscountedPrice(e.target.value)}
            placeholder="e.g. 450"
          />
          {errors.discountedPrice && <span className="bundle-form__error">{errors.discountedPrice}</span>}
        </div>
      </div>

      {/* Quantity */}
      <div className="dash-form__field">
        <label className="dash-form__label" htmlFor="bundle-quantity">Quantity *</label>
        <input
          id="bundle-quantity"
          className="dash-form__input"
          type="number"
          min={1}
          value={quantityTotal}
          onChange={(e) => setQuantityTotal(e.target.value)}
          placeholder="e.g. 5"
        />
        {errors.quantityTotal && <span className="bundle-form__error">{errors.quantityTotal}</span>}
      </div>

      {/* Pickup window */}
      <div className="bundle-form__row">
        <div className="dash-form__field">
          <label className="dash-form__label" htmlFor="bundle-pickup-start">Pickup start (HH:MM) *</label>
          <input
            id="bundle-pickup-start"
            className="dash-form__input"
            type="time"
            value={pickupStartTime}
            onChange={(e) => setPickupStartTime(e.target.value)}
          />
          {errors.pickupStartTime && <span className="bundle-form__error">{errors.pickupStartTime}</span>}
        </div>
        <div className="dash-form__field">
          <label className="dash-form__label" htmlFor="bundle-pickup-end">Pickup end (HH:MM) *</label>
          <input
            id="bundle-pickup-end"
            className="dash-form__input"
            type="time"
            value={pickupEndTime}
            onChange={(e) => setPickupEndTime(e.target.value)}
          />
          {errors.pickupEndTime && <span className="bundle-form__error">{errors.pickupEndTime}</span>}
        </div>
      </div>
      {errors.pickupWindow && <span className="bundle-form__error">{errors.pickupWindow}</span>}

      {/* Submit error */}
      {submitError && <div className="dash-msg dash-msg--error">{submitError}</div>}

      {/* Actions */}
      <div className="bundle-form__actions">
        <button
          type="submit"
          className="dash-btn dash-btn--primary"
          disabled={submitting}
        >
          {submitting ? 'Creating...' : 'Create bundle'}
        </button>
        <button
          type="button"
          className="dash-btn dash-btn--secondary"
          onClick={onCancel}
          disabled={submitting}
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
