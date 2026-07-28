import { useState, useEffect, useRef, useCallback } from 'react'
import TopBar from '@/components/TopBar'
import SearchSection from '@/components/SearchSection'
import ExternalRequestsPanel from '@/components/ExternalRequestsPanel'
import { LinksSection } from '@/components/links/LinksSection'
import { initializeWallpaper } from '@/lib/wallpaper'
import { useTranslation } from 'react-i18next'

function getInitialWallpaperInfo() {
  if (typeof window === 'undefined') return null
  const wallpaperType = localStorage.getItem('wallpaper-type')
  if (wallpaperType === 'bing' || wallpaperType === 'bing-random') {
    return {
      titleKey: wallpaperType === 'bing-random' ? 'wallpaper.bingRandom' : 'wallpaper.bingDaily',
      copyright: '',
    }
  }
  return null
}

function getPrefersReducedMotion() {
  if (typeof window === 'undefined') return false
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

export default function HomePage() {
  const [layoutMode, setLayoutMode] = useState<'normal' | 'focused'>('normal')
  const [wallpaperInfo] = useState<{ titleKey?: string; copyright?: string } | null>(getInitialWallpaperInfo)
  const { t } = useTranslation()
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedEngine, setSelectedEngine] = useState('google')
  const [searchOffsetY, setSearchOffsetY] = useState(0)
  const [contentGlassReady, setContentGlassReady] = useState(true)
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(getPrefersReducedMotion)
  const isTransitioningRef = useRef(false)
  const searchRef = useRef<HTMLDivElement>(null)
  const searchOffsetYRef = useRef(0)
  const contentGlassTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    initializeWallpaper()
  }, [])

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    const handleChange = () => setPrefersReducedMotion(mediaQuery.matches)
    handleChange()
    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  useEffect(() => {
    document.body.style.overflow = layoutMode === 'focused' ? 'hidden' : ''
    return () => { document.body.style.overflow = '' }
  }, [layoutMode])

  useEffect(() => {
    return () => {
      if (contentGlassTimerRef.current) {
        clearTimeout(contentGlassTimerRef.current)
      }
    }
  }, [])

  const setSearchOffset = useCallback((offset: number) => {
    searchOffsetYRef.current = offset
    setSearchOffsetY(offset)
  }, [])

  const calculateFocusedOffset = useCallback(() => {
    const rect = searchRef.current?.getBoundingClientRect()
    if (!rect) return 0
    const untransformedCenterY = rect.top + rect.height / 2 - searchOffsetYRef.current
    return window.innerHeight / 2 - untransformedCenterY
  }, [])

  const handleDoubleClick = useCallback((e: React.MouseEvent) => {
    const target = e.target as HTMLElement
    if (target.closest('button, a, input, select, textarea, [role="dialog"], [data-search-ui]')) return
    if (target.tagName !== 'DIV' && target.tagName !== 'MAIN' && target.tagName !== 'SECTION') return
    if (isTransitioningRef.current) return
    isTransitioningRef.current = true
    const nextMode = layoutMode === 'normal' ? 'focused' : 'normal'

    if (contentGlassTimerRef.current) {
      clearTimeout(contentGlassTimerRef.current)
      contentGlassTimerRef.current = null
    }
    setContentGlassReady(false)
    if (nextMode === 'normal') {
      if (prefersReducedMotion) {
        setContentGlassReady(true)
      } else {
        contentGlassTimerRef.current = setTimeout(() => {
          setContentGlassReady(true)
          contentGlassTimerRef.current = null
        }, 360)
      }
    }

    setSearchOffset(nextMode === 'focused' ? calculateFocusedOffset() : 0)
    setLayoutMode(nextMode)
    setTimeout(() => { isTransitioningRef.current = false }, prefersReducedMotion ? 120 : 800)
  }, [calculateFocusedOffset, layoutMode, prefersReducedMotion, setSearchOffset])

  const isFocused = layoutMode === 'focused'

  useEffect(() => {
    if (!isFocused) return

    const handleResize = () => {
      setSearchOffset(calculateFocusedOffset())
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [calculateFocusedOffset, isFocused, setSearchOffset])

  return (
    <div className="min-h-screen" onDoubleClick={handleDoubleClick}>
      <div
        className="fixed inset-0 z-0 pointer-events-none"
        style={{
          backdropFilter: isFocused ? 'blur(0px)' : 'blur(3px)',
          WebkitBackdropFilter: isFocused ? 'blur(0px)' : 'blur(3px)',
          backgroundColor: isFocused ? 'rgba(255, 255, 255, 0)' : 'rgba(255, 255, 255, 0.12)',
          transition: prefersReducedMotion
            ? 'none'
            : 'backdrop-filter 520ms cubic-bezier(0.22, 1, 0.36, 1), background-color 520ms cubic-bezier(0.22, 1, 0.36, 1)',
        }}
      />

      {/* TopBar */}
      <div
        className="relative z-[200] transition-all duration-700 ease-out"
        style={{
          transform: isFocused ? 'translateY(-100%)' : 'translateY(0)',
          opacity: isFocused ? 0 : 1,
        }}
      >
        <div className="sticky top-0 z-[200] w-full border-b bg-white backdrop-blur">
          <div className="container mx-auto px-4 h-14 flex items-center">
            <TopBar />
          </div>
        </div>
      </div>

        {/* Normal mode search anchor */}
        <div className="relative z-[80] container mx-auto px-4 pt-4 pb-6 text-center min-h-[88px] pointer-events-none">
          <div
            ref={searchRef}
            className="relative z-[90] will-change-transform pointer-events-none"
            style={{
              transform: `translate3d(0, ${searchOffsetY}px, 0)`,
              transition: prefersReducedMotion
                ? 'none'
                : 'transform 680ms cubic-bezier(0.22, 1, 0.36, 1)',
            }}
          >
            <SearchSection
              searchQuery={searchQuery}
              onSearchQueryChange={setSearchQuery}
              selectedEngine={selectedEngine}
              onEngineChange={setSelectedEngine}
            />
          </div>
        </div>

        {/* Content sections (hidden in focused mode) */}
        <div
          className="relative z-10 will-change-transform"
          data-content-glass={contentGlassReady && !isFocused ? 'on' : 'off'}
          style={{
            transform: isFocused ? 'translate3d(0, 48px, 0)' : 'translate3d(0, 0, 0)',
            opacity: isFocused ? 0.001 : 1,
            pointerEvents: isFocused ? 'none' : 'auto',
            isolation: 'isolate',
            transition: prefersReducedMotion
              ? 'none'
              : isFocused
                ? 'transform 420ms cubic-bezier(0.22, 1, 0.36, 1), opacity 180ms ease-out'
                : 'transform 520ms cubic-bezier(0.22, 1, 0.36, 1), opacity 260ms ease-out 80ms',
          }}
        >
          <div className="container mx-auto px-4 pb-4">
            {/* Links in constrained width */}
            <div className="max-w-4xl 2xl:max-w-none mx-auto">
              <ExternalRequestsPanel />

              {/* Links Section */}
              <LinksSection />
            </div>
          </div>
        </div>

        {/* Mode switch hint */}
        <div className="fixed bottom-4 left-4 z-30 max-w-md">
          <div className="bg-black/20 backdrop-blur-sm rounded-lg px-3 py-2 border border-white/10">
            <div className="text-white/60 text-xs text-center whitespace-pre-line">
              {t('wallpaper.switchHint')}
            </div>
          </div>
        </div>

        {/* Wallpaper info watermark (focused mode only) */}
        {wallpaperInfo && (
          <div className={`fixed bottom-4 left-4 z-40 max-w-md ${
            isFocused
              ? 'opacity-100 transition-all duration-500 ease-out delay-300'
              : 'opacity-0 pointer-events-none transition-all duration-200 ease-in'
          }`}>
            <div className="bg-black/20 backdrop-blur-sm rounded-lg px-4 py-2 border border-white/10">
              <p className="text-white/80 text-sm font-medium mb-1">
                {t(wallpaperInfo.titleKey!)}
              </p>
              <p className="text-white/60 text-xs leading-relaxed">
                {wallpaperInfo.copyright || t(wallpaperInfo.titleKey!)}
              </p>
            </div>
          </div>
        )}
    </div>
  )
}
