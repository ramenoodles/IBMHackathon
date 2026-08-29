import { ref } from 'vue'
import type {
  EnrichNodeInput,
  EnrichPatch,
  FlowEdge,
  FlowGraph,
  FlowNode,
  GraphExpandPayload,
  GraphRootPayload,
} from '@/types/flowGraph'

/**
 * Manages scan-first flow graph with progressive step reveal and async LLM labels.
 */
export function useFlowGraph() {
  const allNodes = ref<FlowNode[]>([])
  const allEdges = ref<FlowEdge[]>([])
  const nodes = ref<FlowNode[]>([])
  const edges = ref<FlowEdge[]>([])
  const revealedIds = ref<Set<string>>(new Set())
  const enrichedIds = ref<Set<string>>(new Set())
  const rootId = ref('')
  const symbol = ref('')
  const loading = ref(false)
  const enriching = ref(false)
  const expanding = ref(false)
  const error = ref<string | null>(null)
  const isMock = ref(false)
  const parentPath = ref<string[]>([])
  const currentFilePath = ref('')
  const currentWorkspace = ref('')
  const lastPayload = ref<GraphRootPayload | null>(null)

  function syncVisible(): void {
    const revealed = revealedIds.value
    nodes.value = allNodes.value.filter((n) => revealed.has(n.id))
    edges.value = allEdges.value.filter((e) => revealed.has(e.from) && revealed.has(e.to))
  }

  function applyPatches(patches: EnrichPatch[]): void {
    for (const patch of patches) {
      const node = allNodes.value.find((n) => n.id === patch.id)
      if (!node) continue
      if (patch.title) node.title = patch.title
      if (patch.summary) node.summary = patch.summary
      enrichedIds.value.add(patch.id)
    }
    syncVisible()
  }

  async function enrichNodes(nodeIds: string[]): Promise<void> {
    const payload = lastPayload.value
    if (!payload || nodeIds.length === 0) return

    const pending = nodeIds.filter((id) => !enrichedIds.value.has(id))
    if (pending.length === 0) return

    enriching.value = true
    try {
      const byId = new Map(allNodes.value.map((n) => [n.id, n]))
      const enrichNodes: EnrichNodeInput[] = pending.slice(0, 8).map((id) => {
        const n = byId.get(id)!
        return {
          id: n.id,
          line: n.line ?? 0,
          code: n.code ?? n.summary,
          kind: n.kind,
        }
      })
      const res = await fetch('/api/graph/enrich', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          workspacePath: payload.workspacePath,
          filePath: payload.filePath,
          symbol: payload.symbol,
          nodes: enrichNodes,
          userContext: payload.userContext,
        }),
      })
      if (!res.ok) return
      const data = (await res.json()) as { patches: EnrichPatch[] }
      applyPatches(data.patches ?? [])
    } finally {
      enriching.value = false
    }
  }

  function hasHiddenChildren(nodeId: string): boolean {
    const revealed = revealedIds.value
    return allEdges.value.some((e) => e.from === nodeId && !revealed.has(e.to))
  }

  function revealFromNode(nodeId: string): string[] {
    const revealed = new Set(revealedIds.value)
    const newly: string[] = []
    for (const edge of allEdges.value) {
      if (edge.from === nodeId && !revealed.has(edge.to)) {
        revealed.add(edge.to)
        newly.push(edge.to)
      }
    }
    if (newly.length === 0) return []
    revealedIds.value = revealed
    syncVisible()
    void enrichNodes(newly)
    return newly
  }

  function mergeFragment(fragment: FlowGraph): void {
    const existingIds = new Set(allNodes.value.map((n) => n.id))
    for (const node of fragment.nodes) {
      if (!existingIds.has(node.id)) {
        allNodes.value.push(node)
        existingIds.add(node.id)
      }
    }
    const edgeKeys = new Set(allEdges.value.map((e) => `${e.from}->${e.to}`))
    for (const edge of fragment.edges) {
      const key = `${edge.from}->${edge.to}`
      if (!edgeKeys.has(key)) {
        allEdges.value.push(edge)
        edgeKeys.add(key)
      }
    }
    const expanded = allNodes.value.find((n) => n.id === fragment.rootId)
    if (expanded) {
      expanded.collapsed = false
      expanded.expandable = false
    }
    syncVisible()
  }

  async function loadRoot(payload: GraphRootPayload): Promise<void> {
    loading.value = true
    error.value = null
    parentPath.value = []
    currentFilePath.value = payload.filePath
    currentWorkspace.value = payload.workspacePath
    symbol.value = payload.symbol
    lastPayload.value = payload
    enrichedIds.value = new Set()
    revealedIds.value = new Set()

    try {
      const res = await fetch('/api/graph/root', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw new Error(`Graph load failed (${res.status})`)
      const data = (await res.json()) as FlowGraph
      allNodes.value = data.nodes
      allEdges.value = data.edges
      rootId.value = data.rootId
      isMock.value = Boolean(data.mock) && data.nodes.length === 0

      if (data.rootId) {
        revealedIds.value = new Set([data.rootId])
      }
      syncVisible()
      loading.value = false
      if (data.rootId) {
        void enrichNodes([data.rootId])
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load graph'
      allNodes.value = []
      allEdges.value = []
      nodes.value = []
      edges.value = []
      loading.value = false
    }
  }

  async function expandNode(
    nodeId: string,
    payload: Omit<GraphExpandPayload, 'nodeId' | 'parentPath' | 'expandLimit'>,
    limit = 3,
  ): Promise<void> {
    expanding.value = true
    error.value = null
    const path = [...parentPath.value, nodeId]

    try {
      const res = await fetch('/api/graph/expand', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...payload,
          nodeId,
          parentPath: path,
          expandLimit: limit,
        }),
      })
      if (!res.ok) throw new Error(`Expand failed (${res.status})`)
      const fragment = (await res.json()) as FlowGraph
      mergeFragment(fragment)
      parentPath.value = path
      const newIds = fragment.nodes.map((n) => n.id).filter((id) => !revealedIds.value.has(id))
      for (const id of newIds) {
        revealedIds.value.add(id)
      }
      revealedIds.value = new Set(revealedIds.value)
      syncVisible()
      if (newIds.length) {
        void enrichNodes(newIds)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Expand failed'
    } finally {
      expanding.value = false
    }
  }

  function reset(): void {
    allNodes.value = []
    allEdges.value = []
    nodes.value = []
    edges.value = []
    revealedIds.value = new Set()
    enrichedIds.value = new Set()
    rootId.value = ''
    symbol.value = ''
    error.value = null
    isMock.value = false
    parentPath.value = []
    lastPayload.value = null
  }

  return {
    nodes,
    edges,
    rootId,
    symbol,
    loading,
    enriching,
    expanding,
    error,
    isMock,
    currentFilePath,
    currentWorkspace,
    loadRoot,
    expandNode,
    revealFromNode,
    hasHiddenChildren,
    reset,
  }
}
