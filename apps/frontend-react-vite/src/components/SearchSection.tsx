import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import React, { useCallback, useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Search, ChevronDown, Globe, Building, Plus, Edit, Trash2, Chrome, Zap, Shield } from 'lucide-react';
import { SearchEngineDialog, type SearchEngineFormData } from './SearchEngineDialog';
import type { SearchEngine as APISearchEngine } from '@/types/search-engine';
import { useCreateSearchEngine, useDeleteSearchEngine, useSearchEngines, useUpdateSearchEngine } from '@/queries/use-search-engines';
import { openLink } from '@/lib/linkUtils';
import { toast } from '@/hooks/use-toast';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

interface SearchEngine {
  id: string;
  name: string;
  placeholder: string;
  url: string;
  color: string;
  iconName: string;
  isCustom?: boolean;
}

const getIconComponent = (iconName?: string, engineName?: string) => {
  if (engineName) {
    const name = engineName.toLowerCase();
    if (name.includes('google')) return Chrome;
    if (name.includes('bing')) return Globe;
    if (name.includes('duckduckgo')) return Shield;
    if (name.includes('so') || name.includes('内网') || name.includes('内部')) return Building;
    if (name.includes('必应')) return Globe;
  }
  switch (iconName) {
    case 'globe': return Globe;
    case 'building': return Building;
    case 'plus': return Plus;
    case 'chrome': return Chrome;
    case 'shield': return Shield;
    case 'zap': return Zap;
    default: return Search;
  }
};

const getEngineDescription = (engine: SearchEngine, t: TFunction) => {
  const name = engine.name.toLowerCase();
  if (name.includes('google')) return t('search.descGoogle');
  if (name.includes('bing')) return t('search.descBing');
  if (name.includes('duckduckgo')) return t('search.descDuckduckgo');
  if (name.includes('baidu') || name.includes('百度')) return t('search.descBaidu');
  if (name.includes('so') || name.includes('内网') || name.includes('内部')) return t('search.descInternal');
  if (name.includes('必应')) return t('search.descBingCn');
  if (engine.isCustom) return engine.placeholder || t('search.descCustom');
  return t('search.descDefault');
};

const getColorForEngine = (name: string, index: number) => {
  const colors = ['#6b7280', '#059669', '#dc2626', '#2563eb', '#7c3aed', '#ea580c', '#0891b2', '#be123c'];
  if (name.toLowerCase().includes('google')) return '#6b7280';
  if (name.toLowerCase().includes('bing')) return '#059669';
  if (name.toLowerCase().includes('so') || name.toLowerCase().includes('内部')) return '#dc2626';
  return colors[index % colors.length];
};

const convertAPIEngineToLocal = (apiEngine: APISearchEngine, index: number, t: TFunction): SearchEngine => {
  try {
    const safeUrlTemplate = apiEngine.url_template || 'https://www.google.com/search?q={query}';
    return {
      id: apiEngine.id.toString(),
      name: apiEngine.name,
      placeholder: t('search.placeholder', { name: apiEngine.name }),
      url: safeUrlTemplate.replace('{query}', ''),
      color: apiEngine.color || getColorForEngine(apiEngine.name, index),
      iconName: apiEngine.icon || 'search',
      isCustom: true,
    };
  } catch (error) {
    console.error('Error converting API engine to local:', error, apiEngine);
    return {
      id: 'fallback-google',
      name: 'Google',
      placeholder: t('search.placeholderGoogle'),
      url: 'https://www.google.com/search?q=',
      color: '#4285f4',
      iconName: 'chrome',
      isCustom: false,
    };
  }
};

const getDefaultSearchEngines = (t: TFunction): SearchEngine[] => [
  { id: 'google', name: 'Google', iconName: 'chrome', url: 'https://www.google.com/search?q=', placeholder: t('search.placeholderGoogle'), color: '#4285f4' },
  { id: 'bing', name: 'Bing', iconName: 'globe', url: 'https://www.bing.com/search?q=', placeholder: t('search.placeholderBing'), color: '#00a1f1' },
  { id: 'so', name: 'SO内部搜索', iconName: 'building', url: 'https://so.company.com/search?q=', placeholder: t('search.placeholderSO'), color: '#dc2626' },
];

interface SearchSectionProps {
  searchQuery?: string;
  onSearchQueryChange?: (query: string) => void;
  selectedEngine?: string;
  onEngineChange?: (engineId: string) => void;
}

