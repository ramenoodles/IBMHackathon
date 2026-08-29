/** Confidence level for a flow graph node. */
export type FlowConfidence = 'verified' | 'inferred'

/** A single step in an execution-flow graph. */
export interface FlowNode {
  id: string
  label: string
  title?: string
  summary: string
  kind: string
  confidence: FlowConfidence
  file?: string
  line?: number
  code?: string
  calleeSymbol?: string
  calleeFile?: string
  calleeLine?: number
  expandable: boolean
  childCount: number
  collapsed: boolean
}

/** Edge connecting two flow nodes. */
export interface FlowEdge {
  from: string
  to: string
  label?: string
}

/** Flow graph returned by the backend. */
export interface FlowGraph {
  rootId: string
  nodes: FlowNode[]
  edges: FlowEdge[]
  depth: number
  symbol: string
  mock?: boolean
}

/** Node sent for summary enrichment. */
export interface EnrichNodeInput {
  id: string
  line: number
  code: string
  kind: string
}

/** Summary patch from enrich endpoint. */
export interface EnrichPatch {
  id: string
  title?: string
  summary: string
  relatedSymbols?: string[]
}

/** Enrich API response. */
export interface EnrichResult {
  patches: EnrichPatch[]
  mock?: boolean
}

/** Detailed explanation for a selected node. */
export interface NodeDetail {
  id: string
  title: string
  summary: string
  explanation: string
  confidence: FlowConfidence
  file?: string
  line?: number
  relatedSymbols?: string[]
  mock?: boolean
}

/** Payload for loading a root graph. */
export interface GraphRootPayload {
  workspacePath: string
  filePath: string
  symbol: string
  userContext: {
    primaryLanguage: string
    experienceLevel: string
    workspacePath: string
  }
}

/** Payload for expanding a collapsed node. */
export interface GraphExpandPayload {
  workspacePath: string
  filePath: string
  symbol: string
  nodeId: string
  parentPath: string[]
  expandLimit: number
  userContext: GraphRootPayload['userContext']
}
