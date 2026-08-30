<script setup lang="ts">
/**
 * Flow graph canvas: progressive reveal + Mermaid flowchart + execution trace.
 */
import type { FlowEdge, FlowNode, NodeDetail } from '@/types/flowGraph'
import { computed, nextTick, ref, watch } from 'vue'
import MarkdownView from '@/components/workspace/MarkdownView.vue'
import BeaverFlowLoader from '@/components/ui/BeaverFlowLoader.vue'
import LoadingStatus from '@/components/ui/LoadingStatus.vue'
import { graphStructureKey, nodeDisplayTitle, useFlowMermaid } from '@/composables/useFlowMermaid'
import { useFlowPanZoom } from '@/composables/useFlowPanZoom'
import { useHorizontalResize } from '@/composables/usePanelResize'
import { useWorkspaceLayout } from '@/composables/useWorkspaceLayout'
import ResizeHandle from '@/components/ui/ResizeHandle.vue'
import { AI_LOADING_PHRASES } from '@/constants/aiLoadingPhrases'
import { edgeOrder } from '@/utils/flowGraphUtils'
import { canPreviewCalleeFlow, hasEnrichedLabel, isCompactNode, labelSourceBadge, labelSourcePill } from '@/utils/flowGraphLabels'

const {
  tracePanelOpen,
  detailPanelOpen,
  isMobile,
  traceWidth,
  detailWidth,
  TRACE_MIN,
  TRACE_MAX,
  DETAIL_MIN,
  DETAIL_MAX,
  persistWidths,
  toggleTracePanel,
  toggleDetailPanel,
} = useWorkspaceLayout()

const resizeEnabled = computed(() => !isMobile.value)

const { onPointerDown: onTraceResize } = useHorizontalResize({
  width: traceWidth,
  min: TRACE_MIN,
  max: TRACE_MAX,
  side: 'right',
  enabled: resizeEnabled,
  onEnd: persistWidths,
})

const { onPointerDown: onDetailResize } = useHorizontalResize({
  width: detailWidth,
  min: DETAIL_MIN,
  max: DETAIL_MAX,
  side: 'left',
  enabled: resizeEnabled,
  onEnd: persistWidths,
})

const props = defineProps<{
  nodes: FlowNode[]
  edges: FlowEdge[]
  rootId: string
  loading: boolean
  enriching: boolean
  enrichError: string | null
  expanding: boolean
  mappingFullFlow: boolean
  mappingProgress: number
  fullyExpanded: boolean
  error: string | null
  isMock: boolean
  symbol: string
  selectedNodeId: string
  detail: NodeDetail | null
  detailLoading: boolean
  detailStreaming: boolean
  detailError: string | null
  hasHiddenChildren: (nodeId: string) => boolean
}>()

const emit = defineEmits<{
  selectNode: [node: FlowNode]
  expandNode: [node: FlowNode]
  revealNode: [node: FlowNode]
  showFullFlow: []
  requestDetail: [node: FlowNode]
  viewSource: [file?: string, line?: number]
  goToDefinition: [file: string, symbol: string, line?: number]
  previewCompacted: [node: FlowNode]
}>()

const mermaidContainer = ref<HTMLElement | null>(null)
const panViewport = ref<HTMLElement | null>(null)
const panContent = ref<HTMLElement | null>(null)
const deepDiveOpen = ref(false)
const evidenceOpen = ref(false)

const { bind: bindPanZoom, unbind: unbindPanZoom, zoomIn, zoomOut, centerView, getViewport, setViewport } = useFlowPanZoom(
  panViewport,
  panContent,
)

const lastBoundKey = ref('')
const lastSymbol = ref('')

const graphMountKey = computed(
  () => `${props.symbol}:${graphStructureKey(props.nodes, props.edges)}`,
)

function onGraphRender(reason: 'structure' | 'label'): void {
  if (reason === 'label') return
  void nextTick(() => {
    const key = graphMountKey.value
    const isNewSvg = lastBoundKey.value !== key
    if (isNewSvg) {
      const symbolChanged = lastSymbol.value !== props.symbol
      lastSymbol.value = props.symbol
      // Preserve viewport when structure changes within the same symbol (e.g. node expand).
      // Center the view only when the symbol changes (different graph entirely).
      bindPanZoom(!symbolChanged)
      lastBoundKey.value = key
      if (symbolChanged) {
        centerView()
      }
    }
  })
}

