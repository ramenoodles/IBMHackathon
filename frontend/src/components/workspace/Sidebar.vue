<script setup lang="ts">
/**
 * Slim file explorer sidebar.
 */
import { onMounted, ref, watch } from 'vue'
import FileTreeNode from '@/components/workspace/FileTreeNode.vue'
import { fetchTreeEntries, type TreeEntry } from '@/composables/useFileTree'

const props = defineProps<{
  workspacePath: string
  selectedPath: string
}>()

const emit = defineEmits<{
  select: [path: string]
  toggle: []
}>()

const entries = ref<TreeEntry[]>([])
const loading = ref(false)

async function loadRoot(): Promise<void> {
  if (!props.workspacePath) return
  loading.value = true
  try {
    entries.value = await fetchTreeEntries(props.workspacePath)
  } finally {
    loading.value = false
  }
}

onMounted(loadRoot)
watch(() => props.workspacePath, loadRoot)
</script>

<template>
  <aside class="flex h-full w-44 shrink-0 flex-col border-r border-slate-800 bg-slate-900">
    <div class="flex items-center justify-between border-b border-slate-800 px-3 py-2">
      <span class="text-xs font-medium text-slate-400">Files</span>
      <button
        type="button"
        class="rounded px-1 text-slate-500 hover:bg-slate-800 hover:text-white"
        aria-label="Hide file explorer"
        title="Hide file explorer"
        @click="emit('toggle')"
      >
        ‹
      </button>
    </div>
    <div class="flex-1 overflow-y-auto p-1.5 text-sm">
      <p v-if="loading" class="px-2 py-1 text-xs text-slate-500">Loading...</p>
      <ul v-else class="space-y-0.5">
        <FileTreeNode
          v-for="entry in entries"
          :key="entry.path"
          :workspace-path="workspacePath"
          :entry="entry"
          :selected-path="selectedPath"
          :depth="0"
          @select="emit('select', $event)"
        />
      </ul>
    </div>
  </aside>
</template>
