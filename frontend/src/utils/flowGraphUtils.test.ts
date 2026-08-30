import { describe, expect, it } from 'vitest'
import {
  INITIAL_VISIBLE_COUNT,
  SILENT_BUFFER_STEPS,
  bfsNodeIds,
  createSymbolFlowState,
  enrichmentHorizon,
  pruneGraphToRoot,
  silentPrefetchTargets,
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

describe('silentPrefetchTargets', () => {
  it('returns next buffer steps without including visible nodes', () => {
    expect(silentPrefetchTargets('a', nodes, edges, INITIAL_VISIBLE_COUNT, SILENT_BUFFER_STEPS)).toEqual([
      'b',
      'c',
    ])
  })

  it('returns empty when buffer is zero', () => {
    expect(silentPrefetchTargets('a', nodes, edges, 1, 0)).toEqual([])
  })
})

describe('enrichmentHorizon', () => {
  it('collects nodes within depth hops', () => {
    const horizon = enrichmentHorizon('a', edges, 2)
    expect(horizon).toEqual(['b', 'c'])
  })
})

describe('createSymbolFlowState', () => {
  it('reveals only entry by default', () => {
    const state = createSymbolFlowState(
      { rootId: 'a', nodes, edges, depth: 1, symbol: 'fn' },
    )
    expect([...state.revealedIds]).toEqual(['a'])
    expect(state.fullyExpanded).toBe(false)
  })

  it('pre-reveals first N nodes when requested', () => {
    const state = createSymbolFlowState(
      { rootId: 'a', nodes, edges, depth: 1, symbol: 'fn' },
      3,
    )
    expect([...state.revealedIds]).toEqual(['a', 'b', 'c'])
  })
})

describe('pruneGraphToRoot', () => {
  it('drops disconnected nodes that are not reachable from the current root', () => {
    const staleNodes: FlowNode[] = [
      ...nodes,
      { id: 'x', label: 'x', summary: '', kind: 'call', confidence: 'verified', expandable: false, childCount: 0, collapsed: false },
    ]
    const staleEdges: FlowEdge[] = [
      ...edges,
      { from: 'x', to: 'a', label: 'then' },
    ]

    const pruned = pruneGraphToRoot('a', staleNodes, staleEdges)

    expect(pruned.nodes.map((n) => n.id)).toEqual(['a', 'b', 'c', 'd'])
    expect(pruned.edges.map((e) => `${e.from}->${e.to}`)).toEqual(['a->b', 'b->c', 'c->d'])
  })
})
