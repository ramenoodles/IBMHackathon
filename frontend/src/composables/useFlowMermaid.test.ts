import { describe, expect, it, vi } from 'vitest'
import { compileToMermaid } from '@/composables/useFlowMermaid'
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
