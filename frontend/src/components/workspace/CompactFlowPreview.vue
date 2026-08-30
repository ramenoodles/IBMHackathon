<script setup lang="ts">
/**
 * Read-only preview of a compacted callee's flow graph.
 */
import { ref, watch } from 'vue'
import Modal from '@/components/ui/Modal.vue'
import { api } from '@/api'
import type { FlowEdge, FlowNode, GraphRootPayload } from '@/types/flowGraph'
import { useFlowMermaid } from '@/composables/useFlowMermaid'
import { enrichSymbolNodes } from '@/utils/flowGraphEnrich'

const props = withDefaults(
  defineProps<{
    open: boolean
    workspaceId: string
    filePath: string
    symbol: string
    line?: number
    userContext: GraphRootPayload['userContext']
    scanOnly?: boolean
  }>(),
  { scanOnly: true },
)

const emit = defineEmits<{
  close: []
}>()

const loading = ref(false)
const error = ref<string | null>(null)
const nodes = ref<FlowNode[]>([])
const edges = ref<FlowEdge[]>([])

const mermaidContainer = ref<HTMLElement | null>(null)
const { setContainer, renderError } = useFlowMermaid(
  () => nodes.value,
  () => edges.value,
  '',
  undefined,
  undefined,
  'scan',
)

watch(mermaidContainer, (el) => setContainer(el))

async function loadPreview(): Promise<void> {
  if (!props.open || !props.filePath || !props.symbol) return
  loading.value = true
  error.value = null
  nodes.value = []
  edges.value = []
  try {
    const graph = await api.graph({
      workspaceId: props.workspaceId,
      filePath: props.filePath,
      symbol: props.symbol,
      userContext: props.userContext,
    })
    nodes.value = graph.nodes
    edges.value = graph.edges

    if (!props.scanOnly) {
      void enrichSymbolNodes(
        nodes.value,
        {
          workspaceId: props.workspaceId,
          filePath: props.filePath,
          symbol: props.symbol,
          userContext: props.userContext,
        },
        graph.nodes.map((n) => n.id),
      ).then((result) => {
        const byId = new Map(nodes.value.map((n) => [n.id, n]))
        for (const patch of result.patches) {
          const node = byId.get(patch.id)
          if (!node) continue
          if (patch.title) node.title = patch.title
          if (patch.summary) node.summary = patch.summary
          if (patch.labelSource) {
            node.labelSource = patch.labelSource as FlowNode['labelSource']
            if (patch.labelSource === 'ai' || patch.labelSource === 'heuristic') {
              node.confidence = 'inferred'
            }
          }
        }
        nodes.value = [...nodes.value]
      })
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load compacted flow'
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.open, props.filePath, props.symbol, props.scanOnly] as const,
  () => {
    if (props.open) void loadPreview()
  },
  { immediate: true },
)
</script>

<template>
  <Modal
    :open="open"
    :title="`Code flow: ${symbol}()`"
    size="xl"
    @close="emit('close')"
  >
    <div class="flex h-full min-h-0 flex-col px-5 pb-5">
      <p v-if="filePath" class="mb-3 shrink-0 font-mono text-xs text-slate-500">
        {{ filePath }}<span v-if="line">:{{ line }}</span>
      </p>
      <p v-if="loading" class="py-8 text-center text-sm text-slate-400">Loading code flow…</p>
      <p v-else-if="error" class="py-8 text-center text-sm text-red-400">{{ error }}</p>
      <p v-else-if="renderError" class="py-8 text-center text-sm text-red-400">{{ renderError }}</p>
      <div
        v-else
        class="min-h-0 flex-1 overflow-auto rounded-lg border border-slate-800 bg-slate-950/80 p-4"
      >
        <div
          ref="mermaidContainer"
          class="mermaid-flow-preview mx-auto flex min-h-[200px] min-w-fit items-center justify-center"
        />
      </div>
      <p class="mt-3 shrink-0 text-xs text-slate-500">
        Read-only · scan labels (no AI)
      </p>
    </div>
  </Modal>
</template>

<style scoped>
.mermaid-flow-preview :deep(svg) {
  max-width: none;
  height: auto;
}

.mermaid-flow-preview :deep(svg text),
.mermaid-flow-preview :deep(svg .label),
.mermaid-flow-preview :deep(svg .nodeLabel),
.mermaid-flow-preview :deep(svg foreignObject div) {
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, 'Courier New', monospace !important;
  font-size: 14px !important;
}
</style>
