import { describe, expect, it } from 'vitest'
import { keyValueRowsFromJSON, keyValueRowsToJSON, parserRowsFromJSON, parserRowsToJSON } from './request-form-utils'

describe('external request form conversions', () => {
  it('round trips key value objects and omits empty rows', () => {
    const rows = keyValueRowsFromJSON('{"Authorization":"Bearer token","page":"1"}')
    rows.push({ id: 'empty', key: '', value: '' })
    expect(JSON.parse(keyValueRowsToJSON(rows))).toEqual({ Authorization: 'Bearer token', page: '1' })
  })

  it('rejects duplicate keys', () => {
    expect(() => keyValueRowsToJSON([
      { id: '1', key: 'x', value: '1' },
      { id: '2', key: 'x', value: '2' },
    ])).toThrow('Duplicate key')
  })

  it('validates parser paths', () => {
    const rows = parserRowsFromJSON('[{"label":"Count","path":"$.data.total"}]')
    expect(JSON.parse(parserRowsToJSON(rows))).toEqual([{ label: 'Count', path: '$.data.total' }])
    expect(() => parserRowsToJSON([{ id: '1', label: 'Count', path: 'data.total' }])).toThrow('Path must start')
  })
})
