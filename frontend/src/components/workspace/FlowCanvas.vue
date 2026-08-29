<script setup lang="ts">
/**
 * Flow graph canvas: progressive reveal + Mermaid flowchart + execution trace.
 */
import type { FlowEdge, FlowNode, NodeDetail } from '@/types/flowGraph'
import { computed, ref, watch } from 'vue'
import { nodeDisplayTitle, useFlowMermaid } from '@/composables/useFlowMermaid'
import { useWorkspaceLayout } from '@/composables/useWorkspaceLayout'
import { edgeOrder } from '@/utils/flowGraphUtils'

const {
  tracePanelOpen,
  detailPanelOpen,
  toggleTracePanel,
  toggleDetailPanel,
} = useWorkspaceLayout()

const props = defineProps<{
  nodes: FlowNode[]
  edges: FlowEdge[]
  rootId: string
  loading: boolean
  enriching: boolean
  expanding: boolean
  mappingFullFlow: boolean
  fullyExpanded: boolean
  error: string | null
  isMock: boolean
  symbol: string
  selectedNodeId: string
  detail: NodeDetail | null
  detailLoading: boolean
  hasHiddenChildren: (nodeId: string) => boolean
}>()

const emit = defineEmits<{
  selectNode: [node: FlowNode]
  expandNode: [node: FlowNode]
  revealNode: [node: FlowNode]
  showFullFlow: []
  viewSource: []
  goToDefinition: [file: string, line: number]
}>()

const mermaidContainer = ref<HTMLElement | null>(null)

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

const { renderError, setContainer } = useFlowMermaid(
  () => props.nodes,
  () => props.edges,
  () => props.selectedNodeId,
  onNodeClick,
)

watch(mermaidContainer, (el) => setContainer(el))

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
const hasDetailContent = computed(
  () => Boolean(props.selectedNodeId && (props.detail || props.detailLoading || selectedNode.value)),
)
const showExpandHint = computed(
  () =>
    !props.fullyExpanded &&
    !props.mappingFullFlow &&
    props.nodes.length === 1 &&
    props.rootId &&
    props.hasHiddenChildren(props.rootId),
)

function kindLabel(kind: string): string {
  const labels: Record<string, string> = {
    entry: 'entry', call: 'call', branch: 'branch', return: 'return', assign: 'assign', loop: 'loop',
  }
  return labels[kind] ?? kind
}
</script>

