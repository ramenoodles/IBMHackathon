<script setup lang="ts">
/**
 * Clean graph-first workspace: explorer + timeline + on-demand source modal.
 */
import { computed, ref } from 'vue'
import Sidebar from '@/components/workspace/Sidebar.vue'
import SymbolBar from '@/components/workspace/SymbolBar.vue'
import FlowCanvas from '@/components/workspace/FlowCanvas.vue'
import FileFlowBrief from '@/components/workspace/FileFlowBrief.vue'
import FlowWarmOverlay from '@/components/workspace/FlowWarmOverlay.vue'
import BranchPrompt from '@/components/workspace/BranchPrompt.vue'
import Modal from '@/components/ui/Modal.vue'
import CodePanel from '@/components/workspace/CodePanel.vue'
import { useWorkspaceLayout } from '@/composables/useWorkspaceLayout'
import { useFlowGraphCache } from '@/composables/useFlowGraphCache'
import { useFlowGraph } from '@/composables/useFlowGraph'
import { useSymbolBrief } from '@/composables/useSymbolBrief'
import { useFileFlowWarm } from '@/composables/useFileFlowWarm'
import { useNodeDetail } from '@/composables/useNodeDetail'
import { userContext } from '@/store/userContext'
import type { FlowNode } from '@/types/flowGraph'

type WorkspacePhase = 'idle' | 'brief' | 'warming' | 'tracing'

const { sidebarOpen, isMobile, toggleSidebar } = useWorkspaceLayout()

const graphCache = useFlowGraphCache()
const {
  nodes,
  edges,
  rootId,
  loading,
  enriching,
  expanding,
  error: graphError,
  isMock,
  loadRoot,
  activateSymbol,
  expandNode,
  revealFromNode,
  hasHiddenChildren,
  prefetchAroundNode,
  reset: resetGraph,
} = useFlowGraph(graphCache)

const { symbols, loading: symbolsLoading, error: symbolsError, isLargeFile, load: loadSymbols, reset: resetSymbols } =
  useSymbolBrief()
const { warming, progress, warmFile } = useFileFlowWarm(graphCache)
const { detail, loading: detailLoading, loadDetail } = useNodeDetail()

const workspacePhase = ref<WorkspacePhase>('idle')
const selectedPath = ref('')
const symbol = ref('')
const selectedNodeId = ref('')
const sourceOpen = ref(false)
const sourcePath = ref('')
const warmedSymbolNames = ref<string[]>([])

const branchPromptOpen = ref(false)
const branchNode = ref<FlowNode | null>(null)

const showSymbolBar = computed(
  () =>
    selectedPath.value &&
    (workspacePhase.value === 'tracing' || workspacePhase.value === 'warming'),
)

const symbolBarNames = computed(() => {
  if (warmedSymbolNames.value.length) return warmedSymbolNames.value
  return symbols.value.map((s) => s.name)
})

function graphPayload() {
  return {
    workspacePath: userContext.value.workspacePath,
    filePath: selectedPath.value,
    symbol: symbol.value,
    userContext: { ...userContext.value },
  }
}

async function onSelectFile(path: string): Promise<void> {
  selectedPath.value = path
  symbol.value = ''
  selectedNodeId.value = ''
  warmedSymbolNames.value = []
  resetGraph()

  if (graphCache.isFileWarmed(path)) {
    warmedSymbolNames.value = graphCache.listSymbolsForFile(path)
    workspacePhase.value = 'tracing'
    if (warmedSymbolNames.value[0]) {
      symbol.value = warmedSymbolNames.value[0]
      activateSymbol(warmedSymbolNames.value[0], {
        workspacePath: userContext.value.workspacePath,
        filePath: path,
        symbol: warmedSymbolNames.value[0],
        userContext: { ...userContext.value },
      })
    }
    return
  }

  workspacePhase.value = 'brief'
  await loadSymbols(userContext.value.workspacePath, path)
}

function onDeclineFlowInit(): void {
  workspacePhase.value = 'tracing'
}

async function onConfirmFlowInit(selected: string[]): Promise<void> {
  if (!selectedPath.value || !selected.length) return
  workspacePhase.value = 'warming'
  warmedSymbolNames.value = selected
  const base = {
    workspacePath: userContext.value.workspacePath,
    filePath: selectedPath.value,
    userContext: { ...userContext.value },
  }
  await warmFile(base, selected)
  workspacePhase.value = 'tracing'
  symbol.value = selected[0]!
  selectedNodeId.value = ''
  activateSymbol(selected[0]!, { ...base, symbol: selected[0]! })
}

async function onPickSymbol(name: string): Promise<void> {
  if (!selectedPath.value) return
  symbol.value = name
  selectedNodeId.value = ''
  const payload = { ...graphPayload(), symbol: name }
  if (!activateSymbol(name, payload)) {
    await loadRoot(payload)
  }
}

