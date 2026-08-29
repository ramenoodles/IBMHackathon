import type { FlowEdge, FlowGraph, FlowNode } from '@/types/flowGraph'

/** Traceable symbol from GET /api/file/symbols. */
export interface FileSymbol {
  name: string
  line: number
  kind: string
  signature: string
  hint?: string
}

export interface FileSymbolsResponse {
  symbols: FileSymbol[]
  count: number
}

/** Cached graph state for one symbol in a file. */
export interface SymbolFlowState {
  allNodes: FlowNode[]
  allEdges: FlowEdge[]
  rootId: string
  revealedIds: Set<string>
  enrichedIds: Set<string>
  isMock: boolean
  parentPath: string[]
}

export const LARGE_FILE_SYMBOL_THRESHOLD = 8
export const DEFAULT_WARM_SELECTION = 8
export const WARM_CONCURRENCY = 3
export const WARM_REVEAL_COUNT = 3
export const ENRICHMENT_HORIZON_DEPTH = 2

export function cacheKey(filePath: string, symbol: string): string {
  return `${filePath}::${symbol}`
}

export function edgeOrder(label?: string): number {
  if (label === 'true') return 0
  if (label === 'then' || label === 'each' || label === 'start') return 1
  if (label === 'false' || label === 'done') return 2
  return 3
}

/** BFS node ids from start, up to maxCount nodes. */
export function bfsNodeIds(
  rootId: string,
  nodes: FlowNode[],
  edges: FlowEdge[],
  maxCount: number,
): string[] {
  if (!rootId || maxCount <= 0) return []
  const byId = new Map(nodes.map((n) => [n.id, n]))
  const ordered: string[] = []
  const used = new Set<string>()
  const queue = [rootId]

  while (queue.length && ordered.length < maxCount) {
    const id = queue.shift()!
    if (used.has(id)) continue
    if (!byId.has(id)) continue
    ordered.push(id)
    used.add(id)
    const outs = edges.filter((e) => e.from === id)
    outs.sort((a, b) => edgeOrder(a.label) - edgeOrder(b.label))
    for (const e of outs) {
      if (!used.has(e.to)) queue.push(e.to)
    }
  }
  return ordered
}

/** Nodes within depth hops from start along graph edges (for enrichment prefetch). */
export function enrichmentHorizon(
  startId: string,
  edges: FlowEdge[],
  depth: number,
  exclude?: Set<string>,
): string[] {
  if (!startId || depth <= 0) return []
  const targets = new Set<string>()
  let frontier = [startId]
  const visited = new Set<string>([startId])

  for (let d = 0; d < depth; d++) {
    const next: string[] = []
    for (const id of frontier) {
      for (const e of edges) {
        if (e.from !== id) continue
        if (visited.has(e.to)) continue
        visited.add(e.to)
        if (!exclude?.has(e.to)) targets.add(e.to)
        next.push(e.to)
      }
    }
    frontier = next
  }
  return [...targets]
}

export function createSymbolFlowState(graph: FlowGraph, revealCount = 1): SymbolFlowState {
  const revealed = graph.rootId ? bfsNodeIds(graph.rootId, graph.nodes, graph.edges, revealCount) : []
  return {
    allNodes: [...graph.nodes],
    allEdges: [...graph.edges],
    rootId: graph.rootId,
    revealedIds: new Set(revealed),
    enrichedIds: new Set(),
    isMock: Boolean(graph.mock) && graph.nodes.length === 0,
    parentPath: [],
  }
}

export function cloneSymbolFlowState(state: SymbolFlowState): SymbolFlowState {
  return {
    allNodes: state.allNodes.map((n) => ({ ...n })),
    allEdges: [...state.allEdges],
    rootId: state.rootId,
    revealedIds: new Set(state.revealedIds),
    enrichedIds: new Set(state.enrichedIds),
    isMock: state.isMock,
    parentPath: [...state.parentPath],
  }
}
