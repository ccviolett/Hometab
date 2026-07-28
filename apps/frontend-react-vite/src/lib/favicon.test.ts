import { describe, expect, it } from 'vitest'
import { hostColor, hostFromUrl, letterAvatar } from './favicon'

describe('favicon fallbacks', () => {
  it('normalizes URL hosts', () => {
    expect(hostFromUrl('https://www.example.com/path')).toBe('example.com')
    expect(hostFromUrl('www.example.com')).toBe('example.com')
  })

  it('uses a stable host color', () => {
    expect(hostColor('example.com')).toBe(hostColor('example.com'))
  })

  it('creates an encoded SVG data URI', () => {
    const avatar = letterAvatar('example.com', '<')
    expect(avatar).toMatch(/^data:image\/svg\+xml;utf8,/)
    expect(decodeURIComponent(avatar)).toContain('&lt;')
  })
})
