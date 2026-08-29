import { describe, expect, it } from 'vitest'
import { useFlowGraphCache } from '@/composables/useFlowGraphCache'
import type { SymbolFlowState } from '@/utils/flowGraphUtils'

function mockState(symbol: string, fullyExpanded = false): SymbolFlowState {
  return {
    allNodes: [{ id: 'n1', label: symbol, summary: '', kind: 'entry', confidence: 'verified', expandable: false, childCount: 0, collapsed: false }],
    allEdges: [],
    rootId: 'n1',
    revealedIds: new Set(['n1']),
    enrichedIds: new Set(),
    isMock: false,
    parentPath: [],
    fullyExpanded,
  }
}

describe('useFlowGraphCache', () => {
  it('stores and retrieves symbol state', () => {
    const cache = useFlowGraphCache()
    cache.set('file.py', 'foo', mockState('foo'))
    const got = cache.get('file.py', 'foo')
    expect(got?.allNodes[0]?.label).toBe('foo')
    expect(cache.has('file.py', 'foo')).toBe(true)
  })

  it('lists symbols for a file', () => {
    const cache = useFlowGraphCache()
    cache.set('file.py', 'a', mockState('a'))
    cache.set('file.py', 'b', mockState('b'))
    expect(cache.listSymbolsForFile('file.py').sort()).toEqual(['a', 'b'])
  })

  it('round-trips fullyExpanded flag', () => {
    const cache = useFlowGraphCache()
    cache.set('file.py', 'fn', mockState('fn', true))
    expect(cache.get('file.py', 'fn')?.fullyExpanded).toBe(true)
  })
})