export default function SearchSection({
  searchQuery: externalSearchQuery,
  onSearchQueryChange,
  selectedEngine: externalSelectedEngine,
  onEngineChange
}: SearchSectionProps = {}) {
  const [internalSearchQuery, setInternalSearchQuery] = useState('');
  const [internalSelectedEngine, setInternalSelectedEngine] = useState('google');
  const { t } = useTranslation();

  const searchQuery = externalSearchQuery !== undefined ? externalSearchQuery : internalSearchQuery;
  const selectedEngine = externalSelectedEngine !== undefined ? externalSelectedEngine : internalSelectedEngine;

  const setSearchQuery = useCallback((query: string) => {
    if (onSearchQueryChange) { onSearchQueryChange(query); } else { setInternalSearchQuery(query); }
  }, [onSearchQueryChange]);
  const setSelectedEngine = useCallback((engineId: string) => {
    if (onEngineChange) { onEngineChange(engineId); } else { setInternalSelectedEngine(engineId); }
  }, [onEngineChange]);

  const [showEngineSelector, setShowEngineSelector] = useState(false);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [editingEngine, setEditingEngine] = useState<(APISearchEngine) | null>(null);
  const [deletingEngine, setDeletingEngine] = useState<SearchEngine | null>(null);
  const { data: apiEngines = [] } = useSearchEngines();
  const createSearchEngine = useCreateSearchEngine();
  const updateSearchEngine = useUpdateSearchEngine();
  const deleteSearchEngine = useDeleteSearchEngine();
  const searchEngines = apiEngines.length > 0
    ? apiEngines.map((engine, index) => convertAPIEngineToLocal(engine, index, t))
    : getDefaultSearchEngines(t);

  useEffect(() => {
    const saved = localStorage.getItem('preferred-search-engine');
    if (saved && searchEngines.find(e => e.id === saved)) {
      queueMicrotask(() => setSelectedEngine(saved));
    }
  }, [searchEngines, setSelectedEngine]);

  const currentEngine = searchEngines.find(e => e.id === selectedEngine) || searchEngines[0];

  const handleSearch = () => {
    if (searchQuery.trim()) {
      let searchUrl;
      if (currentEngine.isCustom) {
        const template = currentEngine.url + '{query}';
        searchUrl = template.replace('{query}', encodeURIComponent(searchQuery.trim()));
      } else {
        searchUrl = currentEngine.url + encodeURIComponent(searchQuery.trim());
      }
      openLink(searchUrl);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') { handleSearch(); }
  };

  const selectEngine = (engineId: string) => {
    setSelectedEngine(engineId);
    localStorage.setItem('preferred-search-engine', engineId);
    setShowEngineSelector(false);
  };

  const handleAddSearchEngine = async (data: SearchEngineFormData) => {
    try {
      if (editingEngine) {
        await updateSearchEngine.mutateAsync({ id: Number(editingEngine.id), data });
        setEditingEngine(null);
      } else {
        const newEngine = await createSearchEngine.mutateAsync(data);
        const convertedEngine = convertAPIEngineToLocal(newEngine, searchEngines.length, t);
        setSelectedEngine(convertedEngine.id);
        localStorage.setItem('preferred-search-engine', convertedEngine.id);
      }
    } catch (error) {
      console.error('Failed to save search engine:', error);
      toast({
        title: editingEngine ? t('search.editFailed') : t('search.addFailed'),
        description: t('search.retryConfig'),
        variant: 'destructive',
      });
    }
  };

  const handleEditEngine = (engine: SearchEngine) => {
    const apiEngine = searchEngines.find(e => e.id === engine.id);
    if (apiEngine && apiEngine.isCustom) {
      setEditingEngine({
        id: parseInt(engine.id),
        name: engine.name,
        url_template: engine.url + '{query}',
        icon: engine.iconName,
        description: '',
        color: engine.color,
      });
      setShowAddDialog(true);
      setShowEngineSelector(false);
    }
  };

  const handleDeleteEngine = async () => {
    if (!deletingEngine) return;
    try {
      await deleteSearchEngine.mutateAsync(parseInt(deletingEngine.id));
      if (selectedEngine === deletingEngine.id) {
        const remainingEngines = searchEngines.filter(engine => engine.id !== deletingEngine.id);
        if (remainingEngines.length > 0) {
          setSelectedEngine(remainingEngines[0].id);
          localStorage.setItem('preferred-search-engine', remainingEngines[0].id);
        }
      }
      setDeletingEngine(null);
    } catch (error) {
      console.error('Failed to delete search engine:', error);
      toast({
        title: t('search.deleteFailed'),
        description: t('search.retryLater'),
        variant: 'destructive',
      });
    }
  };

  return (
    <div className="w-full max-w-4xl mx-auto pointer-events-auto">
      <div className="flex gap-3 items-center">
        <div className="relative flex-shrink-0">
          <Button
            variant="outline"
            onClick={() => setShowEngineSelector(!showEngineSelector)}
            className="flex items-center gap-2 px-3 h-12 rounded-xl border border-gray-200 hover:border-gray-300 hover:shadow-sm transition-all text-sm bg-white/80 min-w-[160px]"
          >
            <div className="p-1.5 rounded-lg text-white shadow-sm" style={{ backgroundColor: currentEngine.color }}>
              {React.createElement(getIconComponent(currentEngine.iconName, currentEngine.name), { className: "w-4 h-4" })}
            </div>
            <div className="flex-1 text-left">
              <div className="font-medium text-foreground text-sm truncate">{currentEngine.name}</div>
            </div>
            <ChevronDown className={`h-4 w-4 text-muted-foreground transition-transform duration-200 flex-shrink-0 ${showEngineSelector ? 'rotate-180' : ''}`} />
          </Button>

          {showEngineSelector && (
            <div
              className="absolute top-full left-0 mt-2 bg-white border border-gray-200 rounded-xl shadow-xl z-[120] overflow-hidden backdrop-blur-sm min-w-[280px]"
              data-search-ui
            >
              {searchEngines.map((engine) => (
                <div
                  key={engine.id}
                  onClick={() => selectEngine(engine.id)}
                  className={`group flex items-center gap-3 p-3 cursor-pointer transition-all ${
                    selectedEngine === engine.id ? 'text-white shadow-lg' : 'hover:bg-gray-100'
                  }`}
                  style={selectedEngine === engine.id ? { backgroundColor: engine.color } : {}}
                >
                  {React.createElement(getIconComponent(engine.iconName, engine.name), { className: "w-5 h-5" })}
                  <div className="flex-1">
                    <div className="font-medium">{engine.name}</div>
                    <div className={`text-sm ${selectedEngine === engine.id ? 'text-white/80' : 'text-gray-500'}`}>
                      {getEngineDescription(engine, t)}
                    </div>
                  </div>

                  {engine.isCustom && (
                    <div className="flex items-center gap-1 ml-2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={(e) => { e.stopPropagation(); handleEditEngine(engine); }}
                        className={`p-1.5 rounded-md transition-all duration-200 ${
                          selectedEngine === engine.id ? 'text-white/70 hover:text-white hover:bg-white/20' : 'text-gray-400 hover:text-blue-600 hover:bg-blue-50'
                        }`}
                        title={t('search.editEngine')}
                      >
                        <Edit className="w-3.5 h-3.5" />
                      </button>
                      <button
                        onClick={(e) => { e.stopPropagation(); setDeletingEngine(engine); }}
                        className={`p-1.5 rounded-md transition-all duration-200 ${
                          selectedEngine === engine.id ? 'text-white/70 hover:text-white hover:bg-white/20' : 'text-gray-400 hover:text-red-600 hover:bg-red-50'
                        }`}
                        title={t('search.deleteEngine')}
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  )}
                </div>
              ))}

              <div
                onClick={() => { setShowEngineSelector(false); setShowAddDialog(true); }}
                className="flex items-center gap-3 p-3 m-2 rounded-lg cursor-pointer transition-all hover:bg-blue-50 border border-dashed border-blue-200 hover:border-blue-300"
              >
                <div className="p-2 rounded-lg bg-blue-100 text-blue-600"><Plus className="w-4 h-4" /></div>
                <div>
                  <div className="font-medium text-blue-600">{t('search.addEngine')}</div>
                  <div className="text-xs text-blue-500">{t('search.addEngineHint')}</div>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="relative flex-1">
          <Input
            type="text"
            placeholder={currentEngine.placeholder}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyPress={handleKeyPress}
            className="h-12 text-base px-4 pr-16 rounded-xl border border-gray-200 shadow-sm transition-all duration-200 focus:shadow-md focus:border-blue-500 bg-white/80"
          />
          <Button
            onClick={handleSearch}
            size="lg"
            disabled={!searchQuery.trim()}
            className="absolute right-1 top-1 h-10 px-4 rounded-lg hover:shadow-lg transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            style={{ backgroundColor: currentEngine.color }}
          >
            <Search className="h-5 w-5" />
          </Button>
        </div>
      </div>

      <SearchEngineDialog
        open={showAddDialog}
        onOpenChange={(open) => { setShowAddDialog(open); if (!open) { setEditingEngine(null); } }}
        onSubmit={handleAddSearchEngine}
        editData={editingEngine ? { name: editingEngine.name, url_template: editingEngine.url_template || '', icon: editingEngine.icon || '', description: editingEngine.description || '', color: editingEngine.color || '', id: Number(editingEngine.id) } : undefined}
      />
      <Dialog open={!!deletingEngine} onOpenChange={(open) => { if (!open) setDeletingEngine(null); }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('search.deleteTitle')}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {t('search.confirmDelete', { name: deletingEngine?.name })}
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeletingEngine(null)}>{t('search.cancel')}</Button>
            <Button variant="destructive" onClick={handleDeleteEngine}>{t('search.delete')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