function onNodeClick(node: FlowNode): void {
  if (!props.fullyExpanded && node.collapsed && node.expandable) {
    emit('expandNode', node)
    return
  }
  if (props.hasHiddenChildren(node.id)) {
    emit('revealNode', node)
  }
  emit('selectNode', node)
}

const { renderError, setContainer, renderStructural } = useFlowMermaid(
  () => (props.mappingFullFlow ? [] : props.nodes),
  () => (props.mappingFullFlow ? [] : props.edges),
  () => props.selectedNodeId,
  onNodeClick,
  onGraphRender,
)

watch(
  () => props.symbol,
  () => {
    lastBoundKey.value = ''
  },
)

// When mappingFullFlow clears, do one final render with the complete graph
watch(
  () => props.mappingFullFlow,
  (active) => {
    if (!active) void renderStructural()
  },
)

watch(mermaidContainer, (el) => setContainer(el))

watch(
  () => props.selectedNodeId,
  () => {
    deepDiveOpen.value = false
    evidenceOpen.value = false
  },
)

watch(panViewport, (el, prev) => {
  if (!el && prev) unbindPanZoom()
})

const orderedNodes = computed(() => {
  if (!props.nodes.length) return []
  const byId = new Map(props.nodes.map((n) => [n.id, n]))
  const used = new Set<string>()
  const ordered: FlowNode[] = []
  const queue: string[] = []

  const startId = props.rootId || props.nodes[0]?.id
  if (startId) queue.push(startId)

  while (queue.length) {
    const id = queue.shift()!
    if (used.has(id)) continue
    const node = byId.get(id)
    if (!node) continue
    ordered.push(node)
    used.add(id)
    const outs = props.edges.filter((e) => e.from === id)
    outs.sort((a, b) => edgeOrder(a.label) - edgeOrder(b.label))
    for (const e of outs) {
      if (!used.has(e.to)) queue.push(e.to)
    }
  }

  const rest = props.nodes.filter((n) => !used.has(n.id))
  rest.sort((a, b) => (a.line ?? 0) - (b.line ?? 0))
  return [...ordered, ...rest]
})

const selectedNode = computed(() => props.nodes.find((n) => n.id === props.selectedNodeId))
const hasDetailContent = computed(() => Boolean(props.selectedNodeId && selectedNode.value))
const showEnrichedSummary = computed(() => {
  const node = selectedNode.value
  if (!node?.summary?.trim()) return false
  return hasEnrichedLabel(node)
})

const enrichedBadge = computed(() => labelSourceBadge(selectedNode.value?.labelSource))

const verifiedExplanation = computed(
  () => props.detail?.verifiedExplanation?.trim() || '',
)
const inferredExplanation = computed(
  () => props.detail?.inferredExplanation?.trim() || '',
)
const hasEvidence = computed(() => (props.detail?.evidence?.length ?? 0) > 0)

const showDetailLoading = computed(
  () => props.detailLoading && !props.detail?.explanation && !verifiedExplanation.value,
)

function kindLabel(kind: string): string {
  const labels: Record<string, string> = {
    entry: 'entry', call: 'call', branch: 'branch', return: 'return', assign: 'assign', loop: 'loop',
  }
  return labels[kind] ?? kind
}

function openDeepDive(): void {
  if (!selectedNode.value) return
  deepDiveOpen.value = true
  emit('requestDetail', selectedNode.value)
}

</script>

