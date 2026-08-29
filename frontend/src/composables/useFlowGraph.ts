import { ref } from 'vue'
import type {
  FlowEdge,
  FlowGraph,
  FlowNode,
  GraphExpandPayload,
  GraphRootPayload,
} from '@/types/flowGraph'

/**
 * Manages flow graph state with root load and lazy branch expansion.
 */
export function useFlowGraph() {
  const nodes = ref<FlowNode[]>([])
  const edges = ref<FlowEdge[]>([])
  const rootId = ref('')
  const symbol = ref('')
  const loading = ref(false)
  const expanding = ref(false)
  const error = ref<string | null>(null)
  const isMock = ref(false)
  const parentPath = ref<string[]>([])
  const currentFilePath = ref('')
  const currentWorkspace = ref('')

  /**
   * Merge a graph fragment into the current graph state.
   * @param fragment - Partial graph from expand endpoint.
   */
  function mergeFragment(fragment: FlowGraph): void {
    const existingIds = new Set(nodes.value.map((n) => n.id))
    for (const node of fragment.nodes) {
      if (!existingIds.has(node.id)) {
        nodes.value.push(node)
        existingIds.add(node.id)
      }
    }
    const edgeKeys = new Set(edges.value.map((e) => `${e.from}->${e.to}`))
    for (const edge of fragment.edges) {
      const key = `${edge.from}->${edge.to}`
      if (!edgeKeys.has(key)) {
        edges.value.push(edge)
        edgeKeys.add(key)
      }
    }
    const expanded = nodes.value.find((n) => n.id === fragment.rootId)
    if (expanded) {
      expanded.collapsed = false
      expanded.expandable = false
    }
  }

  /**
   * Load the root execution-flow graph for a symbol.
   * @param payload - Root graph request payload.
   */
  async function loadRoot(payload: GraphRootPayload): Promise<void> {
    loading.value = true
    error.value = null
    parentPath.value = []
    currentFilePath.value = payload.filePath
    currentWorkspace.value = payload.workspacePath
    symbol.value = payload.symbol

    try {
      const res = await fetch('/api/graph/root', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw new Error(`Graph load failed (${res.status})`)
      const data = (await res.json()) as FlowGraph
      nodes.value = data.nodes
      edges.value = data.edges
      rootId.value = data.rootId
      isMock.value = Boolean(data.mock)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load graph'
      nodes.value = []
      edges.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * Expand a collapsed branch node lazily.
   * @param nodeId - Node to expand.
   * @param payload - Expand request context.
   * @param limit - Max child nodes to fetch.
   */
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
      if (fragment.mock) isMock.value = true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Expand failed'
    } finally {
      expanding.value = false
    }
  }

  /** Reset graph state. */
  function reset(): void {
    nodes.value = []
    edges.value = []
    rootId.value = ''
    symbol.value = ''
    error.value = null
    isMock.value = false
    parentPath.value = []
  }

  return {
    nodes,
    edges,
    rootId,
    symbol,
    loading,
    expanding,
    error,
    isMock,
    currentFilePath,
    currentWorkspace,
    loadRoot,
    expandNode,
    reset,
  }
}
