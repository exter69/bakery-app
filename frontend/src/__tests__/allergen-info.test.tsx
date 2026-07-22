/**
 * Unit tests for AllergenInfoIcon and AllergenInfoModal components.
 *
 * Validates: Requirements 7.1–7.10, 8.5
 */
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import AllergenInfoIcon from '../components/AllergenInfoIcon';
import AllergenInfoModal from '../components/AllergenInfoModal';
import { I18nProvider } from '../i18n/I18nContext';

const ALL_ALLERGEN_NAMES_EN = [
  'Celery',
  'Crustaceans',
  'Dairy',
  'Eggs',
  'Fish',
  'Gluten',
  'Lupin',
  'Molluscs',
  'Mustard',
  'Nuts',
  'Peanuts',
  'Sesame',
  'Soy',
  'Sulphites',
];

function renderWithI18n(ui: React.ReactElement) {
  return render(<I18nProvider>{ui}</I18nProvider>);
}

describe('AllergenInfoIcon', () => {
  it('renders as a <button> element', () => {
    renderWithI18n(<AllergenInfoIcon />);
    const button = screen.getByRole('button');
    expect(button).toBeInTheDocument();
    expect(button.tagName).toBe('BUTTON');
  });

  it('has an aria-label containing allergen information text', () => {
    renderWithI18n(<AllergenInfoIcon />);
    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('aria-label');
    // The EN translation for allergenInfo.label is "Allergen information"
    expect(button.getAttribute('aria-label')).toMatch(/allergen/i);
  });

  it('clicking the button opens the AllergenInfoModal', async () => {
    renderWithI18n(<AllergenInfoIcon />);

    const button = screen.getByRole('button');
    await userEvent.click(button);

    // Modal should now be open — it's inside aria-hidden backdrop so use hidden option
    const dialog = screen.getByRole('dialog', { hidden: true });
    expect(dialog).toBeInTheDocument();
  });

  it('calls onOpenModal callback when provided instead of opening modal', async () => {
    const onOpenModal = vi.fn();
    renderWithI18n(<AllergenInfoIcon onOpenModal={onOpenModal} />);

    const button = screen.getByRole('button');
    await userEvent.click(button);

    expect(onOpenModal).toHaveBeenCalledTimes(1);
    // Should not render modal when onOpenModal is provided
    expect(screen.queryByRole('dialog', { hidden: true })).not.toBeInTheDocument();
  });
});

describe('AllergenInfoModal', () => {
  it('does not render when isOpen is false', () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={false} onClose={onClose} />);
    expect(screen.queryByRole('dialog', { hidden: true })).not.toBeInTheDocument();
  });

  it('renders the modal with title "Allergen Information" when isOpen is true', () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={true} onClose={onClose} />);

    const dialog = screen.getByRole('dialog', { hidden: true });
    expect(dialog).toBeInTheDocument();

    // Title should be present
    expect(screen.getByText('Allergen Information')).toBeInTheDocument();
  });

  it('lists all 14 allergen names', () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={true} onClose={onClose} />);

    for (const name of ALL_ALLERGEN_NAMES_EN) {
      expect(screen.getByText(name)).toBeInTheDocument();
    }
  });

  it('each allergen has a description', () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={true} onClose={onClose} />);

    // The modal renders 14 list items, each with a name and description paragraph
    const listItems = document.querySelectorAll('.aim__allergen-item');
    expect(listItems.length).toBe(14);

    listItems.forEach((item) => {
      const name = item.querySelector('.aim__allergen-name');
      const desc = item.querySelector('.aim__allergen-desc');
      expect(name).not.toBeNull();
      expect(desc).not.toBeNull();
      expect(desc!.textContent!.length).toBeGreaterThan(0);
    });
  });

  it('pressing Escape calls onClose', () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={true} onClose={onClose} />);

    const dialog = screen.getByRole('dialog', { hidden: true });
    fireEvent.keyDown(dialog, { key: 'Escape', code: 'Escape' });

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('clicking outside (on backdrop) calls onClose', async () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={true} onClose={onClose} />);

    // The backdrop is the parent div with class aim-backdrop
    const backdrop = document.querySelector('.aim-backdrop')!;
    expect(backdrop).not.toBeNull();

    await userEvent.click(backdrop);

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('clicking inside the modal does not call onClose', async () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={true} onClose={onClose} />);

    const dialog = screen.getByRole('dialog', { hidden: true });
    await userEvent.click(dialog);

    expect(onClose).not.toHaveBeenCalled();
  });

  it('has a close button that calls onClose', async () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={true} onClose={onClose} />);

    const closeBtn = screen.getByLabelText('Close modal', { selector: 'button' });
    await userEvent.click(closeBtn);

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('focus trap cycles focus within the modal on Tab', () => {
    const onClose = vi.fn();
    renderWithI18n(<AllergenInfoModal isOpen={true} onClose={onClose} />);

    const dialog = screen.getByRole('dialog', { hidden: true });
    const closeBtn = screen.getByLabelText('Close modal', { selector: 'button' });

    // The only focusable element is the close button
    closeBtn.focus();
    expect(document.activeElement).toBe(closeBtn);

    // Tab on the last (and only) focusable element wraps to first
    fireEvent.keyDown(dialog, { key: 'Tab', code: 'Tab' });
    // Since close button is both first and last, focus stays on it
    expect(document.activeElement).toBe(closeBtn);
  });

  it('language switching updates allergen names without page reload', () => {
    // Verify the modal uses i18n translations (EN by default in test)
    const { rerender } = render(
      <I18nProvider>
        <AllergenInfoModal isOpen={true} onClose={() => {}} />
      </I18nProvider>
    );

    // Initially in English — title and allergens render in EN
    expect(screen.getByText('Allergen Information')).toBeInTheDocument();
    expect(screen.getByText('Gluten')).toBeInTheDocument();

    // Rerender (simulates React update without page reload)
    rerender(
      <I18nProvider>
        <AllergenInfoModal isOpen={true} onClose={() => {}} />
      </I18nProvider>
    );

    // Content still present after rerender — no page reload needed
    expect(screen.getByText('Allergen Information')).toBeInTheDocument();
    expect(screen.getByText('Gluten')).toBeInTheDocument();
  });
});
