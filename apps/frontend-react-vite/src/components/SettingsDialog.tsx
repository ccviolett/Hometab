import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { Label } from './ui/label';
import { RadioGroup, RadioGroupItem } from './ui/radio-group';
import { ExternalLink, Target, Replace } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { SUPPORTED_LANGUAGES, setLanguage, type AppLanguage } from '@/i18n';

interface SettingsDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

type LinkOpenMode = 'new-tab' | 'current-tab' | 'new-window';

export function SettingsDialog({ isOpen, onOpenChange }: SettingsDialogProps) {
  const { t, i18n } = useTranslation();
  const [linkOpenMode, setLinkOpenMode] = useState<LinkOpenMode>(() => {
    if (typeof window === 'undefined') return 'new-tab';
    const savedMode = localStorage.getItem('link-open-mode') as LinkOpenMode | null;
    return savedMode || 'new-tab';
  });
  const handleModeChange = (mode: LinkOpenMode) => {
    setLinkOpenMode(mode);
    localStorage.setItem('link-open-mode', mode);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ExternalLink className="h-5 w-5" />
            {t('settings.title')}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6">
          {/* Link Open Mode */}
          <div className="space-y-3">
            <Label className="text-base font-medium">{t('settings.linkOpenModeLabel')}</Label>
            <p className="text-sm text-muted-foreground">
              {t('settings.linkOpenModeHint')}
            </p>

            <RadioGroup
              value={linkOpenMode}
              onValueChange={handleModeChange}
              className="space-y-3"
            >
              <div className="flex items-center space-x-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors">
                <RadioGroupItem value="new-tab" id="new-tab" />
                <Label htmlFor="new-tab" className="flex items-center gap-2 cursor-pointer flex-1">
                  <Target className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <div className="font-medium">{t('settings.newTabLabel')}</div>
                    <div className="text-sm text-muted-foreground">{t('settings.newTabDesc')}</div>
                  </div>
                </Label>
              </div>

              <div className="flex items-center space-x-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors">
                <RadioGroupItem value="current-tab" id="current-tab" />
                <Label htmlFor="current-tab" className="flex items-center gap-2 cursor-pointer flex-1">
                  <Replace className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <div className="font-medium">{t('settings.currentTabLabel')}</div>
                    <div className="text-sm text-muted-foreground">{t('settings.currentTabDesc')}</div>
                  </div>
                </Label>
              </div>

              <div className="flex items-center space-x-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors">
                <RadioGroupItem value="new-window" id="new-window" />
                <Label htmlFor="new-window" className="flex items-center gap-2 cursor-pointer flex-1">
                  <ExternalLink className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <div className="font-medium">{t('settings.newWindowLabel')}</div>
                    <div className="text-sm text-muted-foreground">{t('settings.newWindowDesc')}</div>
                  </div>
                </Label>
              </div>
            </RadioGroup>
          </div>

          <div className="space-y-3">
            <Label className="text-base font-medium">{t('settings.languageLabel')}</Label>
            <p className="text-sm text-muted-foreground">
              {t('settings.languageHint')}
            </p>

            <RadioGroup
              value={i18n.language}
              onValueChange={(value) => setLanguage(value as AppLanguage)}
              className="space-y-3"
            >
              {SUPPORTED_LANGUAGES.map((l) => (
                <div key={l.code} className="flex items-center space-x-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors">
                  <RadioGroupItem value={l.code} id={`lang-${l.code}`} />
                  <Label htmlFor={`lang-${l.code}`} className="cursor-pointer flex-1">{l.label}</Label>
                </div>
              ))}
            </RadioGroup>
          </div>

          <div className="pt-4 border-t">
            <p className="text-xs text-muted-foreground">
              {t('settings.autoSaveHint')}
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
