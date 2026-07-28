import { describe, expect, it } from 'vitest'
import { moveId } from './reorder'

describe('moveId', () => {
  it('moves an item without mutating the source', () => {
    const ids = ['a', 'b', 'c']
    expect(moveId(ids, 'c', 'a')).toEqual(['c', 'a', 'b'])
    expect(ids).toEqual(['a', 'b', 'c'])
  })

  it('returns null for no-op and invalid targets', () => {
    expect(moveId(['a'], 'a', 'a')).toBeNull()
    expect(moveId(['a'], 'x', 'a')).toBeNull()
  })
})
