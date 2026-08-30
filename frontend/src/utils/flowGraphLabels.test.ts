import { describe, expect, it } from 'vitest'
import { hasEnrichedLabel, isCompactNode, labelSourceBadge, labelSourcePill } from '@/utils/flowGraphLabels'
import type { FlowNode } from '@/types/flowGraph'

const baseNode = (overrides: Partial<FlowNode> = {}): FlowNode => ({
  id: 'n1',
  label: 'x',
  summary: 'summary',
  kind: 'call',
  confidence: 'verified',
  expandable: false,
  childCount: 0,
  collapsed: false,
  ...overrides,
})

describe('flowGraphLabels', () => {
  it('returns pills for heuristic and ai sources', () => {
    expect(labelSourcePill('heuristic')).toBe('Auto')
    expect(labelSourcePill('ai')).toBe('AI')
    expect(labelSourcePill('scan')).toBeNull()
    expect(labelSourcePill(undefined)).toBeNull()
  })

  it('returns badges for details panel', () => {
    expect(labelSourceBadge('heuristic')).toBe('Auto label')
    expect(labelSourceBadge('ai')).toBe('AI label')
    expect(labelSourceBadge(undefined)).toBeNull()
  })

  it('detects enriched labels', () => {
    expect(hasEnrichedLabel(baseNode())).toBe(false)
    expect(hasEnrichedLabel(baseNode({ labelSource: 'heuristic' }))).toBe(true)
    expect(hasEnrichedLabel(baseNode({ labelSource: 'ai' }))).toBe(true)
  })

  it('detects compact previewable nodes', () => {
    expect(isCompactNode(baseNode())).toBe(false)
    expect(
      isCompactNode(
        baseNode({
          collapsed: true,
          expandable: true,
          calleeFile: 'foo.go',
          calleeSymbol: 'close',
        }),
      ),
    ).toBe(true)
    expect(
      isCompactNode(
        baseNode({
          collapsed: true,
          expandable: true,
          calleeFile: 'foo.go',
        }),
      ),
    ).toBe(false)
  })
})
