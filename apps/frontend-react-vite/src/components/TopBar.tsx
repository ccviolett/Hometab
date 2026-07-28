import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Settings, Download, Image, HardDrive, Info, Globe, Check } from 'lucide-react';
import { SUPPORTED_LANGUAGES, setLanguage } from '@/i18n';
import DataManagement from './DataManagement';
import WallpaperDialog from './WallpaperDialog';
import { SettingsDialog } from './SettingsDialog';
import { CacheManagementDialog } from './CacheManagementDialog';
import { AboutDialog } from './AboutDialog';

export default function TopBar() {
  const { t, i18n } = useTranslation();
  const [isDataManagementOpen, setIsDataManagementOpen] = useState(false);
  const [wallpaperOpen, setWallpaperOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [cacheManageOpen, setCacheManageOpen] = useState(false);
  const [aboutOpen, setAboutOpen] = useState(false);

  return (
    <>
      <div className="flex items-center justify-between w-full">
        <div className="flex items-center gap-3">
          <button
            onClick={() => window.location.reload()}
            className="text-lg font-semibold hover:text-blue-600 transition-colors cursor-pointer"
          >
            Hometab
          </button>
        </div>

        <div className="flex items-center gap-2">
          <DropdownMenu modal={false}>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" title={t('topbar.language')}>
                <Globe className="h-4 w-4" />
                <span className="ml-1 text-xs">{i18n.language === 'en' ? 'EN' : '中'}</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-40">
              {SUPPORTED_LANGUAGES.map((l) => (
                <DropdownMenuItem
                  key={l.code}
                  onClick={() => setLanguage(l.code)}
                  className="cursor-pointer flex items-center justify-between"
                >
                  <span>{l.label}</span>
                  {i18n.language === l.code && <Check className="h-4 w-4" />}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
          <DropdownMenu modal={false}>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm">
                <Settings className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuItem onClick={() => setWallpaperOpen(true)} className="cursor-pointer">
                <Image className="h-4 w-4 mr-2" />
                {t('topbar.wallpaperSettings')}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => setIsDataManagementOpen(true)} className="cursor-pointer">
                <Download className="h-4 w-4 mr-2" />
                {t('topbar.dataManagement')}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setCacheManageOpen(true)} className="cursor-pointer">
                <HardDrive className="h-4 w-4 mr-2" />
                {t('topbar.cacheManagement')}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => setSettingsOpen(true)} className="cursor-pointer">
                <Settings className="h-4 w-4 mr-2" />
                {t('topbar.settings')}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => setAboutOpen(true)} className="cursor-pointer">
                <Info className="h-4 w-4 mr-2" />
                {t('topbar.about')}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <Dialog open={isDataManagementOpen} onOpenChange={setIsDataManagementOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('topbar.dataManagement')}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              {t('topbar.dataManagementDesc')}
            </p>
            <DataManagement />
          </div>
        </DialogContent>
      </Dialog>

      <WallpaperDialog open={wallpaperOpen} onOpenChange={setWallpaperOpen} />
      <SettingsDialog isOpen={settingsOpen} onOpenChange={setSettingsOpen} />
      <CacheManagementDialog isOpen={cacheManageOpen} onOpenChange={setCacheManageOpen} />
      <AboutDialog isOpen={aboutOpen} onOpenChange={setAboutOpen} />
    </>
  );
}
