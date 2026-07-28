import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Search, Globe, Building, Plus } from 'lucide-react';

interface SearchEngineDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: SearchEngineFormData) => void;
  editData?: SearchEngineFormData & { id?: number };
  mode?: 'create' | 'edit';
}

export interface SearchEngineFormData {
  name: string;
  url_template: string;
  icon: string;
  description: string;
  color: string;
}

const iconOptions = [
  { value: 'search', label: 'Search', icon: Search },
  { value: 'globe', label: 'Globe', icon: Globe },
  { value: 'building', label: 'Building', icon: Building },
  { value: 'plus', label: 'Plus', icon: Plus },
];

const colorOptions = [
  { value: '#6b7280', i18nKey: 'searchengine.colorGray', preview: '#6b7280' },
  { value: '#059669', i18nKey: 'searchengine.colorGreen', preview: '#059669' },
  { value: '#dc2626', i18nKey: 'searchengine.colorRed', preview: '#dc2626' },
  { value: '#2563eb', i18nKey: 'searchengine.colorBlue', preview: '#2563eb' },
  { value: '#7c3aed', i18nKey: 'searchengine.colorPurple', preview: '#7c3aed' },
  { value: '#ea580c', i18nKey: 'searchengine.colorOrange', preview: '#ea580c' },
  { value: '#0891b2', i18nKey: 'searchengine.colorCyan', preview: '#0891b2' },
  { value: '#be123c', i18nKey: 'searchengine.colorRose', preview: '#be123c' },
];

export function SearchEngineDialog({ open, onOpenChange, onSubmit, editData, mode = 'create' }: SearchEngineDialogProps) {
  const { t } = useTranslation();
  const [formData, setFormData] = useState<SearchEngineFormData>({
    name: '',
    url_template: '',
    icon: 'search',
    description: '',
    color: '#6b7280',
  });

  const [errors, setErrors] = useState<Partial<SearchEngineFormData>>({});

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const newErrors: Partial<SearchEngineFormData> = {};

    if (!formData.name.trim()) {
      newErrors.name = t('searchengine.errorNameRequired');
    }

    if (!formData.url_template.trim()) {
      newErrors.url_template = t('searchengine.errorUrlRequired');
    } else if (!formData.url_template.includes('{query}')) {
      newErrors.url_template = t('searchengine.errorUrlPlaceholder', { query: '{query}' });
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    onSubmit(formData);
    handleClose();
  };

  React.useEffect(() => {
    if (editData && mode === 'edit') {
      setFormData({
        name: editData.name,
        url_template: editData.url_template,
        icon: editData.icon,
        description: editData.description,
        color: editData.color,
      });
    }
  }, [editData, mode]);

  const handleClose = () => {
    setFormData({
      name: '',
      url_template: '',
      icon: 'search',
      description: '',
      color: '#6b7280',
    });
    setErrors({});
    onOpenChange(false);
  };

  const handleInputChange = (field: keyof SearchEngineFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{mode === 'edit' ? t('searchengine.editTitle') : t('searchengine.addTitle')}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">{t('searchengine.nameLabel')}</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              placeholder={t('searchengine.namePlaceholder')}
              className={errors.name ? 'border-red-500' : ''}
            />
            {errors.name && <p className="text-sm text-red-500">{errors.name}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="url_template">{t('searchengine.urlLabel')}</Label>
            <Input
              id="url_template"
              value={formData.url_template}
              onChange={(e) => handleInputChange('url_template', e.target.value)}
              placeholder="https://www.baidu.com/s?wd={query}"
              className={errors.url_template ? 'border-red-500' : ''}
            />
            <p className="text-xs text-gray-500">
              {t('searchengine.queryPlaceholderHint', { query: '{query}' })}
            </p>
            {errors.url_template && <p className="text-sm text-red-500">{errors.url_template}</p>}
          </div>

          <div className="space-y-2">
            <Label htmlFor="icon">{t('searchengine.iconLabel')}</Label>
            <div className="grid grid-cols-4 gap-2">
              {iconOptions.map((option) => {
                const IconComponent = option.icon;
                return (
                  <button
                    key={option.value}
                    type="button"
                    onClick={() => handleInputChange('icon', option.value)}
                    className={`p-3 rounded-lg border-2 flex flex-col items-center gap-1 transition-colors ${
                      formData.icon === option.value
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    <IconComponent className="w-5 h-5" />
                    <span className="text-xs">{option.label}</span>
                  </button>
                );
              })}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="color">{t('searchengine.colorLabel')}</Label>
            <div className="grid grid-cols-4 gap-2">
              {colorOptions.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => handleInputChange('color', option.value)}
                  className={`p-3 rounded-lg border-2 flex flex-col items-center gap-1 transition-colors ${
                    formData.color === option.value
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <div
                    className="w-6 h-6 rounded-full border border-gray-300"
                    style={{ backgroundColor: option.preview }}
                  />
                  <span className="text-xs">{t(option.i18nKey)}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">{t('searchengine.descriptionLabel')}</Label>
            <Input
              id="description"
              value={formData.description}
              onChange={(e) => handleInputChange('description', e.target.value)}
              placeholder={t('searchengine.descriptionPlaceholder')}
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose}>
              {t('searchengine.cancel')}
            </Button>
            <Button type="submit">
              {mode === 'edit' ? t('searchengine.save') : t('searchengine.addTitle')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
