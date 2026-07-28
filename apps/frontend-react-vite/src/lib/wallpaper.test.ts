import { describe, expect, it } from 'vitest'
import { isLegacyWallpaperUrl } from './wallpaper'

describe('wallpaper URL migration', () => {
  it('detects the retired bing.img.run provider', () => {
    expect(isLegacyWallpaperUrl('https://bing.img.run/1920x1080.php')).toBe(true)
    expect(isLegacyWallpaperUrl('https://www.bing.com/th?id=example')).toBe(false)
    expect(isLegacyWallpaperUrl(null)).toBe(false)
  })
})
