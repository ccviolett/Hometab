import { useState } from 'react'
import IconService from '@/lib/iconService'
import { hostFromUrl, letterAvatar } from '@/lib/favicon'

function FaviconImage({ url, size, version }: { url: string; size: 'small' | 'normal'; version: number }) {
  const [failed, setFailed] = useState(false)
  const containerClass = size === 'small'
    ? 'w-4 h-4 rounded-sm bg-primary/10 flex items-center justify-center overflow-hidden flex-shrink-0'
    : 'w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center overflow-hidden'
  const iconClass = size === 'small' ? 'w-3 h-3' : 'w-6 h-6'
  const src = failed ? letterAvatar(hostFromUrl(url)) : IconService.getIconUrl(url, version)

  return <div className={containerClass}><img src={src} alt={url} className={iconClass} onError={() => setFailed(true)} /></div>
}

export function FaviconIcon({ url, size = 'small', version = 0 }: { url: string; size?: 'small' | 'normal'; version?: number }) {
  return <FaviconImage key={`${url}:${version}`} url={url} size={size} version={version} />
}