function onRevealNode(node: FlowNode): void {
  revealFromNode(node.id)
}

function onSelectNode(node: FlowNode): void {
  selectedNodeId.value = node.id
  prefetchAroundNode(node.id)
  void loadDetail({
    workspace: userContext.value.workspacePath,
    nodeId: node.id,
    symbol: symbol.value,
    file: node.file ?? selectedPath.value,
    line: node.line,
    title: node.title ?? node.label,
    confidence: node.confidence,
    code: node.code,
    experience: userContext.value.experienceLevel,
    language: userContext.value.primaryLanguage,
  })
}

function onViewSource(file?: string): void {
  sourcePath.value = file ?? selectedPath.value
  sourceOpen.value = true
}

function onGoToDefinition(file: string, _line: number): void {
  sourcePath.value = file
  sourceOpen.value = true
}

function onExpandNode(node: FlowNode): void {
  if (node.childCount > 3) {
    branchNode.value = node
    branchPromptOpen.value = true
  } else {
    void expandNode(node.id, graphPayload(), node.childCount || 3)
  }
}

function onConfirmExpand(limit: number): void {
  if (!branchNode.value) return
  void expandNode(branchNode.value.id, graphPayload(), limit)
  branchNode.value = null
}

function fileName(): string {
  return selectedPath.value.split('/').pop() ?? selectedPath.value
}
</script>

<template>
  <div class="flex h-screen flex-col bg-slate-950">
    <header class="flex items-center gap-3 border-b border-slate-800 bg-slate-900 px-4 py-2.5">
      <button
        type="button"
        class="rounded p-1 text-slate-400 hover:bg-slate-800 hover:text-white"
        aria-label="Toggle files"
        @click="toggleSidebar"
      >
        ☰
      </button>
      <span class="text-sm font-semibold text-slate-200">
        On<span class="text-onbober-primary">Bober</span>
      </span>
      <span v-if="selectedPath" class="truncate text-xs text-slate-500">
        {{ fileName() }}
      </span>
    </header>

    <div class="flex min-h-0 flex-1">
      <Sidebar
        v-show="sidebarOpen || !isMobile"
        :workspace-path="userContext.workspacePath"
        :selected-path="selectedPath"
        @select="onSelectFile"
        @toggle="toggleSidebar"
      />

      <div class="relative flex min-w-0 flex-1 flex-col">
        <SymbolBar
          v-if="showSymbolBar"
          :workspace-path="userContext.workspacePath"
          :file-path="selectedPath"
          :active-symbol="symbol"
          :symbol-names="symbolBarNames"
          @pick="onPickSymbol"
        />

        <FileFlowBrief
          v-if="workspacePhase === 'brief' && selectedPath"
          :file-name="fileName()"
          :symbols="symbols"
          :loading="symbolsLoading"
          :error="symbolsError"
          :is-large-file="isLargeFile"
          @confirm="onConfirmFlowInit"
          @decline="onDeclineFlowInit"
        />

        <FlowCanvas
          v-else-if="workspacePhase === 'tracing' || workspacePhase === 'warming'"
          :nodes="nodes"
          :edges="edges"
          :root-id="rootId"
          :loading="loading"
          :enriching="enriching"
          :expanding="expanding"
          :error="graphError"
          :is-mock="isMock"
          :symbol="symbol"
          :selected-node-id="selectedNodeId"
          :detail="detail"
          :detail-loading="detailLoading"
          :has-hidden-children="hasHiddenChildren"
          @select-node="onSelectNode"
          @reveal-node="onRevealNode"
          @expand-node="onExpandNode"
          @view-source="onViewSource()"
          @go-to-definition="onGoToDefinition"
        />

        <div
          v-else
          class="flex flex-1 flex-col items-center justify-center gap-2 p-8 text-center text-slate-500"
        >
          <p class="text-lg font-medium text-slate-300">Select a file to begin</p>
          <p class="max-w-sm text-sm">Choose a file from the explorer to see traceable symbols.</p>
        </div>

        <FlowWarmOverlay
          :open="workspacePhase === 'warming' && warming"
          :file-name="fileName()"
          :done="progress.done"
          :total="progress.total"
          :current-symbol="progress.currentSymbol"
        />
      </div>
    </div>

    <Modal :open="sourceOpen" title="Source" @close="sourceOpen = false">
      <div class="max-h-[60vh] overflow-auto">
        <CodePanel
          v-if="sourcePath || selectedPath"
          :workspace-path="userContext.workspacePath"
          :file-path="sourcePath || selectedPath"
        />
      </div>
    </Modal>

    <BranchPrompt
      :open="branchPromptOpen"
      :node-label="branchNode?.label ?? ''"
      :child-count="branchNode?.childCount ?? 0"
      @close="branchPromptOpen = false"
      @expand="onConfirmExpand"
    />
  </div>
</template>
