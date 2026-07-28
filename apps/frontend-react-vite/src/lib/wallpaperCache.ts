import { isLegacyWallpaperUrl } from './wallpaper';

interface WallpaperCacheData {
  url: string;
  title?: string;
  copyright?: string;
  timestamp: number;
  type: 'daily' | 'random';
}

const CACHE_KEY_PREFIX = 'wallpaper-cache-';
const DAILY_CACHE_DURATION = 60 * 60 * 1000; // 1小时缓存
const RANDOM_CACHE_DURATION = 24 * 60 * 60 * 1000; // 24小时缓存

export class WallpaperCache {
  /**
   * 获取缓存的壁纸数据
   */
  static get(type: 'daily' | 'random', resolution: string): WallpaperCacheData | null {
    const cacheKey = `${CACHE_KEY_PREFIX}${type}-${resolution}`;
    const cached = localStorage.getItem(cacheKey);

    if (!cached) return null;

    try {
      const data: WallpaperCacheData = JSON.parse(cached);
      if (isLegacyWallpaperUrl(data.url)) {
        localStorage.removeItem(cacheKey);
        return null;
      }
      const now = Date.now();
      const maxAge = type === 'daily' ? DAILY_CACHE_DURATION : RANDOM_CACHE_DURATION;

      // 检查缓存是否过期
      if (now - data.timestamp > maxAge) {
        localStorage.removeItem(cacheKey);
        return null;
      }

      return data;
    } catch {
      localStorage.removeItem(cacheKey);
      return null;
    }
  }

  /**
   * 设置壁纸缓存
   */
  static set(type: 'daily' | 'random', resolution: string, data: Omit<WallpaperCacheData, 'timestamp' | 'type'>) {
    const cacheKey = `${CACHE_KEY_PREFIX}${type}-${resolution}`;
    const cacheData: WallpaperCacheData = {
      ...data,
      type,
      timestamp: Date.now()
    };

    localStorage.setItem(cacheKey, JSON.stringify(cacheData));
  }

  /**
   * 清除所有壁纸缓存
   */
  static clear() {
    const keys = Object.keys(localStorage);
    keys.forEach(key => {
      if (key.startsWith(CACHE_KEY_PREFIX)) {
        localStorage.removeItem(key);
      }
    });
  }

  /**
   * 清除过期的缓存
   */
  static clearExpired() {
    const keys = Object.keys(localStorage);
    const now = Date.now();

    keys.forEach(key => {
      if (key.startsWith(CACHE_KEY_PREFIX)) {
        try {
          const data: WallpaperCacheData = JSON.parse(localStorage.getItem(key) || '');
          const maxAge = data.type === 'daily' ? DAILY_CACHE_DURATION : RANDOM_CACHE_DURATION;

          if (now - data.timestamp > maxAge) {
            localStorage.removeItem(key);
          }
        } catch {
          localStorage.removeItem(key);
        }
      }
    });
  }
}
