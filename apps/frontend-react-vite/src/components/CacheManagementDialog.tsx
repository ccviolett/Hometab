import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { Button } from './ui/button';
import { RefreshCw, Trash2, HardDrive, Database } from 'lucide-react';
import { WallpaperCache } from '@/lib/wallpaperCache';
import IconService, { type IconRefreshAllResult } from '@/lib/iconService';
import { IconManagementPanel } from './IconManagementPanel';

interface CacheManagementDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CacheManagementDialog({ isOpen, onOpenChange }: CacheManagementDialogProps) {
  const [isClearing, setIsClearing] = useState(false);
  const [isRefreshingIcons, setIsRefreshingIcons] = useState(false);
  const [clearSuccess, setClearSuccess] = useState(false);
  const [iconRefreshResult, setIconRefreshResult] = useState<IconRefreshAllResult | null>(null);
  const [iconRefreshVersion, setIconRefreshVersion] = useState(0);
  const [cacheStats, setCacheStats] = useState<{
    wallpaperCached: boolean;
  } | null>(null);

  const { t } = useTranslation();

  const calculateCacheStats = () => {
    const wallpaperCached = !!localStorage.getItem('wallpaper-cache');

    setCacheStats({
      wallpaperCached,
    });
  };

  useEffect(() => {
    if (isOpen) {
      calculateCacheStats();
    }
  }, [isOpen]);

  const handleClearCache = async () => {
    setIsClearing(true);
    setClearSuccess(false);

    try {
      WallpaperCache.clear();
      calculateCacheStats();
      setClearSuccess(true);

      setTimeout(() => {
        window.location.reload();
      }, 2000);
    } catch (error) {
      console.error('Failed to clear cache:', error);
    } finally {
      setIsClearing(false);
    }
  };

  const handleRefreshIcons = async () => {
    setIsRefreshingIcons(true);
    setIconRefreshResult(null);
    try {
      const result = await IconService.refreshAll();
      setIconRefreshResult(result);
      setIconRefreshVersion((version) => version + 1);
    } catch (error) {
      console.error('Failed to refresh icons:', error);
    } finally {
      setIsRefreshingIcons(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <HardDrive className="h-5 w-5" />
            {t('cache.title')}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="p-4 rounded-lg border bg-gray-50/50 space-y-3">
            <div className="flex items-center gap-2">
              <Database className="h-4 w-4 text-muted-foreground" />
              <span className="font-medium text-sm">{t('cache.currentStatus')}</span>
            </div>

            {cacheStats && (
              <div className="grid grid-cols-1 gap-4 text-sm">
                <div className="space-y-1">
                  <div className="text-muted-foreground">{t('cache.wallpaperCache')}</div>
                  <div className="text-lg font-semibold text-primary">
                    {cacheStats.wallpaperCached ? t('cache.cached') : t('cache.notCached')}
                  </div>
                </div>
              </div>
            )}
          </div>

          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              {t('cache.clearHint')}
            </p>

            <div className="p-4 rounded-lg border space-y-3">
              <div className="flex items-start space-x-3">
                <RefreshCw className="h-5 w-5 text-muted-foreground mt-0.5 flex-shrink-0" />
                <div className="flex-1 space-y-3">
                  <div>
                    <div className="font-medium">{t('cache.refreshIcons')}</div>
                    <div className="text-sm text-muted-foreground mt-1">
                      {t('cache.refreshIconsDesc')}
                    </div>
                  </div>

                  {iconRefreshResult && (
                    <div className="text-xs text-muted-foreground bg-blue-50 p-3 rounded-md">
                      {t('cache.iconResult', {
                        total_hosts: iconRefreshResult.total_hosts,
                        ready: iconRefreshResult.ready,
                        failed: iconRefreshResult.failed,
                        conflicts: iconRefreshResult.conflicts,
                      })}
                    </div>
                  )}

                  <Button
                    onClick={handleRefreshIcons}
                    disabled={isRefreshingIcons}
                    variant="outline"
                    className="w-full"
                  >
                    <RefreshCw className={`h-4 w-4 mr-2 ${isRefreshingIcons ? 'animate-spin' : ''}`} />
                    {isRefreshingIcons ? t('cache.refreshingIcons') : t('cache.refreshIcons')}
                  </Button>
                </div>
              </div>
            </div>

            <div className="p-4 rounded-lg border space-y-3">
              <div className="flex items-start space-x-3">
                <Trash2 className="h-5 w-5 text-muted-foreground mt-0.5 flex-shrink-0" />
                <div className="flex-1 space-y-3">
                  <div>
                    <div className="font-medium">{t('cache.clearBrowserCache')}</div>
                    <div className="text-sm text-muted-foreground mt-1">
                      {t('cache.clearBrowserCacheDesc')}
                    </div>
                  </div>

                  <div className="space-y-2 text-sm text-muted-foreground">
                    <div className="flex items-start gap-2">
                      <span className="text-amber-500">•</span>
                      <span>{t('cache.wallpaperCache')}</span>
                    </div>
                    <div className="flex items-start gap-2">
                      <span className="text-amber-500">•</span>
                      <span>{t('cache.nextVisitReload')}</span>
                    </div>
                  </div>

                  <Button
                    onClick={handleClearCache}
                    disabled={isClearing}
                    variant={clearSuccess ? "default" : "destructive"}
                    className="w-full"
                  >
                    {isClearing ? (
                      <>
                        <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                        {t('cache.clearing')}
                      </>
                    ) : clearSuccess ? (
                      <>
                        {t('cache.clearedReloading')}
                      </>
                    ) : (
                      <>
                        <Trash2 className="h-4 w-4 mr-2" />
                        {t('cache.clearAllCache')}
                      </>
                    )}
                  </Button>
                </div>
              </div>
            </div>

            <IconManagementPanel refreshVersion={iconRefreshVersion} />

            <div className="text-xs text-muted-foreground bg-amber-50 dark:bg-amber-900/20 p-3 rounded-lg">
              <strong>{t('cache.note')}</strong> {t('cache.clearWarning')}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
