import { describe, expect, it, vi } from 'vitest'
import { compileToMermaid, graphLabelKey, graphStructureKey, nodeMermaidClasses, nodeScanTitle } from '@/composables/useFlowMermaid'
import type { FlowNode } from '@/types/flowGraph'

const baseNode = (id: string, overrides: Partial<FlowNode> = {}): FlowNode => ({
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

describe('compileToMermaid', () => {
  it('does not embed selected class in DSL', () => {
    const nodes = [baseNode('a', { kind: 'entry' }), baseNode('b')]
    const code = compileToMermaid(nodes, [{ from: 'a', to: 'b' }])
    expect(code).not.toContain('selected')
    expect(code).toContain('classDef verified')
  })

  it('does not append child count suffix on collapsed nodes', () => {
    const nodes = [
      baseNode('a', {
        title: 'Initialize Topic Configuration',
        collapsed: true,
        expandable: true,
        childCount: 1,
      }),
    ]
    const code = compileToMermaid(nodes, [])
    expect(code).toContain('Initialize Topic Configuration')
    expect(code).not.toMatch(/\(\+\d+\)/)
    expect(code).toContain('collapsed')
  })

  it('maps labelSource to mermaid classes', () => {
    expect(nodeMermaidClasses(baseNode('a'))).toEqual(['verified'])
    expect(nodeMermaidClasses(baseNode('a', { labelSource: 'ai' }))).toEqual(['inferred'])
    expect(nodeMermaidClasses(baseNode('a', { labelSource: 'heuristic' }))).toEqual(['heuristic'])
    expect(nodeMermaidClasses(baseNode('a', { labelSource: 'ai', collapsed: true }))).toEqual(['collapsed'])
  })

  it('uses scan labels when labelMode is scan', () => {
    const nodes = [
      baseNode('a', {
        label: 'L12 m.mu.Lock()',
        title: 'Acquire lock',
        summary: 'Locks the mutex',
      }),
    ]
    expect(nodeScanTitle(nodes[0]!)).toBe('m.mu.Lock()')
    const code = compileToMermaid(nodes, [], 'scan')
    expect(code).toContain('m.mu.Lock()')
    expect(code).not.toContain('Acquire lock')
  })

  it('separates structure and label keys', () => {
    const nodes = [baseNode('a', { title: 'One' }), baseNode('b', { kind: 'branch' })]
    const edges = [{ from: 'a', to: 'b' }]
    const structure = graphStructureKey(nodes, edges)
    expect(graphStructureKey(nodes, edges)).toBe(structure)
    expect(graphLabelKey(nodes)).toContain('One')

    const relabeled = [baseNode('a', { title: 'Two' }), baseNode('b', { kind: 'branch' })]
    expect(graphStructureKey(relabeled, edges)).toBe(structure)
    expect(graphLabelKey(relabeled)).not.toBe(graphLabelKey(nodes))
  })
})

describe('useNodeDetail cache', () => {
  it('returns cached detail without a second fetch', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          id: 'n1',
          title: 'Step',
          summary: 's',
          explanation: 'cached',
          confidence: 'verified',
        }),
      }),
    )

    const { useNodeDetail } = await import('@/composables/useNodeDetail')
    const { loadDetail, detail } = useNodeDetail()
    const params = {
      workspaceId: 'ws',
      nodeId: 'n1',
      symbol: 'fn',
      file: 'a.py',
    }

    await loadDetail(params)
    expect(detail.value?.explanation).toBe('cached')
    await loadDetail(params)
    expect(fetch).toHaveBeenCalledTimes(1)

    vi.unstubAllGlobals()
  })
})
