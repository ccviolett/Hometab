// Offline letter avatar used when the backend cannot resolve a site icon.

const PALETTE = [
  'hsl(221 83% 53%)', // blue-600
  'hsl(262 83% 58%)', // violet
  'hsl(199 89% 48%)', // sky
  'hsl(160 84% 39%)', // emerald
  'hsl(32 95% 44%)', // amber
  'hsl(0 72% 51%)', // red
  'hsl(291 64% 55%)', // fuchsia
  'hsl(173 80% 36%)', // teal
]

function hashString(s: string): number {
  let h = 0
  for (let i = 0; i < s.length; i++) {
    h = (h << 5) - h + s.charCodeAt(i)
    h |= 0
  }
  return Math.abs(h)
}

export function hostFromUrl(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '')
  } catch {
    return url.replace(/^www\./, '')
  }
}

export function hostColor(host: string): string {
  return PALETTE[hashString(host) % PALETTE.length]
}

function escapeXml(s: string): string {
  return s.replace(/[<>&'"]/g, (c) => {
    switch (c) {
      case '<': return '&lt;'
      case '>': return '&gt;'
      case '&': return '&amp;'
      case "'": return '&apos;'
      default: return '&quot;'
    }
  })
}

/** 生成一张内联 SVG 字母头像（正方形圆角，居中首字母），返回 data URI。 */
export function letterAvatar(host: string, label?: string): string {
  const color = hostColor(host)
  const cleanHost = host.replace(/^www\./, '')
  const letter = (label && label.trim()[0]) || cleanHost.match(/[a-z0-9]/i)?.[0] || '?'
  const svg =
    `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">` +
    `<rect width="32" height="32" rx="7" fill="${color}"/>` +
    `<text x="16" y="22" font-family="-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif" ` +
    `font-size="17" font-weight="600" fill="#ffffff" text-anchor="middle">${escapeXml(letter.toUpperCase())}</text>` +
    `</svg>`
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`
}
