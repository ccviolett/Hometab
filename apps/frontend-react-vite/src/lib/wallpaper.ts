// 壁纸管理工具函数

// 默认壁纸 - 本地自托管资产（slate/blue 极简风，离线可用，零外部依赖）
export const DEFAULT_WALLPAPER = `${import.meta.env.BASE_URL}wallpaper-default.jpg`;

export function isLegacyWallpaperUrl(url: string | null): boolean {
  if (!url) return false;
  try {
    return new URL(url, 'http://localhost').hostname === 'bing.img.run';
  } catch {
    return false;
  }
}

export const initializeWallpaper = () => {
  const savedUrl = localStorage.getItem('wallpaper-url');
  if (isLegacyWallpaperUrl(savedUrl)) {
    localStorage.removeItem('wallpaper-url');
    localStorage.removeItem('wallpaper-custom-url');
  }
  const wallpaperUrl = savedUrl && !isLegacyWallpaperUrl(savedUrl) ? savedUrl : DEFAULT_WALLPAPER;

  document.body.style.backgroundImage = `url(${wallpaperUrl})`;
  document.body.style.backgroundSize = 'cover';
  document.body.style.backgroundPosition = 'center';
  document.body.style.backgroundRepeat = 'no-repeat';
  document.body.style.backgroundAttachment = 'fixed';
};

export const applyWallpaper = (url: string) => {
  document.body.style.backgroundImage = `url(${url})`;
  document.body.style.backgroundSize = 'cover';
  document.body.style.backgroundPosition = 'center';
  document.body.style.backgroundRepeat = 'no-repeat';
  document.body.style.backgroundAttachment = 'fixed';
  localStorage.setItem('wallpaper-url', url);
};

export const removeWallpaper = () => {
  document.body.style.backgroundImage = '';
  localStorage.removeItem('wallpaper-type');
  localStorage.removeItem('wallpaper-custom-url');
  localStorage.removeItem('wallpaper-url');
};
