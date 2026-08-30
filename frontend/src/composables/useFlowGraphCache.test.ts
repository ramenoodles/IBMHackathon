import { describe, expect, it } from 'vitest'
import { useFlowGraph } from '@/composables/useFlowGraph'
import { useFlowGraphCache } from '@/composables/useFlowGraphCache'
import type { FlowEdge, FlowNode } from '@/types/flowGraph'
import { createSymbolFlowState, type SymbolFlowState } from '@/utils/flowGraphUtils'

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

  it('restores the cached reveal set when hydrating a symbol', () => {
    const cache = useFlowGraphCache()
    const nodes: FlowNode[] = [
      { id: 'a', label: 'a', summary: '', kind: 'entry', confidence: 'verified', expandable: false, childCount: 0, collapsed: false },
      { id: 'b', label: 'b', summary: '', kind: 'call', confidence: 'verified', expandable: false, childCount: 0, collapsed: false },
      { id: 'c', label: 'c', summary: '', kind: 'return', confidence: 'verified', expandable: false, childCount: 0, collapsed: false },
    ]
    const edges: FlowEdge[] = [
      { from: 'a', to: 'b', label: 'then' },
      { from: 'b', to: 'c', label: 'then' },
    ]
    const state = createSymbolFlowState({ rootId: 'a', nodes, edges, depth: 2, symbol: 'fn' }, 1)
    state.revealedIds = new Set(['a', 'b'])

    cache.set('file.py', 'fn', state)
    const graph = useFlowGraph(cache)

    const ok = graph.activateSymbol('fn', {
      workspaceId: 'ws',
      filePath: 'file.py',
      symbol: 'fn',
      userContext: {
        primaryLanguage: 'typescript',
        experienceLevel: 'intermediate',
        workspaceId: 'ws',
      },
    })

    expect(ok).toBe(true)
    expect(graph.nodes.value.map((node) => node.id)).toEqual(['a', 'b'])
  })
})
