<script setup lang="ts">
/**
 * Slim file explorer sidebar.
 */
import { onMounted, provide, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import FileSearchBar from '@/components/workspace/FileSearchBar.vue'
import FileTreeNode from '@/components/workspace/FileTreeNode.vue'
import { fetchTreeEntries, type TreeEntry } from '@/composables/useFileTree'
import { ApiError } from '@/api'
import { clearUserContext } from '@/store/userContext'

const props = defineProps<{
  workspaceId: string
  selectedPath: string
  open: boolean
  width?: number
}>()

const emit = defineEmits<{
  select: [path: string]
  toggle: []
}>()

const entries = ref<TreeEntry[]>([])
const loading = ref(false)
const collapseAllTick = ref(0)

provide('fileTreeCollapseAll', collapseAllTick)

function collapseAll(): void {
  collapseAllTick.value++
}

const router = useRouter()

async function loadRoot(): Promise<void> {
  if (!props.workspaceId) return
  loading.value = true
  try {
    entries.value = await fetchTreeEntries(props.workspaceId)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      // Workspace no longer exists (backend restarted). Clear stale session and
      // send the user back to onboarding to pick a workspace again.
      clearUserContext()
      router.push({ name: 'onboarding' })
      return
    }
    throw err
  } finally {
    loading.value = false
  }
}

onMounted(loadRoot)
watch(() => props.workspaceId, loadRoot)
</script>

<template>
  <aside
    class="sidebar-panel flex h-full shrink-0 flex-col border-r border-slate-800 bg-slate-900"
    :class="open ? 'sidebar-open' : 'sidebar-closed'"
    :style="open ? { width: `${width ?? 176}px` } : {}"
  >
    <div class="flex items-center justify-between border-b border-slate-800 px-3 py-2">
      <span class="text-xs font-medium text-slate-400">Files</span>
      <div class="flex items-center gap-0.5">
        <button
          type="button"
          class="rounded p-1 text-slate-500 hover:bg-slate-800 hover:text-white"
          aria-label="Collapse all folders"
          title="Collapse all folders"
          @click="collapseAll"
        >
          <svg class="h-3.5 w-3.5" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M2 4h8M2 8h8M6 2v2M6 8v2" />
          </svg>
        </button>
        <button
          type="button"
          class="rounded p-1 text-slate-500 hover:bg-slate-800 hover:text-white"
          aria-label="Hide file explorer"
          title="Hide file explorer"
          @click="emit('toggle')"
        >
          <svg class="h-3.5 w-3.5" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M8 2L4 6l4 4" />
          </svg>
        </button>
      </div>
    </div>
    <FileSearchBar :workspace-id="workspaceId" @select="emit('select', $event)" />
    <div class="flex-1 overflow-y-auto p-1.5 text-sm">
      <p v-if="loading" class="px-2 py-1 text-xs text-slate-500">Loading...</p>
      <ul v-else class="space-y-0.5">
        <FileTreeNode
          v-for="entry in entries"
          :key="entry.path"
          :workspace-id="workspaceId"
          :entry="entry"
          :selected-path="selectedPath"
          :depth="0"
          @select="emit('select', $event)"
        />
      </ul>
    </div>
  </aside>
</template>

<style scoped>
.sidebar-panel {
  overflow: hidden;
  transition: width 0.22s ease, opacity 0.22s ease;
}

.sidebar-closed {
  width: 0 !important;
  opacity: 0;
  pointer-events: none;
}
</style>
