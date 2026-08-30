import { describe, expect, it } from 'vitest'
import { buildFlowStepTree, expandBranchIdsForTarget, flattenStepTree } from '@/utils/flowStepTree'
import type { FlowEdge, FlowNode } from '@/types/flowGraph'

const base = (id: string, overrides: Partial<FlowNode> = {}): FlowNode => ({
  id,
  label: id,
  summary: '',
  kind: 'call',
  confidence: 'verified',
  expandable: false,
  childCount: 0,
  collapsed: false,
  ...overrides,
})

describe('buildFlowStepTree', () => {
  it('keeps linear chains at the same depth', () => {
    const nodes = [
      base('a', { kind: 'entry' }),
      base('b'),
      base('c'),
    ]
    const edges: FlowEdge[] = [
      { from: 'a', to: 'b', label: 'then' },
      { from: 'b', to: 'c', label: 'then' },
    ]
    const rows = flattenStepTree(buildFlowStepTree('a', nodes, edges))
    expect(rows.map((r) => r.node.id)).toEqual(['a', 'b', 'c'])
    expect(rows.every((r) => r.depth === 0)).toBe(true)
  })

  it('indents branch children', () => {
    const nodes = [
      base('a', { kind: 'entry' }),
      base('b', { kind: 'branch' }),
      base('c'),
      base('d'),
    ]
    const edges: FlowEdge[] = [
      { from: 'a', to: 'b', label: 'then' },
      { from: 'b', to: 'c', label: 'true' },
      { from: 'b', to: 'd', label: 'false' },
    ]
    const tree = buildFlowStepTree('a', nodes, edges)
    const branch = tree.find((t) => t.node.id === 'b')
    expect(branch?.branch).toBe(true)
    expect(branch?.children.map((c) => c.node.id)).toEqual(['c', 'd'])
    expect(branch?.children.every((c) => c.depth === 1)).toBe(true)
    expect(branch?.children[0]?.edgeLabel).toBe('true')
    expect(branch?.children[1]?.edgeLabel).toBe('false')
  })

  it('visits reconverged nodes only once', () => {
    const nodes = [
      base('a', { kind: 'entry' }),
      base('b', { kind: 'branch' }),
      base('c'),
      base('d'),
      base('merge'),
    ]
    const edges: FlowEdge[] = [
      { from: 'a', to: 'b' },
      { from: 'b', to: 'c', label: 'true' },
      { from: 'b', to: 'd', label: 'false' },
      { from: 'c', to: 'merge' },
      { from: 'd', to: 'merge' },
    ]
    const rows = flattenStepTree(buildFlowStepTree('a', nodes, edges))
    expect(rows.filter((r) => r.node.id === 'merge')).toHaveLength(1)
  })
})

describe('expandBranchIdsForTarget', () => {
  it('returns enclosing branch ids', () => {
    const nodes = [
      base('a', { kind: 'entry' }),
      base('b', { kind: 'branch' }),
      base('c'),
      base('d'),
    ]
    const edges: FlowEdge[] = [
      { from: 'a', to: 'b' },
      { from: 'b', to: 'c', label: 'true' },
      { from: 'b', to: 'd', label: 'false' },
    ]
    const tree = buildFlowStepTree('a', nodes, edges)
    expect(expandBranchIdsForTarget(tree, 'd')).toEqual(['b'])
    expect(expandBranchIdsForTarget(tree, 'a')).toEqual([])
  })
})
