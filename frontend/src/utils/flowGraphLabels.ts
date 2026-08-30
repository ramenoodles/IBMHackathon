import type { FlowNode, LabelSource } from '@/types/flowGraph'

/** Short pill text for a node's label provenance. */
export function labelSourcePill(source?: LabelSource): string | null {
  switch (source) {
    case 'heuristic':
      return 'Auto'
    case 'ai':
      return 'AI'
    default:
      return null
  }
}

/** Human-readable badge label for the details panel. */
export function labelSourceBadge(source?: LabelSource): string | null {
  switch (source) {
    case 'heuristic':
      return 'Auto label'
    case 'ai':
      return 'Brief'
    default:
      return null
  }
}

/** Whether the node has an enriched display label (not raw scan text). */
export function hasEnrichedLabel(node: FlowNode): boolean {
  return node.labelSource === 'ai' || node.labelSource === 'heuristic'
}

/** Callee folded into one compact node — previewable without inline expand. */
export function isCompactNode(node: FlowNode): boolean {
  return node.collapsed && node.expandable && Boolean(node.calleeFile && node.calleeSymbol)
}

/** Whether the callee's flow graph can be opened in the scan-only preview modal. */
export function canPreviewCalleeFlow(node: FlowNode): boolean {
  return Boolean(node.calleeFile && node.calleeSymbol)
}
