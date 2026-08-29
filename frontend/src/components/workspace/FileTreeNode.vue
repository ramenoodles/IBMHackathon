<script setup lang="ts">
/**
 * Recursive file tree node with lazy-loaded folder children.
 */
import { ref } from 'vue'
import FileTreeNode from './FileTreeNode.vue'
import type { TreeEntry } from '@/composables/useFileTree'
import { fetchTreeEntries } from '@/composables/useFileTree'

const props = defineProps<{
  /** Workspace root path. */
  workspacePath: string
  /** This node's entry metadata. */
  entry: TreeEntry
  /** Currently selected file path relative to workspace. */
  selectedPath: string
  /** Nesting depth for indentation. */
  depth?: number
}>()

const emit = defineEmits<{
  /** Emitted when a file is selected. */
  select: [path: string]
}>()

const expanded = ref(false)
const children = ref<TreeEntry[]>([])
const loading = ref(false)
const loaded = ref(false)
const error = ref<string | null>(null)

const depth = props.depth ?? 0
const indent = `${depth * 12 + 8}px`

/**
 * Toggle folder expansion and lazy-load children on first open.
 */
async function onFolderClick(): Promise<void> {
  expanded.value = !expanded.value
  if (expanded.value && !loaded.value) {
    await loadChildren()
  }
}

/** Fetch child entries for this folder from the API. */
async function loadChildren(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    children.value = await fetchTreeEntries(props.workspacePath, props.entry.path)
    loaded.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load folder'
  } finally {
    loading.value = false
  }
}

/** Emit file selection to parent. */
function onFileClick(): void {
  emit('select', props.entry.path)
}
</script>

<template>
  <li>
    <button
      v-if="entry.isDir"
      type="button"
      class="flex w-full items-center gap-1 rounded py-1 pr-2 text-left hover:bg-slate-800"
      :style="{ paddingLeft: indent }"
      @click="onFolderClick"
    >
      <span class="w-4 shrink-0 text-xs text-slate-500">{{ expanded ? '▼' : '▶' }}</span>
      <span class="shrink-0">📁</span>
      <span class="truncate text-slate-300">{{ entry.name }}</span>
    </button>

    <button
      v-else
      type="button"
      class="flex w-full items-center gap-1 rounded py-1 pr-2 text-left hover:bg-slate-800"
      :class="selectedPath === entry.path ? 'bg-onbober-primary/10 text-onbober-primary' : 'text-slate-300'"
      :style="{ paddingLeft: `calc(${indent} + 16px)` }"
      @click="onFileClick"
    >
      <span class="shrink-0">📄</span>
      <span class="truncate">{{ entry.name }}</span>
    </button>

    <ul v-if="entry.isDir && expanded" class="space-y-0.5">
      <li v-if="loading" class="py-1 text-xs text-slate-500" :style="{ paddingLeft: `calc(${indent} + 24px)` }">
        Loading...
      </li>
      <li v-else-if="error" class="py-1 text-xs text-red-400" :style="{ paddingLeft: `calc(${indent} + 24px)` }">
        {{ error }}
      </li>
      <FileTreeNode
        v-for="child in children"
        :key="child.path"
        :workspace-path="workspacePath"
        :entry="child"
        :selected-path="selectedPath"
        :depth="depth + 1"
        @select="emit('select', $event)"
      />
    </ul>
  </li>
</template>
