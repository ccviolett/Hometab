export interface KeyValueRow {
  id: string
  key: string
  value: string
}

export interface ParserRow {
  id: string
  label: string
  path: string
}

function nextId() {
  return Math.random().toString(36).slice(2)
}

export function keyValueRowsFromJSON(raw: string): KeyValueRow[] {
  if (!raw.trim()) return [{ id: nextId(), key: '', value: '' }]
  try {
    const value = JSON.parse(raw) as unknown
    if (!value || Array.isArray(value) || typeof value !== 'object') return [{ id: nextId(), key: '', value: '' }]
    const rows = Object.entries(value).map(([key, item]) => ({ id: nextId(), key, value: String(item) }))
    return rows.length > 0 ? rows : [{ id: nextId(), key: '', value: '' }]
  } catch {
    return [{ id: nextId(), key: '', value: '' }]
  }
}

export function keyValueRowsToJSON(rows: KeyValueRow[]): string {
  const result: Record<string, string> = {}
  for (const row of rows) {
    const key = row.key.trim()
    if (!key && !row.value.trim()) continue
    if (!key) throw new Error('Key is required')
    if (Object.prototype.hasOwnProperty.call(result, key)) throw new Error(`Duplicate key: ${key}`)
    result[key] = row.value
  }
  return Object.keys(result).length > 0 ? JSON.stringify(result) : ''
}

export function parserRowsFromJSON(raw: string): ParserRow[] {
  if (!raw.trim()) return [{ id: nextId(), label: '', path: '' }]
  try {
    const items = JSON.parse(raw) as Array<{ label?: unknown; path?: unknown }>
    if (!Array.isArray(items)) return [{ id: nextId(), label: '', path: '' }]
    const rows = items.map((item) => ({ id: nextId(), label: String(item.label ?? ''), path: String(item.path ?? '') }))
    return rows.length > 0 ? rows : [{ id: nextId(), label: '', path: '' }]
  } catch {
    return [{ id: nextId(), label: '', path: '' }]
  }
}

export function parserRowsToJSON(rows: ParserRow[]): string {
  const result: Array<{ label: string; path: string }> = []
  for (const row of rows) {
    const label = row.label.trim()
    const path = row.path.trim()
    if (!label && !path) continue
    if (!label || !path) throw new Error('Label and path are required')
    if (path !== '$' && !path.startsWith('$.')) throw new Error('Path must start with $.')
    result.push({ label, path })
  }
  return result.length > 0 ? JSON.stringify(result) : ''
}