<template>
  <div class="flex h-full min-w-0 flex-1">
    <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <div v-if="isMock" class="shrink-0 bg-amber-900/20 px-4 py-1.5 text-xs text-amber-300">
        No scan data for this symbol
      </div>

      <div v-if="loading" class="flex flex-1 items-center justify-center px-6">
        <BeaverFlowLoader
          mode="indeterminate"
          :active="loading"
          class="w-full max-w-2xl"
        />
      </div>

      <div v-else-if="!symbol" class="flex flex-1 flex-col items-center justify-center gap-2 p-8 text-center">
        <p class="text-lg font-medium text-slate-300">Pick a function to trace</p>
        <p class="max-w-sm text-sm text-slate-500">Select a file, then choose a function from the bar above.</p>
      </div>

      <div v-else-if="!nodes.length" class="flex flex-1 flex-col items-center justify-center gap-1 text-center text-sm text-slate-500">
        <span>No flow data for: <span class="font-mono text-onbober-primary">{{ symbol }}</span></span>
        <span class="text-xs text-slate-600">This is not an error. This function is isolated, it has no calls to or from other functions.</span>
      </div>

      <template v-else>
        <div class="flex shrink-0 items-center justify-between gap-2 border-b border-slate-800 px-4 py-2">
          <div class="flex min-w-0 items-center gap-2">
            <span class="shrink-0 text-xs text-slate-600">
              {{ nodes.length }} step{{ nodes.length !== 1 ? 's' : '' }}
            </span>
            <span
              v-if="enriching"
              class="shrink-0 text-xs font-medium text-amber-400/90"
            >
              Labeling steps…
            </span>
          </div>
          <div class="flex min-w-0 flex-wrap items-center justify-end gap-2 sm:gap-3">
            <div
              class="hidden items-center gap-2 border-r border-slate-800 pr-2 text-[10px] font-medium uppercase tracking-wide text-slate-500 sm:flex"
              aria-label="Flow map legend"
            >
              <span class="flex items-center gap-1" title="Box title still from static code scan">
                <span class="h-2.5 w-2.5 rounded-sm border-2 border-green-400 bg-green-950" />
                Scan
              </span>
              <span class="flex items-center gap-1" title="Onboarding label from Watsonx">
                <span class="h-2.5 w-2.5 rounded-sm border-2 border-dashed border-amber-400 bg-amber-950" />
                AI
              </span>
              <span class="flex items-center gap-1" title="Pattern-based onboarding label">
                <span class="h-2.5 w-2.5 rounded-sm border-2 border-cyan-400 bg-cyan-950" />
                Auto
              </span>
              <span class="flex items-center gap-1" title="Callee folded into one node — preview or click to expand">
                <span class="h-2.5 w-2.5 rounded-sm border-2 border-onbober-primary bg-slate-800" />
                Compact
              </span>
            </div>
            <div class="flex items-center gap-1">
              <button
                type="button"
                class="rounded px-2.5 py-1 text-xs font-medium transition"
                :class="
                  tracePanelOpen
                    ? 'bg-slate-800 text-slate-300'
                    : 'text-slate-500 hover:bg-slate-800 hover:text-slate-300'
                "
                :aria-pressed="tracePanelOpen"
                title="Toggle steps panel"
                @click="toggleTracePanel"
              >
                Steps
              </button>
              <button
                v-if="hasDetailContent"
                type="button"
                class="rounded px-2.5 py-1 text-xs font-medium transition"
                :class="
                  detailPanelOpen
                    ? 'bg-slate-800 text-slate-300'
                    : 'text-slate-500 hover:bg-slate-800 hover:text-slate-300'
                "
                :aria-pressed="detailPanelOpen"
                title="Toggle details panel"
                @click="toggleDetailPanel"
              >
                Details
              </button>
            </div>
            <div class="flex items-center gap-0.5 rounded-md border border-slate-700 p-0.5">
              <button
                type="button"
                class="rounded px-2 py-0.5 text-xs text-slate-400 hover:bg-slate-800 hover:text-white"
                title="Zoom out"
                @click="zoomOut"
              >
                −
              </button>
              <button
                type="button"
                class="rounded px-2 py-0.5 text-xs text-slate-400 hover:bg-slate-800 hover:text-white"
                title="Centre view"
                @click="centerView"
              >
                Center
              </button>
              <button
                type="button"
                class="rounded px-2 py-0.5 text-xs text-slate-400 hover:bg-slate-800 hover:text-white"
                title="Zoom in"
                @click="zoomIn"
              >
                +
              </button>
            </div>
            <button
              v-if="!fullyExpanded && !mappingFullFlow"
              type="button"
              class="rounded-md border border-slate-700 px-3 py-1.5 text-sm font-medium text-slate-300 transition hover:border-onbober-primary/50 hover:text-white"
              @click="emit('showFullFlow')"
            >
              Show full flow
            </button>
          </div>
        </div>

        <p v-if="error" class="shrink-0 px-4 py-2 text-sm text-red-400">{{ error }}</p>
        <p v-if="enrichError && !error" class="shrink-0 px-4 py-1.5 text-xs text-amber-400/90">
          {{ enrichError }} — showing verified labels only.
        </p>

        <div class="flex min-h-0 flex-1">
          <div
            class="steps-panel flex shrink-0 flex-col border-r border-slate-800 bg-slate-900/40"
            :class="tracePanelOpen ? 'steps-panel-open' : 'steps-panel-closed'"
            :style="tracePanelOpen ? { width: `${traceWidth}px` } : { width: '2.75rem' }"
          >
            <template v-if="tracePanelOpen">
              <div class="flex shrink-0 items-center justify-between border-b border-slate-800 px-2 py-1.5">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-400">Steps</span>
                <button
                  type="button"
                  class="rounded p-1 text-slate-500 hover:bg-slate-800 hover:text-white"
                  aria-label="Hide steps panel"
                  title="Hide steps"
                  @click="toggleTracePanel"
                >
                  <svg class="h-3 w-3" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M8 2L4 6l4 4"/></svg>
                </button>
              </div>
              <ol class="min-h-0 flex-1 overflow-y-auto p-2.5">
                <li v-for="(node, i) in orderedNodes" :key="node.id" class="mb-1.5">
                  <button
                    type="button"
                    class="w-full rounded-md border px-2.5 py-2 text-left transition"
                    :class="
                      selectedNodeId === node.id
                        ? 'border-onbober-primary/50 bg-onbober-primary/5'
                        : 'border-transparent hover:border-slate-700 hover:bg-slate-800/50'
                    "
                    @click="onNodeClick(node)"
                  >
                    <div class="flex items-center gap-2">
                      <span
                        class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-bold"
                        :class="
                          selectedNodeId === node.id
                            ? 'bg-onbober-primary text-white'
                            : 'bg-slate-800 text-slate-400'
                        "
                      >
                        {{ i + 1 }}
                      </span>
                      <span class="min-w-0 flex-1 truncate text-sm font-medium text-slate-100">
                        {{ nodeDisplayTitle(node) }}
                      </span>
                      <span
                        v-if="labelSourcePill(node.labelSource)"
                        class="shrink-0 rounded px-1 py-0.5 text-[10px] font-bold uppercase"
                        :class="
                          node.labelSource === 'ai'
                            ? 'bg-amber-900/30 text-amber-400'
                            : 'bg-cyan-900/30 text-cyan-400'
                        "
                      >
                        {{ labelSourcePill(node.labelSource) }}
                      </span>
                      <button
                        v-if="isCompactNode(node)"
                        type="button"
                        class="shrink-0 rounded px-1 py-0.5 text-[10px] font-bold uppercase text-onbober-primary hover:bg-onbober-primary/10"
                        title="Expand inline"
                        aria-label="Expand inline"
                        @click.stop="emit('expandNode', node)"
                      >
                        Expand
                      </button>
                      <button
                        v-if="canPreviewCalleeFlow(node)"
                        type="button"
                        class="flex shrink-0 items-center gap-1 rounded p-0.5 text-onbober-primary hover:bg-onbober-primary/10"
                        title="View code flow (scan labels)"
                        aria-label="View code flow"
                        @click.stop="emit('previewCompacted', node)"
                      >
                        <svg class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                          <path d="M8 2L2 5.5v5L8 14l6-3.5v-5L8 2z" />
                          <path d="M2 5.5L8 9l6-3.5M8 9v5" />
                        </svg>
                        <span class="hidden font-semibold sm:inline">View flow</span>
                      </button>
                      <span
                        v-if="hasHiddenChildren(node.id)"
                        class="shrink-0 rounded bg-onbober-primary/20 px-1.5 py-0.5 text-xs font-bold uppercase text-onbober-primary"
                      >
                        +
                      </span>
                      <span class="shrink-0 text-xs font-medium uppercase text-slate-500">{{ kindLabel(node.kind) }}</span>
                    </div>
                    <p v-if="node.summary && node.summary !== nodeDisplayTitle(node)" class="mt-1.5 text-sm leading-snug text-slate-400">
                      {{ node.summary }}
                    </p>
                    <p v-if="node.code" class="mt-1 line-clamp-2 break-all font-mono text-xs leading-relaxed text-slate-400">{{ node.code }}</p>
                  </button>
                </li>
              </ol>
            </template>

            <div v-else class="flex h-full justify-center pt-2">
              <button
                type="button"
                class="flex h-6 w-6 items-center justify-center rounded text-slate-400 transition hover:bg-slate-800 hover:text-white"
                aria-label="Show steps panel"
                title="Show steps"
                @click="toggleTracePanel"
              >
                <svg class="h-3.5 w-3.5" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 2l4 4-4 4"/></svg>
              </button>
            </div>
          </div>

          <ResizeHandle
            v-if="tracePanelOpen && !isMobile"
            label="Resize steps panel"
            @pointerdown="onTraceResize"
          />

          <div
            ref="panViewport"
            class="relative min-w-0 flex-1 overflow-hidden bg-slate-950/50"
            title="Drag background to pan, scroll to zoom"
          >
            <div
              v-if="mappingFullFlow"
              class="absolute inset-0 z-10 flex items-center justify-center bg-slate-950/70 backdrop-blur-sm px-4"
            >
              <BeaverFlowLoader
                mode="progress"
                :active="mappingFullFlow"
                :progress="mappingProgress"
                compact
                class="w-full max-w-2xl"
              />
            </div>
            <div ref="panContent" class="absolute inset-0 p-4">
              <div
                ref="mermaidContainer"
                class="mermaid-flow mx-auto flex min-h-[180px] min-w-fit items-center justify-center"
              />
            </div>
            <p v-if="renderError" class="pointer-events-none absolute bottom-2 left-0 right-0 text-center text-xs text-red-400">
              {{ renderError }}
            </p>
          </div>
        </div>
      </template>
    </div>

    <ResizeHandle
      v-if="hasDetailContent && detailPanelOpen && !isMobile"
      label="Resize details panel"
      @pointerdown="onDetailResize"
    />

    <aside
      class="detail-panel flex shrink-0 flex-col border-l border-slate-800 bg-slate-900"
      :class="hasDetailContent && detailPanelOpen ? 'detail-panel-open' : 'detail-panel-closed'"
      :style="hasDetailContent && detailPanelOpen ? { width: `${detailWidth}px` } : { width: '2.75rem' }"
    >
      <template v-if="hasDetailContent && detailPanelOpen">
        <div class="flex shrink-0 items-center justify-between border-b border-slate-800 px-3 py-2">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-400">Details</span>
          <button
            type="button"
            class="rounded p-1 text-slate-500 hover:bg-slate-800 hover:text-white"
            aria-label="Hide details panel"
            title="Hide details"
            @click="toggleDetailPanel"
          >
            <svg class="h-3 w-3" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 2l4 4-4 4"/></svg>
          </button>
        </div>
        <div v-if="selectedNode" class="min-h-0 flex-1 overflow-y-auto p-4">
          <h3 class="text-lg font-semibold text-white">{{ nodeDisplayTitle(selectedNode) }}</h3>

          <section class="mt-3">
            <span class="inline-block rounded bg-green-900/40 px-2 py-0.5 text-xs font-bold uppercase text-green-400">
              verified structure
            </span>
            <span class="ml-1 inline-block rounded border border-slate-700 px-2 py-0.5 text-xs uppercase text-slate-400">
              {{ selectedNode.kind }}
            </span>
            <pre
              v-if="selectedNode.code"
              class="mt-2 overflow-x-auto rounded border border-slate-800 bg-slate-950 p-3 font-mono text-sm leading-relaxed text-slate-200"
            >{{ selectedNode.code }}</pre>
            <p v-else class="mt-2 text-sm text-slate-500">No source snippet for this step.</p>
            <p
              v-if="canPreviewCalleeFlow(selectedNode)"
              class="mt-3 text-sm leading-relaxed text-slate-400"
            >
              <template v-if="isCompactNode(selectedNode)">
                This step is folded. Use <strong class="font-medium text-slate-300">View code flow</strong> to see the full callee graph with scan labels, or <strong class="font-medium text-slate-300">Expand inline</strong> to unfold it in the main diagram.
              </template>
              <template v-else>
                Use <strong class="font-medium text-slate-300">View code flow</strong> to open the callee graph with scan labels.
              </template>
            </p>
          </section>

          <section v-if="showEnrichedSummary && enrichedBadge" class="mt-4">
            <span
              class="inline-block rounded px-2 py-0.5 text-xs font-bold"
              :class="
                selectedNode.labelSource === 'ai'
                  ? 'border border-dashed border-amber-600/50 bg-amber-900/20 text-amber-400'
                  : 'border border-dashed border-cyan-600/50 bg-cyan-900/20 text-cyan-400'
              "
            >
              {{ enrichedBadge }}
            </span>
            <p
              class="mt-2 text-sm leading-relaxed"
              :class="selectedNode.labelSource === 'ai' ? 'text-amber-300/90' : 'text-cyan-300/90'"
            >
              {{ selectedNode.summary }}
            </p>
          </section>

          <section class="mt-4">
            <button
              v-if="!deepDiveOpen"
              type="button"
              class="w-full rounded-md border border-slate-700 px-3 py-2.5 text-left text-sm text-slate-300 transition hover:border-onbober-primary/50 hover:text-white"
              @click="openDeepDive"
            >
              Explain this step
            </button>
            <template v-else>
              <span class="inline-block rounded border border-dashed border-amber-600/50 bg-amber-900/20 px-2 py-0.5 text-xs font-bold uppercase text-amber-400">
                Deep dive
              </span>
              <LoadingStatus
                class="mt-3"
                :active="showDetailLoading"
                :phrases="AI_LOADING_PHRASES"
              />
              <p v-if="detailError" class="mt-2 text-sm text-red-400">{{ detailError }}</p>
              <p v-if="detail?.mock" class="mt-2 rounded border border-amber-800/50 bg-amber-900/20 px-2 py-1 text-xs text-amber-300">
                Watsonx explanation unavailable
              </p>

              <div v-if="verifiedExplanation || (!detailStreaming && detail?.explanation && !inferredExplanation)" class="mt-3">
                <span class="inline-block rounded bg-green-900/40 px-2 py-0.5 text-xs font-bold uppercase text-green-400">
                  Verified
                </span>
                <MarkdownView :content="verifiedExplanation || detail?.explanation || ''" class="mt-2" />
                <span v-if="detailStreaming && !inferredExplanation" class="inline-block w-1.5 animate-pulse bg-onbober-primary">|</span>
              </div>

              <div v-if="inferredExplanation" class="mt-4">
                <span class="inline-block rounded border border-dashed border-amber-600/50 bg-amber-900/20 px-2 py-0.5 text-xs font-bold uppercase text-amber-400">
                  Inferred
                </span>
                <p class="mt-1 text-xs text-slate-400">Based on codebase patterns — not verified against external API docs.</p>
                <MarkdownView :content="inferredExplanation" class="mt-2 [&_.md-prose]:text-amber-300/90" />
                <span v-if="detailStreaming" class="inline-block w-1.5 animate-pulse bg-onbober-primary">|</span>
              </div>

              <div v-if="hasEvidence" class="mt-4">
                <button
                  type="button"
                  class="text-left text-sm text-slate-400 hover:text-slate-200"
                  @click="evidenceOpen = !evidenceOpen"
                >
                  {{ evidenceOpen ? 'Hide' : 'Show' }} evidence used ({{ detail?.evidence?.length }})
                </button>
                <ul v-if="evidenceOpen" class="mt-2 space-y-1">
                  <li
                    v-for="(line, idx) in detail?.evidence"
                    :key="idx"
                    class="truncate font-mono text-xs text-slate-400"
                    :title="line"
                  >
                    {{ line }}
                  </li>
                </ul>
              </div>
            </template>
          </section>

          <p
            v-if="!fullyExpanded && hasHiddenChildren(selectedNode.id)"
            class="mt-3 text-sm text-onbober-primary"
          >
            Click again to reveal the next step{{ selectedNode.kind === 'branch' ? 's (branches)' : '' }}.
          </p>

          <div class="mt-4 flex flex-col gap-2">
            <button
              v-if="selectedNode && canPreviewCalleeFlow(selectedNode)"
              type="button"
              class="w-full rounded-md border border-onbober-primary/40 bg-onbober-primary/5 px-3 py-2.5 text-left text-sm font-medium text-onbober-primary transition hover:border-onbober-primary/60 hover:bg-onbober-primary/10"
              @click="emit('previewCompacted', selectedNode)"
            >
              View code flow
              <span class="mt-0.5 block text-xs font-normal text-slate-400">Full callee graph · scan labels only</span>
            </button>
            <button
              v-if="selectedNode && isCompactNode(selectedNode)"
              type="button"
              class="text-left text-sm text-slate-400 transition hover:text-onbober-primary"
              @click="emit('expandNode', selectedNode)"
            >
              Expand inline
            </button>
            <button
              v-if="selectedNode.calleeFile"
              type="button"
              class="text-left text-sm text-onbober-primary hover:underline"
              @click="emit('goToDefinition', selectedNode.calleeFile!, selectedNode.calleeSymbol ?? selectedNode.label, selectedNode.calleeLine ?? 1)"
            >
              Go to {{ selectedNode.calleeSymbol }} ({{ selectedNode.calleeFile }}:{{ selectedNode.calleeLine }})
            </button>
            <button
              v-if="selectedNode.file || detail?.file"
              type="button"
              class="text-left text-sm text-slate-400 hover:text-onbober-primary hover:underline"
              @click="emit('viewSource', selectedNode.file ?? detail?.file, selectedNode.line ?? undefined)"
            >
              View source{{ selectedNode.line ? ` (line ${selectedNode.line})` : '' }}
            </button>
          </div>
        </div>
      </template>

      <div v-else class="flex h-full justify-center pt-2">
        <button
          type="button"
          class="flex h-6 w-6 items-center justify-center rounded text-slate-400 transition hover:bg-slate-800 hover:text-white"
          aria-label="Show details panel"
          title="Show details"
          @click="toggleDetailPanel"
        >
          <svg class="h-3.5 w-3.5" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 2l4 4-4 4"/></svg>
        </button>
      </div>
    </aside>
  </div>
