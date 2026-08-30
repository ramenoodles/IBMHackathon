import type { EnrichNodeInput, GraphRootPayload } from '@/types/flowGraph'
import type { FlowGraph } from '@/types/flowGraph'
import type { SymbolFlowState } from '@/utils/flowGraphUtils'
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

export async function enrichSymbolNodes(
  state: SymbolFlowState,
  payload: GraphRootPayload,
  nodeIds: string[],
  inFlight?: Set<string>,
  priorityIds?: Set<string>,
): Promise<{ enrichError: string | null }> {
  let pending = nodeIds.filter((id) => !state.enrichedIds.has(id) && !inFlight?.has(id))
  if (pending.length === 0) return { enrichError: null }

  if (priorityIds?.size) {
    pending = sortByPriority(pending, priorityIds)
  }

  for (const id of pending) inFlight?.add(id)

  const byId = new Map(state.allNodes.map((n) => [n.id, n]))
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
          const node = state.allNodes.find((n) => n.id === patch.id)
          if (!node) continue
          if (patch.title) node.title = patch.title
          if (patch.summary) node.summary = patch.summary
          if (patch.labelSource) {
            node.labelSource = patch.labelSource
            if (patch.labelSource === 'ai' || patch.labelSource === 'heuristic') {
              node.confidence = 'inferred'
            }
          }
          state.enrichedIds.add(patch.id)
        }
      } catch (err) {
        enrichError = err instanceof Error ? err.message : 'Failed to label flow steps'
        console.warn('[enrich]', enrichError)
      }
    }
  } finally {
    for (const id of pending) inFlight?.delete(id)
  }
  return { enrichError }
}
