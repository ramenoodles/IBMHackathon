import type { EnrichNodeInput, FlowNode, FlowGraph, GraphRootPayload } from '@/types/flowGraph'
import { api } from '@/api'

export async function fetchGraphRoot(payload: GraphRootPayload): Promise<FlowGraph> {
  return api.graph(payload)
}

function sortByPriority(ids: string[], priority: Set<string>): string[] {
  return [...ids].sort((a, b) => {
    const av = priority.has(a) ? 0 : 1
    const bv = priority.has(b) ? 0 : 1
    return av - bv
  })
}

export interface EnrichPatchResult {
  patches: Array<{
    id: string
    title?: string
    summary?: string
    labelSource?: string
  }>
  enrichError: string | null
}

export async function enrichSymbolNodes(
  nodes: FlowNode[],
  payload: GraphRootPayload,
  nodeIds: string[],
  inFlight?: Set<string>,
  alreadyEnriched?: Set<string>,
  priorityIds?: Set<string>,
): Promise<EnrichPatchResult> {
  let pending = nodeIds.filter((id) => !alreadyEnriched?.has(id) && !inFlight?.has(id))
  if (pending.length === 0) return { patches: [], enrichError: null }

  if (priorityIds?.size) {
    pending = sortByPriority(pending, priorityIds)
  }

  for (const id of pending) inFlight?.add(id)

  const byId = new Map(nodes.map((n) => [n.id, n]))
  const allPatches: EnrichPatchResult['patches'] = []
  let enrichError: string | null = null
  try {
    for (let i = 0; i < pending.length; i += 8) {
      const batch = pending.slice(i, i + 8)
      const enrichInputs: EnrichNodeInput[] = batch.map((id) => {
        const n = byId.get(id)!
        return {
          id: n.id,
          line: n.line ?? 0,
          code: n.code ?? n.summary,
          kind: n.kind,
          label: n.label,
        }
      })
      try {
        const data = await api.enrich(payload.workspaceId, {
          workspaceId: payload.workspaceId,
          filePath: payload.filePath,
          symbol: payload.symbol,
          nodes: enrichInputs,
          userContext: payload.userContext,
        })
        for (const patch of data.patches ?? []) {
          allPatches.push({ id: patch.id, title: patch.title, summary: patch.summary, labelSource: patch.labelSource })
        }
      } catch (err) {
        enrichError = err instanceof Error ? err.message : 'Failed to label flow steps'
        console.warn('[enrich]', enrichError)
      }
    }
  } finally {
    for (const id of pending) inFlight?.delete(id)
  }
  return { patches: allPatches, enrichError }
}
