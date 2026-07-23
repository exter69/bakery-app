import { useState, useEffect } from 'react';
import { useI18n } from '../../i18n';
import { listSavedLists } from '../../api/b2b-client';
import type { SavedList } from '../../types/b2b';

interface Props {
  bakeryId: string;
  onSelect: (items: { productId: string; quantity: number }[]) => void;
}

export function SavedListPicker({ bakeryId, onSelect }: Props) {
  const { t } = useI18n();
  const [lists, setLists] = useState<SavedList[]>([]);

  useEffect(() => {
    if (!bakeryId) return;
    listSavedLists(bakeryId)
      .then(setLists)
      .catch(() => setLists([]));
  }, [bakeryId]);

  if (lists.length === 0) return null;

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selectedId = e.target.value;
    if (!selectedId) return;
    const list = lists.find((l) => l.id === selectedId);
    if (list) {
      onSelect(list.items.map((i) => ({ productId: i.productId, quantity: i.quantity })));
    }
    // Reset select to show placeholder again
    e.target.value = '';
  };

  return (
    <select
      className="saved-list-picker"
      onChange={handleChange}
      defaultValue=""
      aria-label={t('comptoir.commander.savedLists')}
    >
      <option value="" disabled>{t('comptoir.commander.savedLists')}</option>
      {lists.map((list) => (
        <option key={list.id} value={list.id}>{list.name}</option>
      ))}
    </select>
  );
}
