# Implementation Plan: Comptoir Gaps (MA-69)

## Overview

Fix four issues in the B2B Comptoir portal: wire the Recurrences page to the existing recurring-orders API, activate the dead Editer button on LivraisonsPage, replace swallowed errors with visible error states, and correct French accent-stripped i18n strings.

## Tasks

- [ ] 1. Fix French accent-stripped i18n strings
  - [ ] 1.1 Correct all accent-stripped FR strings in `frontend/src/i18n/translations.ts` comptoir section
    - Fix: "A venir" -> "À venir", "Passees" -> "Passées", "Editer" -> "Éditer", "Confirmee" -> "Confirmée", "Prete" -> "Prête", "Livree" -> "Livrée", "recurrentes" -> "récurrentes", "recurrence" -> "récurrence", "Frequence" -> "Fréquence", "Desactiver" -> "Désactiver", "depassee" -> "dépassée", "Creneau" -> "Créneau", "sauvegardees" -> "sauvegardées", "derniere" -> "dernière", "Payee" -> "Payée"
    - _Requirements: 4.1, 4.2, 4.3_
  - [ ] 1.2 Add `comptoir.deliveries.itemCount` i18n key for EN, FR, and NL
    - EN: `"{n} items"`, FR: `"{n} articles"`, NL: `"{n} artikelen"`
    - _Requirements: 4.5_
  - [ ] 1.3 Replace hardcoded `{d.items.length} items` in `LivraisonsPage.tsx` with `t('comptoir.deliveries.itemCount', { n: d.items.length })`
    - Both the upcoming and past tables have this hardcoded string
    - _Requirements: 4.4_

- [ ] 2. Create shared ErrorState component
  - [ ] 2.1 Create `frontend/src/components/ErrorState.tsx`
    - Props: `message: string`, `onRetry: () => void`
    - Render an error icon (SVG), the message, and a retry button
    - Style with `error-state` BEM class, visually distinct from empty states (red/warning color)
    - _Requirements: 3.4, 3.5_

- [ ] 3. Fix swallowed errors on FacturesPage
  - [ ] 3.1 Add error state to `FacturesPage.tsx`
    - Add `const [error, setError] = useState<string | null>(null)` state
    - In the `catch` block of `fetchData`, set error message instead of `setInvoices([])`
    - Render `<ErrorState>` when `error` is set and `!loading`
    - Clear error on successful fetch
    - _Requirements: 3.1_
  - [ ] 3.2 Fix silent PDF download failure in `FacturesPage.tsx`
    - Replace `// Silently fail` catch block with a visible error notification
    - Add a `downloadError` state, render it as an inline alert that auto-dismisses after 5 seconds
    - _Requirements: 3.2_

- [ ] 4. Fix swallowed errors on LivraisonsPage
  - [ ] 4.1 Add error state to `LivraisonsPage.tsx`
    - Add `const [error, setError] = useState<string | null>(null)` state
    - In the `catch` block of `fetchData`, set error message instead of `setDeliveries([])`
    - Render `<ErrorState>` when `error` is set and `!loading`
    - Clear error on successful fetch
    - _Requirements: 3.3_

- [ ] 5. Wire the Editer button on LivraisonsPage
  - [ ] 5.1 Add onClick handler to the "Editer" button in `LivraisonsPage.tsx`
    - Pass delivery ID to a navigation or modal-open handler
    - Create an `editingOrderId` state; when set, show an inline edit modal
    - The modal fetches products for the bakery and allows quantity modification
    - On save, call `editOrder(orderId, items)` from `b2b-client.ts`
    - On failure, display error in the modal
    - On success, close modal and refetch deliveries
    - _Requirements: 2.1, 2.2, 2.3_

- [ ] 6. Checkpoint - Verify error states and i18n
  - Ensure TypeScript compiles without errors
  - Ensure all three locales render correctly with accents
  - Ensure error states are visually distinct from empty states

