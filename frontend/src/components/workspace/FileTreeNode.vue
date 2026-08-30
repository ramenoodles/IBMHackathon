<script setup lang="ts">
/**
 * Recursive file tree node with lazy-loaded folder children.
 */
import { inject, ref, watch, type Ref } from 'vue'
import FileTreeNode from './FileTreeNode.vue'
import type { TreeEntry } from '@/composables/useFileTree'
import { fetchTreeEntries } from '@/composables/useFileTree'
import { getFileIconUrl, getFolderIconUrl } from '@/composables/useFileIcon'

const props = defineProps<{
  /** Workspace root path. */
  workspaceId: string
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

const collapseAllTick = inject<Ref<number>>('fileTreeCollapseAll', ref(0))
watch(collapseAllTick, () => {
  expanded.value = false
})

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
    children.value = await fetchTreeEntries(props.workspaceId, props.entry.path)
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
      <svg
        class="h-3 w-3 shrink-0 text-slate-500 transition-transform duration-150"
        :class="expanded ? 'rotate-90' : ''"
        viewBox="0 0 12 12"
        fill="none"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M4 2l4 4-4 4" />
      </svg>
      <img :src="getFolderIconUrl(entry.name, expanded)" alt="" class="h-4 w-4 shrink-0" />
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
      <img :src="getFileIconUrl(entry.name)" alt="" class="h-4 w-4 shrink-0" />
      <span class="truncate">{{ entry.name }}</span>
    </button>

    <Transition name="tree-children">
      <ul v-if="entry.isDir && expanded" class="tree-children space-y-0.5">
        <li v-if="loading" class="py-1 text-xs text-slate-500" :style="{ paddingLeft: `calc(${indent} + 24px)` }">
          Loading...
        </li>
        <li v-else-if="error" class="py-1 text-xs text-red-400" :style="{ paddingLeft: `calc(${indent} + 24px)` }">
          {{ error }}
        </li>
        <FileTreeNode
          v-for="child in children"
          :key="child.path"
          :workspace-id="workspaceId"
          :entry="child"
          :selected-path="selectedPath"
          :depth="depth + 1"
          @select="emit('select', $event)"
        />
      </ul>
    </Transition>
  </li>
</template>

<style scoped>
.tree-children {
  overflow: hidden;
}

.tree-children-enter-active,
.tree-children-leave-active {
  transition: opacity 0.14s ease, transform 0.14s ease, max-height 0.16s ease;
  max-height: 40rem;
}

.tree-children-enter-from,
.tree-children-leave-to {
  opacity: 0;
  transform: translateY(-3px);
  max-height: 0;
}

.tree-children-enter-to,
.tree-children-leave-from {
  opacity: 1;
  transform: translateY(0);
  max-height: 40rem;
}
</style>