<template>
  <div class="flex h-full min-w-0 flex-1">
    <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <div v-if="isMock" class="shrink-0 bg-amber-900/20 px-4 py-1.5 text-xs text-amber-300">
        No scan data for this symbol
      </div>

      <div v-if="loading" class="flex flex-1 items-center justify-center text-sm text-slate-500">
        Mapping execution path...
      </div>

      <div v-else-if="!symbol" class="flex flex-1 flex-col items-center justify-center gap-2 p-8 text-center">
        <p class="text-lg font-medium text-slate-300">Pick a function to trace</p>
        <p class="max-w-sm text-sm text-slate-500">Select a file, then choose a function from the bar above.</p>
      </div>

      <div v-else-if="!nodes.length" class="flex flex-1 items-center justify-center text-sm text-slate-500">
        No flow data for <span class="ml-1 font-mono text-onbober-primary">{{ symbol }}</span>
      </div>

      <template v-else>
        <div class="flex shrink-0 items-center justify-between border-b border-slate-800 px-4 py-2">
          <p class="text-xs text-slate-500">
            Tracing <span class="font-mono text-slate-300">{{ symbol }}</span>
            <span v-if="mappingFullFlow" class="ml-2 text-onbober-primary">Mapping full flow...</span>
            <span v-else-if="expanding" class="ml-2 text-onbober-primary">Expanding...</span>
          </p>
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-1">
              <button
                type="button"
                class="rounded px-2 py-1 text-[10px] font-medium transition"
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
                class="rounded px-2 py-1 text-[10px] font-medium transition"
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
            <button
              v-if="!fullyExpanded && !mappingFullFlow"
              type="button"
              class="rounded-md border border-slate-700 px-2.5 py-1 text-[11px] font-medium text-slate-300 transition hover:border-onbober-primary/50 hover:text-white"
              @click="emit('showFullFlow')"
            >
              Show full flow
            </button>
            <div class="flex items-center gap-3 text-[10px] text-slate-500">
              <span class="flex items-center gap-1"><span class="inline-block h-2 w-2 rounded-sm border border-green-500 bg-green-900/40" /> structure</span>
              <span class="flex items-center gap-1"><span class="inline-block h-2 w-2 rounded-sm border border-dashed border-amber-500 bg-amber-900/40" /> AI label</span>
            </div>
          </div>
        </div>

        <p v-if="showExpandHint" class="shrink-0 border-b border-slate-800 bg-slate-900/60 px-4 py-2 text-xs text-slate-400">
          Click the entry step to reveal the next part of the flow.
        </p>

        <p v-if="error" class="shrink-0 px-4 py-2 text-sm text-red-400">{{ error }}</p>

        <div class="flex min-h-0 flex-1">
          <button
            v-if="!tracePanelOpen"
            type="button"
            class="flex w-8 shrink-0 flex-col items-center justify-center gap-1 border-r border-slate-800 bg-slate-900/60 text-slate-500 transition hover:bg-slate-800 hover:text-slate-200"
            aria-label="Show steps panel"
            title="Show steps"
            @click="toggleTracePanel"
          >
            <span class="text-sm">›</span>
            <span class="text-[9px] font-medium uppercase tracking-wide [writing-mode:vertical-rl]">Steps</span>
          </button>

          <div
            v-else
            class="flex w-72 shrink-0 flex-col border-r border-slate-800 bg-slate-900/40"
          >
            <div class="flex shrink-0 items-center justify-between border-b border-slate-800 px-2 py-1.5">
              <span class="text-[10px] font-medium uppercase tracking-wide text-slate-500">Steps</span>
              <button
                type="button"
                class="rounded px-1.5 py-0.5 text-slate-500 hover:bg-slate-800 hover:text-white"
                aria-label="Hide steps panel"
                title="Hide steps"
                @click="toggleTracePanel"
              >
                ‹
              </button>
            </div>
            <ol class="min-h-0 flex-1 overflow-y-auto p-2">
            <li v-for="(node, i) in orderedNodes" :key="node.id" class="mb-1">
              <button
                type="button"
                class="w-full rounded-md border px-2 py-1.5 text-left transition"
                :class="
                  selectedNodeId === node.id
                    ? 'border-onbober-primary/50 bg-onbober-primary/5'
                    : 'border-transparent hover:border-slate-700 hover:bg-slate-800/50'
                "
                @click="onNodeClick(node)"
              >
                <div class="flex items-center gap-2">
                  <span
                    class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-[10px] font-bold"
                    :class="
                      selectedNodeId === node.id
                        ? 'bg-onbober-primary text-white'
                        : 'bg-slate-800 text-slate-400'
                    "
                  >
                    {{ i + 1 }}
                  </span>
                  <span class="min-w-0 flex-1 truncate text-xs font-medium text-slate-200">
                    {{ nodeDisplayTitle(node) }}
                  </span>
                  <span
                    v-if="hasHiddenChildren(node.id)"
                    class="shrink-0 rounded bg-onbober-primary/20 px-1 text-[9px] font-bold uppercase text-onbober-primary"
                  >
                    +
                  </span>
                  <span class="shrink-0 text-[10px] uppercase text-slate-600">{{ kindLabel(node.kind) }}</span>
                </div>
                <p v-if="node.summary && node.summary !== nodeDisplayTitle(node)" class="mt-1 text-xs text-slate-400">
                  {{ node.summary }}
                </p>
                <p v-if="node.code" class="mt-0.5 truncate font-mono text-[10px] text-slate-600">{{ node.code }}</p>
              </button>
            </li>
          </ol>
          </div>

          <div class="relative min-w-0 flex-1 overflow-auto bg-slate-950/50 p-4">
            <div
              v-if="mappingFullFlow"
              class="absolute inset-0 z-10 flex items-center justify-center bg-slate-950/70 backdrop-blur-sm"
            >
              <p class="text-sm text-slate-300">Mapping full flow for {{ symbol }}...</p>
            </div>
            <div
              ref="mermaidContainer"
              class="mermaid-flow mx-auto flex min-h-[180px] min-w-fit items-center justify-center"
            />
            <p v-if="renderError" class="mt-2 text-center text-xs text-red-400">{{ renderError }}</p>
          </div>
        </div>
      </template>
    </div>

    <button
      v-if="hasDetailContent && !detailPanelOpen"
      type="button"
      class="flex w-8 shrink-0 flex-col items-center justify-center gap-1 border-l border-slate-800 bg-slate-900/60 text-slate-500 transition hover:bg-slate-800 hover:text-slate-200"
      aria-label="Show details panel"
      title="Show details"
      @click="toggleDetailPanel"
    >
      <span class="text-sm">‹</span>
      <span class="text-[9px] font-medium uppercase tracking-wide [writing-mode:vertical-rl]">Details</span>
    </button>

    <aside
      v-else-if="hasDetailContent && detailPanelOpen"
      class="flex w-72 shrink-0 flex-col border-l border-slate-800 bg-slate-900"
    >
      <div class="flex shrink-0 items-center justify-between border-b border-slate-800 px-3 py-2">
        <span class="text-[10px] font-medium uppercase tracking-wide text-slate-500">Details</span>
        <button
          type="button"
          class="rounded px-1.5 py-0.5 text-slate-500 hover:bg-slate-800 hover:text-white"
          aria-label="Hide details panel"
          title="Hide details"
          @click="toggleDetailPanel"
        >
          ›
        </button>
      </div>
      <div class="min-h-0 flex-1 overflow-y-auto p-4">
      <p v-if="detailLoading" class="text-sm text-slate-500">Loading...</p>
      <template v-else>
        <h3 class="font-semibold text-white">{{ selectedNode ? nodeDisplayTitle(selectedNode) : detail?.title }}</h3>
        <span class="mt-1 inline-block rounded bg-green-900/40 px-1.5 py-0.5 text-[10px] font-bold uppercase text-green-400">
          verified structure
        </span>
        <p v-if="selectedNode?.summary" class="mt-2 text-sm leading-relaxed text-amber-300/90">{{ selectedNode.summary }}</p>
        <pre
          v-if="selectedNode?.code"
          class="mt-3 overflow-x-auto rounded border border-slate-800 bg-slate-950 p-2 font-mono text-[11px] leading-relaxed text-slate-300"
        >{{ selectedNode.code }}</pre>
        <p v-if="detail" class="mt-3 text-sm leading-relaxed text-slate-300">{{ detail.explanation }}</p>
        <p
          v-if="selectedNode && !fullyExpanded && hasHiddenChildren(selectedNode.id)"
          class="mt-3 text-xs text-onbober-primary"
        >
          Click again to reveal the next step{{ selectedNode.kind === 'branch' ? 's (branches)' : '' }}.
        </p>
        <div class="mt-4 flex flex-col gap-2">
          <button
            v-if="selectedNode?.calleeFile"
            type="button"
            class="text-left text-xs text-onbober-primary hover:underline"
            @click="emit('goToDefinition', selectedNode.calleeFile!, selectedNode.calleeLine ?? 1)"
          >
            Go to {{ selectedNode.calleeSymbol }} ({{ selectedNode.calleeFile }}:{{ selectedNode.calleeLine }})
          </button>
          <button
            v-if="selectedNode?.file || detail?.file"
            type="button"
            class="text-left text-xs text-slate-400 hover:text-onbober-primary hover:underline"
            @click="emit('viewSource')"
          >
            View source{{ selectedNode?.line ? ` (line ${selectedNode.line})` : '' }}
          </button>
        </div>
      </template>
      </div>
    </aside>
  </div>
</template>

<style scoped>
.mermaid-flow :deep(svg) {
  max-width: 100%;
  height: auto;
}
</style>
