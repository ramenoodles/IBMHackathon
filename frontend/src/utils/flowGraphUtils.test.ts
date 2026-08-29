import { describe, expect, it } from 'vitest'
import {
  bfsNodeIds,
  createSymbolFlowState,
  enrichmentHorizon,
} from '@/utils/flowGraphUtils'
import type { FlowEdge, FlowNode } from '@/types/flowGraph'

const nodes: FlowNode[] = [
  { id: 'a', label: 'a', summary: '', kind: 'entry', confidence: 'verified', expandable: false, childCount: 0, collapsed: false },
  { id: 'b', label: 'b', summary: '', kind: 'call', confidence: 'verified', expandable: false, childCount: 0, collapsed: false },
  { id: 'c', label: 'c', summary: '', kind: 'branch', confidence: 'verified', expandable: false, childCount: 0, collapsed: false },
  { id: 'd', label: 'd', summary: '', kind: 'return', confidence: 'verified', expandable: false, childCount: 0, collapsed: false },
]

const edges: FlowEdge[] = [
  { from: 'a', to: 'b', label: 'then' },
  { from: 'b', to: 'c', label: 'then' },
  { from: 'c', to: 'd', label: 'true' },
]

describe('bfsNodeIds', () => {
  it('returns up to maxCount nodes in BFS order', () => {
    expect(bfsNodeIds('a', nodes, edges, 3)).toEqual(['a', 'b', 'c'])
  })

  it('returns empty for missing root', () => {
    expect(bfsNodeIds('', nodes, edges, 3)).toEqual([])
  })
})

describe('enrichmentHorizon', () => {
  it('collects nodes within depth hops', () => {
    const horizon = enrichmentHorizon('a', edges, 2)
    expect(horizon).toEqual(['b', 'c'])
  })
})

describe('createSymbolFlowState', () => {
  it('pre-reveals first N nodes', () => {
    const state = createSymbolFlowState(
      { rootId: 'a', nodes, edges, depth: 1, symbol: 'fn' },
      3,
    )
    expect([...state.revealedIds]).toEqual(['a', 'b', 'c'])
  })
})
