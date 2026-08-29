<script setup lang="ts">
/**
 * Clean graph-first workspace: explorer + timeline + on-demand source modal.
 */
import { ref } from 'vue'
import Sidebar from '@/components/workspace/Sidebar.vue'
import SymbolBar from '@/components/workspace/SymbolBar.vue'
import FlowCanvas from '@/components/workspace/FlowCanvas.vue'
import BranchPrompt from '@/components/workspace/BranchPrompt.vue'
import Modal from '@/components/ui/Modal.vue'
import CodePanel from '@/components/workspace/CodePanel.vue'
import { useWorkspaceLayout } from '@/composables/useWorkspaceLayout'
import { useFlowGraph } from '@/composables/useFlowGraph'
import { useNodeDetail } from '@/composables/useNodeDetail'
import { userContext } from '@/store/userContext'
import type { FlowNode } from '@/types/flowGraph'

const { sidebarOpen, isMobile, toggleSidebar } = useWorkspaceLayout()

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
  expandNode,
  revealFromNode,
  hasHiddenChildren,
} = useFlowGraph()

const { detail, loading: detailLoading, loadDetail } = useNodeDetail()

const selectedPath = ref('')
const symbol = ref('')
const selectedNodeId = ref('')
const sourceOpen = ref(false)
const sourcePath = ref('')

const branchPromptOpen = ref(false)
const branchNode = ref<FlowNode | null>(null)

function onSelectFile(path: string): void {
  selectedPath.value = path
  symbol.value = ''
  selectedNodeId.value = ''
}

function graphPayload() {
  return {
    workspacePath: userContext.value.workspacePath,
    filePath: selectedPath.value,
    symbol: symbol.value,
    userContext: { ...userContext.value },
  }
}

async function onPickSymbol(name: string): Promise<void> {
  if (!selectedPath.value) return
  symbol.value = name
  selectedNodeId.value = ''
  await loadRoot({ ...graphPayload(), symbol: name })
}

function onRevealNode(node: FlowNode): void {
  revealFromNode(node.id)
}

function onSelectNode(node: FlowNode): void {
  selectedNodeId.value = node.id
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

      <div class="flex min-w-0 flex-1 flex-col">
        <SymbolBar
          :workspace-path="userContext.workspacePath"
          :file-path="selectedPath"
          :active-symbol="symbol"
          @pick="onPickSymbol"
        />

        <FlowCanvas
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