- [ ] 7. Wire RecurrencesPage to the recurring-orders API
  - [ ] 7.1 Add `createRecurringOrder` function to `frontend/src/api/recurring.ts`
    - Accepts `CreateRecurringOrderPayload` (bakeryId, items, scheduledDay, scheduledTime, frequency, selectionMode)
    - POSTs to `/recurring-orders`
    - _Requirements: 1.3_
  - [ ] 7.2 Rewrite `RecurrencesPage.tsx` data fetching
    - Replace local empty state with `fetchRecurringOrders(page)` on mount and page change
    - Add loading, error, page, and total state
    - Render ErrorState on fetch failure
    - Display paginated table with bakery name, frequency, item count, estimated total, and active/inactive badge
    - _Requirements: 1.1, 1.7_
  - [ ] 7.3 Implement the creation form in `RecurrencesPage.tsx`
    - Bakery dropdown populated from `listApprovedBakeries()`
    - Product picker for items (productId + quantity)
    - Frequency selector (weekly / biweekly)
    - Scheduled day dropdown (Monday-Sunday)
    - Scheduled time inputs (start, end)
    - Selection mode radio (fixed / bakeryChoice / randomFavorites)
    - On submit, call `createRecurringOrder` and append result to list
    - Show validation errors from API response
    - _Requirements: 1.2, 1.3_
  - [ ] 7.4 Implement pause/resume toggle in `RecurrencesPage.tsx`
    - Wire toggle button to call `pauseRecurringOrder(id)` or `resumeRecurringOrder(id)`
    - Update row status optimistically, revert on error
    - _Requirements: 1.4, 1.5_
  - [ ] 7.5 Implement delete action in `RecurrencesPage.tsx`
    - Add delete button with confirmation prompt
    - Call `deleteRecurringOrder(id)` and remove row on success
    - _Requirements: 1.6_

- [ ] 8. Add i18n keys for new RecurrencesPage and error UI strings
  - [ ] 8.1 Add missing i18n keys for the creation form, error messages, and edit modal
    - Keys for: form labels, validation messages, confirmation prompts, error state messages
    - All three locales (EN, FR, NL)
    - Ensure FR strings use correct accents from the start
    - _Requirements: 1.7, 4.1, 4.2_

- [ ] 9. Final checkpoint - Full compilation and review
  - Ensure `tsc --noEmit` passes for the frontend
  - Ensure all pages render without console errors
  - Verify no hardcoded English remains in comptoir pages
  - Ensure all error catches show user-visible feedback

## Notes

- The backend recurring-orders API already exists with full CRUD (handler, service, repository, migration). No backend changes needed.
- The scheduler worker that materializes recurring orders into real orders is out of scope (separate ticket).
- `editOrder` already exists in `b2b-client.ts` — the button just needs an `onClick` wiring.
- The `frontend/src/api/recurring.ts` already has `fetchRecurringOrders`, `pauseRecurringOrder`, `resumeRecurringOrder`, and `deleteRecurringOrder`. Only `createRecurringOrder` is missing.
- Error states must be visually distinct from empty states — use a different color/icon to avoid confusion during outages.

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "2.1"], "description": "i18n fixes and shared ErrorState component (no dependencies)" },
    { "id": 1, "tasks": ["1.3", "3.1", "3.2", "4.1"], "description": "Wire ErrorState into pages and fix hardcoded strings (depends on 1.2 and 2.1)" },
    { "id": 2, "tasks": ["5.1"], "description": "Wire Editer button (depends on error state pattern from wave 1)" },
    { "id": 3, "tasks": ["7.1"], "description": "Add createRecurringOrder API client function" },
    { "id": 4, "tasks": ["7.2", "7.3", "7.4", "7.5"], "description": "Rewrite RecurrencesPage (depends on 7.1)" },
    { "id": 5, "tasks": ["8.1"], "description": "Add remaining i18n keys for new components" },
    { "id": 6, "tasks": ["6", "9"], "description": "Final verification checkpoints" }
  ]
}
```
