import { useCallback, useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { searchProducts } from '../api/bakeries';
import { useI18n } from '../i18n';
import HealthScoreDisplay from './HealthScoreDisplay';
import type { ProductSearchResult, ProductSearchParams } from '../api/bakeries';
import './SearchBar.css';

const DEBOUNCE_MS = 300;

const ALLERGEN_OPTIONS = [
  'gluten', 'dairy', 'eggs', 'nuts', 'peanuts', 'soy',
  'fish', 'crustaceans', 'celery', 'mustard', 'sesame',
  'sulphites', 'lupin', 'molluscs',
];

const CATEGORY_OPTIONS = ['Viennoiseries', 'Pains', 'Pâtisseries', 'Snacks', 'Boissons'];

const HEALTH_SCORE_OPTIONS = [1, 2, 3, 4, 5];

export default function SearchBar() {
  const { t } = useI18n();
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState('');
  const [excludeAllergens, setExcludeAllergens] = useState<string[]>([]);
  const [minHealthScore, setMinHealthScore] = useState<number | undefined>(undefined);
  const [results, setResults] = useState<ProductSearchResult[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const doSearch = useCallback(async (params: ProductSearchParams) => {
    setLoading(true);
    try {
      const res = await searchProducts(params);
      setResults(res.items);
      setTotal(res.total);
      setHasSearched(true);
    } catch {
      setResults([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, []);

  // Trigger search when filters or page change (debounced for query)
  useEffect(() => {
    if (!query && !category && excludeAllergens.length === 0 && !minHealthScore) {
      setResults([]);
      setTotal(0);
      setHasSearched(false);
      return;
    }

    if (debounceRef.current) clearTimeout(debounceRef.current);

    debounceRef.current = setTimeout(() => {
      doSearch({ q: query, category, excludeAllergens, minHealthScore, page });
    }, DEBOUNCE_MS);

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [query, category, excludeAllergens, minHealthScore, page, doSearch]);

  const handleQueryChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value);
    setPage(1);
  };

  const toggleAllergen = (allergen: string) => {
    setExcludeAllergens((prev) =>
      prev.includes(allergen) ? prev.filter((a) => a !== allergen) : [...prev, allergen],
    );
    setPage(1);
  };

  const isActive = query || category || excludeAllergens.length > 0 || minHealthScore;

  return (
    <div className="search-bar">
      <div className="search-bar__input-row">
        <div className="search-bar__input-wrap">
          <svg
            className="search-bar__icon"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="search"
            className="search-bar__input"
            placeholder={t('search.placeholder')}
            value={query}
            onChange={handleQueryChange}
            aria-label={t('search.placeholder')}
          />
        </div>
        <button
          type="button"
          className={`search-bar__filter-toggle${filtersOpen ? ' search-bar__filter-toggle--active' : ''}`}
          onClick={() => setFiltersOpen(!filtersOpen)}
          aria-expanded={filtersOpen}
          aria-label={t('search.filters.category')}
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <line x1="4" y1="6" x2="20" y2="6" />
            <line x1="8" y1="12" x2="16" y2="12" />
            <line x1="11" y1="18" x2="13" y2="18" />
          </svg>
        </button>
      </div>

      {filtersOpen && (
        <div className="search-bar__filters">
          <div className="search-bar__filter-group">
            <label className="search-bar__filter-label">{t('search.filters.category')}</label>
            <select
              className="search-bar__select"
              value={category}
              onChange={(e) => { setCategory(e.target.value); setPage(1); }}
            >
              <option value="">—</option>
              {CATEGORY_OPTIONS.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>

          <div className="search-bar__filter-group">
            <label className="search-bar__filter-label">{t('search.filters.allergens')}</label>
            <div className="search-bar__allergen-chips">
              {ALLERGEN_OPTIONS.map((allergen) => (
                <button
                  key={allergen}
                  type="button"
                  className={`search-bar__allergen-chip${excludeAllergens.includes(allergen) ? ' search-bar__allergen-chip--active' : ''}`}
                  onClick={() => toggleAllergen(allergen)}
                >
                  {t(`allergen.${allergen}`)}
                </button>
              ))}
            </div>
          </div>

          <div className="search-bar__filter-group">
            <label className="search-bar__filter-label">{t('search.filters.healthScore')}</label>
            <select
              className="search-bar__select"
              value={minHealthScore ?? ''}
              onChange={(e) => { setMinHealthScore(e.target.value ? Number(e.target.value) : undefined); setPage(1); }}
            >
              <option value="">—</option>
              {HEALTH_SCORE_OPTIONS.map((s) => (
                <option key={s} value={s}>{s}+</option>
              ))}
            </select>
          </div>
        </div>
      )}

      {isActive && (
        <div className="search-bar__results" role="region" aria-live="polite">
          {loading && <div className="search-bar__loading"><div className="spinner" aria-label="Loading" /></div>}

          {!loading && hasSearched && results.length === 0 && (
            <p className="search-bar__no-results">{t('search.noResults')}</p>
          )}

          {!loading && results.length > 0 && (
            <>
              <div className="search-bar__result-grid">
                {results.map((item) => (
                  <Link
                    key={item.product.id}
                    to={`/bakeries/${item.bakeryId}`}
                    className="search-result-card"
                  >
                    {item.product.photoUrl ? (
                      <img
                        src={item.product.photoUrl}
                        alt={item.product.name}
                        className="search-result-card__photo"
                        loading="lazy"
                      />
                    ) : (
                      <div className="search-result-card__photo-placeholder" aria-hidden="true" />
                    )}
                    <div className="search-result-card__body">
                      <span className="search-result-card__name">{item.product.name}</span>
                      <span className="search-result-card__bakery">{item.bakeryName}</span>
                      <span className="search-result-card__price">
                        &euro;{(item.product.price / 100).toFixed(2)}
                      </span>
                      {item.product.healthScore != null && (
                        <HealthScoreDisplay score={item.product.healthScore} />
                      )}
                    </div>
                  </Link>
                ))}
              </div>

              {total > results.length && (
                <div className="search-bar__pagination">
                  <button
                    type="button"
                    className="search-bar__page-btn"
                    disabled={page <= 1}
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                  >
                    &larr;
                  </button>
                  <span className="search-bar__page-info">{page}</span>
                  <button
                    type="button"
                    className="search-bar__page-btn"
                    disabled={page * 20 >= total}
                    onClick={() => setPage((p) => p + 1)}
                  >
                    &rarr;
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}
