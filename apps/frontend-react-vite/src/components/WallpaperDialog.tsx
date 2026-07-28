import { useCallback, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Image, Globe, Loader2 } from 'lucide-react';
import { WallpaperCache } from '@/lib/wallpaperCache';

interface WallpaperDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface BingWallpaper {
  url: string;
  title: string;
  copyright: string;
}

export default function WallpaperDialog({ open, onOpenChange }: WallpaperDialogProps) {
  const { t } = useTranslation();
  const [wallpaperType, setWallpaperType] = useState<'bing' | 'bing-random' | 'custom'>('bing');
  const [customUrl, setCustomUrl] = useState('');
  const [bingWallpaper, setBingWallpaper] = useState<BingWallpaper | null>(null);
  const [loading, setLoading] = useState(false);
  const [previewUrl, setPreviewUrl] = useState('');
  const [resolution, setResolution] = useState<'uhd' | '1920x1080' | '1366x768' | 'mobile'>('1920x1080');

  const fetchWallpaperMetadata = useCallback(async (type: 'daily' | 'random', selectedResolution: string): Promise<BingWallpaper> => {
    const resolutionMap: Record<string, string> = {
      uhd: 'UHD',
      '1920x1080': '1920',
      '1366x768': '1366',
      mobile: '1080',
    };
    const index = type === 'random' ? Math.floor(Math.random() * 8) : 0;
    const params = new URLSearchParams({
      resolution: resolutionMap[selectedResolution] || '1920',
      format: 'json',
      index: String(index),
      mkt: 'zh-CN',
    });
    const response = await fetch(`https://bing.biturl.top/?${params.toString()}`);
    if (!response.ok) throw new Error(`Bing wallpaper API returned ${response.status}`);
    const data = await response.json() as { url?: string; title?: string; copyright?: string };
    if (!data.url) throw new Error('Bing wallpaper API returned no image URL');
    return {
      url: data.url,
      title: data.title || (type === 'daily' ? t('wallpaper.bingDaily') : t('wallpaper.bingRandom')),
      copyright: data.copyright || '',
    };
  }, [t]);

  const fetchBingWallpaper = useCallback(async () => {
    setLoading(true);
    try {
      const cached = WallpaperCache.get('daily', resolution);
      if (cached) {
        const wallpaper: BingWallpaper = {
          url: cached.url,
          title: cached.title || t('wallpaper.bingDaily'),
          copyright: cached.copyright || ''
        };
        setBingWallpaper(wallpaper);
        if (wallpaperType === 'bing') {
          setPreviewUrl(wallpaper.url);
        }
        setLoading(false);
        return;
      }

      const wallpaper = await fetchWallpaperMetadata('daily', resolution);
      WallpaperCache.set('daily', resolution, {
        url: wallpaper.url,
        title: wallpaper.title,
        copyright: wallpaper.copyright
      });
      setBingWallpaper(wallpaper);
      if (wallpaperType === 'bing') {
        setPreviewUrl(wallpaper.url);
      }
    } catch (error) {
      console.error('Bing wallpaper API failed:', error);
      setBingWallpaper(null);
      setPreviewUrl('');
    } finally {
      setLoading(false);
    }
  }, [fetchWallpaperMetadata, resolution, wallpaperType, t]);

  const fetchRandomBingWallpaper = useCallback(async () => {
    setLoading(true);
    try {
      const wallpaper = await fetchWallpaperMetadata('random', resolution);
      WallpaperCache.set('random', resolution, wallpaper);
      setBingWallpaper(wallpaper);
      setPreviewUrl(wallpaper.url);
    } catch (error) {
      console.error('Failed to fetch random Bing wallpaper:', error);
    } finally {
      setLoading(false);
    }
  }, [fetchWallpaperMetadata, resolution]);

  useEffect(() => {
    if (open) {
      const savedType = localStorage.getItem('wallpaper-type') as 'bing' | 'bing-random' | 'custom' || 'bing';
      const savedCustomUrl = localStorage.getItem('wallpaper-custom-url') || '';
      const savedResolution = localStorage.getItem('wallpaper-resolution') as 'uhd' | '1920x1080' | '1366x768' | 'mobile' || '1920x1080';
      setWallpaperType(savedType);
      setCustomUrl(savedCustomUrl);
      setResolution(savedResolution);
    }
  }, [open]);

  useEffect(() => {
    if (open && (wallpaperType === 'bing' || wallpaperType === 'bing-random')) {
      if (wallpaperType === 'bing') {
        fetchBingWallpaper();
      } else {
        fetchRandomBingWallpaper();
      }
    }
  }, [open, wallpaperType, resolution, fetchBingWallpaper, fetchRandomBingWallpaper]);

  useEffect(() => {
    if ((wallpaperType === 'bing' || wallpaperType === 'bing-random') && bingWallpaper) {
      setPreviewUrl(bingWallpaper.url);
    } else if (wallpaperType === 'custom' && customUrl) {
      setPreviewUrl(customUrl);
    } else {
      setPreviewUrl('');
    }
  }, [wallpaperType, customUrl, bingWallpaper]);

  const handleSave = () => {
    localStorage.setItem('wallpaper-type', wallpaperType);
    localStorage.setItem('wallpaper-resolution', resolution);
    if (wallpaperType === 'custom') {
      localStorage.setItem('wallpaper-custom-url', customUrl);
    }

    const finalUrl = (wallpaperType === 'bing' || wallpaperType === 'bing-random') ? bingWallpaper?.url : customUrl;
    if (finalUrl) {
      document.body.style.backgroundImage = `url(${finalUrl})`;
      document.body.style.backgroundSize = 'cover';
      document.body.style.backgroundPosition = 'center';
      document.body.style.backgroundRepeat = 'no-repeat';
      document.body.style.backgroundAttachment = 'fixed';
      localStorage.setItem('wallpaper-url', finalUrl);
    }

    onOpenChange(false);
  };

  const handleReset = () => {
    document.body.style.backgroundImage = '';
    localStorage.removeItem('wallpaper-type');
    localStorage.removeItem('wallpaper-custom-url');
    localStorage.removeItem('wallpaper-url');
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Image className="h-5 w-5" />
            {t('wallpaper.title')}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6">
          <div className="space-y-3">
            <Label className="text-base font-medium">{t('wallpaper.typeLabel')}</Label>
            <RadioGroup value={wallpaperType} onValueChange={(value) => setWallpaperType(value as 'bing' | 'bing-random' | 'custom')}>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="bing" id="bing" />
                <Label htmlFor="bing" className="flex items-center gap-2 cursor-pointer">
                  <Globe className="h-4 w-4" />
                  {t('wallpaper.bingDaily')}
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="bing-random" id="bing-random" />
                <Label htmlFor="bing-random" className="flex items-center gap-2 cursor-pointer">
                  <Globe className="h-4 w-4" />
                  {t('wallpaper.bingRandom')}
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="custom" id="custom" />
                <Label htmlFor="custom" className="flex items-center gap-2 cursor-pointer">
                  <Image className="h-4 w-4" />
                  {t('wallpaper.customUrl')}
                </Label>
              </div>
            </RadioGroup>
          </div>

          {(wallpaperType === 'bing' || wallpaperType === 'bing-random') && (
            <div className="space-y-3">
              <Label className="text-base font-medium">{t('wallpaper.resolution')}</Label>
              <RadioGroup value={resolution} onValueChange={(value) => setResolution(value as 'uhd' | '1920x1080' | '1366x768' | 'mobile')}>
                <div className="grid grid-cols-2 gap-2">
                  <div className="flex items-center space-x-2">
                    <RadioGroupItem value="1920x1080" id="1920x1080" />
                    <Label htmlFor="1920x1080" className="cursor-pointer text-sm">1920×1080</Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <RadioGroupItem value="1366x768" id="1366x768" />
                    <Label htmlFor="1366x768" className="cursor-pointer text-sm">1366×768</Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <RadioGroupItem value="uhd" id="uhd" />
                    <Label htmlFor="uhd" className="cursor-pointer text-sm">{t('wallpaper.resUhd')}</Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <RadioGroupItem value="mobile" id="mobile" />
                    <Label htmlFor="mobile" className="cursor-pointer text-sm">{t('wallpaper.resMobile')}</Label>
                  </div>
                </div>
              </RadioGroup>
            </div>
          )}

          {wallpaperType === 'bing' && (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label className="text-base font-medium">{t('wallpaper.today')}</Label>
                <Button variant="outline" size="sm" onClick={fetchBingWallpaper} disabled={loading}>
                  {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : t('wallpaper.refresh')}
                </Button>
              </div>
              {bingWallpaper && (
                <div className="p-3 bg-gray-50 rounded-lg">
                  <p className="font-medium text-sm">{bingWallpaper.title}</p>
                  {bingWallpaper.copyright && (
                    <p className="text-xs text-gray-600 mt-1">{bingWallpaper.copyright}</p>
                  )}
                </div>
              )}
            </div>
          )}

          {wallpaperType === 'bing-random' && (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label className="text-base font-medium">{t('wallpaper.random')}</Label>
                <Button variant="outline" size="sm" onClick={fetchRandomBingWallpaper} disabled={loading}>
                  {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : t('wallpaper.changeOne')}
                </Button>
              </div>
              {bingWallpaper && (
                <div className="p-3 bg-gray-50 rounded-lg">
                  <p className="font-medium text-sm">{bingWallpaper.title}</p>
                  <p className="text-xs text-gray-600 mt-1">{t('wallpaper.randomHint')}</p>
                </div>
              )}
            </div>
          )}

          {wallpaperType === 'custom' && (
            <div className="space-y-3">
              <Label htmlFor="custom-url" className="text-base font-medium">{t('wallpaper.urlLabel')}</Label>
              <Input
                id="custom-url"
                type="url"
                placeholder="https://example.com/wallpaper.jpg"
                value={customUrl}
                onChange={(e) => setCustomUrl(e.target.value)}
              />
              <p className="text-xs text-gray-500">{t('wallpaper.urlHint')}</p>
            </div>
          )}

          {previewUrl && (
            <div className="space-y-3">
              <Label className="text-base font-medium">{t('wallpaper.preview')}</Label>
              <div className="relative h-32 bg-gray-100 rounded-lg overflow-hidden">
                <img
                  src={previewUrl}
                  alt={t('wallpaper.previewAlt')}
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    (e.target as HTMLImageElement).hidden = true;
                    (e.target as HTMLImageElement).alt = t('wallpaper.imgLoadFailed');
                  }}
                />
              </div>
            </div>
          )}

          <div className="flex justify-between pt-4">
            <Button variant="outline" onClick={handleReset}>{t('wallpaper.remove')}</Button>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => onOpenChange(false)}>{t('wallpaper.cancel')}</Button>
              <Button onClick={handleSave} disabled={wallpaperType === 'custom' && !customUrl.trim()}>
                {t('wallpaper.apply')}
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
