import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import i18n from '@/i18n';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { Info, Clock, Code, Package } from 'lucide-react';

interface AboutDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

interface BuildInfo {
  version: string;
  build_time: string;
  go_version: string;
}

export function AboutDialog({ isOpen, onOpenChange }: AboutDialogProps) {
  const { t } = useTranslation();
  const [buildInfo, setBuildInfo] = useState<BuildInfo | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    queueMicrotask(() => setLoading(true));
    fetch('/api/build-info')
      .then(res => res.json())
      .then(data => setBuildInfo(data))
      .catch(() => setBuildInfo(null))
      .finally(() => setLoading(false));
  }, [isOpen]);

  const formatBuildTime = (raw: string): string => {
    if (!raw || raw === 'unknown') return t('about.devMode');
    try {
      const d = new Date(raw.replace(' UTC', 'Z').replace(' ', 'T'));
      return d.toLocaleString(i18n.language, {
        year: 'numeric', month: '2-digit', day: '2-digit',
        hour: '2-digit', minute: '2-digit', second: '2-digit',
      });
    } catch {
      return raw;
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Info className="h-5 w-5" />
            {t('about.title')}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="text-center py-2">
            <h3 className="text-xl font-bold">Hometab</h3>
            <p className="text-sm text-muted-foreground">{t('about.subtitle')}</p>
          </div>

          {loading ? (
            <p className="text-sm text-center text-muted-foreground">{t('about.loading')}</p>
          ) : buildInfo ? (
            <div className="space-y-3 rounded-lg border p-4 text-sm">
              <div className="flex items-center justify-between">
                <span className="flex items-center gap-2 text-muted-foreground">
                  <Package className="h-3.5 w-3.5" /> {t('about.version')}
                </span>
                <span className="font-mono font-medium">{buildInfo.version}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="flex items-center gap-2 text-muted-foreground">
                  <Clock className="h-3.5 w-3.5" /> {t('about.buildTime')}
                </span>
                <span className="font-mono font-medium">{formatBuildTime(buildInfo.build_time)}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="flex items-center gap-2 text-muted-foreground">
                  <Code className="h-3.5 w-3.5" /> {t('about.runtime')}
                </span>
                <span className="font-mono font-medium">{buildInfo.go_version}</span>
              </div>
            </div>
          ) : (
            <p className="text-sm text-center text-muted-foreground">{t('about.unableToGetBuildInfo')}</p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
