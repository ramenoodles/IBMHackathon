import type { EnrichNodeInput, EnrichPatch, GraphRootPayload } from '@/types/flowGraph'
import type { FlowGraph } from '@/types/flowGraph'
import type { SymbolFlowState } from '@/utils/flowGraphUtils'

export async function fetchGraphRoot(payload: GraphRootPayload): Promise<FlowGraph> {
  const res = await fetch(`/api/workspaces/${encodeURIComponent(payload.workspaceId)}/graphs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  if (!res.ok) throw new Error(`Graph load failed (${res.status})`)
  return (await res.json()) as FlowGraph
}

export async function enrichSymbolNodes(
  state: SymbolFlowState,
  payload: GraphRootPayload,
  nodeIds: string[],
  inFlight?: Set<string>,
): Promise<void> {
  const pending = nodeIds.filter((id) => !state.enrichedIds.has(id) && !inFlight?.has(id))
  if (pending.length === 0) return

  for (const id of pending) inFlight?.add(id)

  const byId = new Map(state.allNodes.map((n) => [n.id, n]))
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
        }
      })
      const res = await fetch(`/api/workspaces/${encodeURIComponent(payload.workspaceId)}/explain`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          workspaceId: payload.workspaceId,
          filePath: payload.filePath,
          symbol: payload.symbol,
          nodes: enrichInputs,
          userContext: payload.userContext,
        }),
      })
      if (!res.ok) continue
      const data = (await res.json()) as { patches: EnrichPatch[] }
      for (const patch of data.patches ?? []) {
        const node = state.allNodes.find((n) => n.id === patch.id)
        if (!node) continue
        if (patch.title) node.title = patch.title
        if (patch.summary) node.summary = patch.summary
        state.enrichedIds.add(patch.id)
      }
    }
  } finally {
    for (const id of pending) inFlight?.delete(id)
  }
}