</template>

<style scoped>
/* Steps panel: animate width so flex space moves with it */
.steps-panel {
  overflow: hidden;
  transition: width 0.22s ease, opacity 0.22s ease;
}
.steps-panel-closed {
  width: 2.75rem !important;
  min-width: 2.75rem;
  opacity: 1;
  pointer-events: auto;
}

/* Details panel: animate width from the right */
.detail-panel {
  overflow: hidden;
  transition: width 0.22s ease, opacity 0.22s ease;
}
.detail-panel-closed {
  width: 2.75rem !important;
  min-width: 2.75rem;
  opacity: 1;
  pointer-events: auto;
}

.mermaid-flow :deep(svg) {
  max-width: none;
  height: auto;
}

.mermaid-flow :deep(svg text),
.mermaid-flow :deep(svg .label),
.mermaid-flow :deep(svg .nodeLabel),
.mermaid-flow :deep(svg foreignObject div) {
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, 'Courier New', monospace !important;
  font-size: 15px !important;
}

.mermaid-flow :deep(svg .edgeLabel) {
  pointer-events: none;
}

.mermaid-flow :deep(g.node.is-selected rect),
.mermaid-flow :deep(g.node.is-selected polygon),
.mermaid-flow :deep(g.node.is-selected path) {
  stroke: #ff3366 !important;
  stroke-width: 3px !important;
}

.mermaid-flow :deep(g.node.is-selected .label),
.mermaid-flow :deep(g.node.is-selected .nodeLabel) {
  fill: #fff !important;
}
</style>
