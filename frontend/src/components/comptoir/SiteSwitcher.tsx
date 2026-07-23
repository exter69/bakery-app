import { useState, useEffect, createContext, useContext } from 'react';
import { listSites } from '../../api/b2b-client';
import { useI18n } from '../../i18n';
import type { DeliverySite } from '../../types/b2b';

// Context for selected site, accessible from checkout and other components
interface SiteContextValue {
  selectedSite: DeliverySite | null;
  setSelectedSite: (site: DeliverySite) => void;
  sites: DeliverySite[];
}

const SiteContext = createContext<SiteContextValue>({
  selectedSite: null,
  setSelectedSite: () => {},
  sites: [],
});

export function useSiteContext() {
  return useContext(SiteContext);
}

export function SiteProvider({ children }: { children: React.ReactNode }) {
  const [sites, setSites] = useState<DeliverySite[]>([]);
  const [selectedSite, setSelectedSite] = useState<DeliverySite | null>(null);

  useEffect(() => {
    listSites()
      .then((result) => {
        setSites(result);
        if (result.length > 0) {
          setSelectedSite(result[0]);
        }
      })
      .catch(() => {});
  }, []);

  return (
    <SiteContext.Provider value={{ selectedSite, setSelectedSite, sites }}>
      {children}
    </SiteContext.Provider>
  );
}

export function SiteSwitcher() {
  const { t } = useI18n();
  const { selectedSite, setSelectedSite, sites } = useSiteContext();

  if (sites.length === 0) {
    return null;
  }

  if (sites.length === 1) {
    return (
      <span className="comptoir-nav__site-single" title={t('comptoir.siteSwitcher.single')}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
          <circle cx="12" cy="10" r="3" />
        </svg>
        {sites[0].name}
      </span>
    );
  }

  return (
    <label className="comptoir-nav__site-switcher">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
        <circle cx="12" cy="10" r="3" />
      </svg>
      <select
        value={selectedSite?.id ?? ''}
        onChange={(e) => {
          const site = sites.find((s) => s.id === e.target.value);
          if (site) setSelectedSite(site);
        }}
        aria-label={t('comptoir.siteSwitcher.label')}
      >
        {sites.map((site) => (
          <option key={site.id} value={site.id}>
            {site.name}
          </option>
        ))}
      </select>
    </label>
  );
}
