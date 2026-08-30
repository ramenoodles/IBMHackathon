<script setup lang="ts">
/**
 * Compact filename search for the workspace file explorer sidebar.
 */
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { buildFileIndex, searchFiles } from '@/composables/useFileIndex'
import type { TreeEntry } from '@/composables/useFileTree'

const props = defineProps<{
  workspaceId: string
}>()

const emit = defineEmits<{
  select: [path: string]
}>()

const query = ref('')
const debouncedQuery = ref('')
const indexing = ref(false)
const fileIndex = ref<TreeEntry[]>([])
const showDropdown = ref(false)
const highlightedIndex = ref(0)
const rootEl = ref<HTMLElement | null>(null)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const results = computed(() => {
  if (!debouncedQuery.value.trim()) return []
  return searchFiles(fileIndex.value, debouncedQuery.value, 20)
})

const totalMatches = computed(() => {
  if (!debouncedQuery.value.trim() || fileIndex.value.length === 0) return 0
  return searchFiles(fileIndex.value, debouncedQuery.value, Number.MAX_SAFE_INTEGER).length
})

const hasMore = computed(() => totalMatches.value > results.value.length)

watch(query, (value) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    debouncedQuery.value = value
  }, 150)
})

watch(debouncedQuery, async (value) => {
  const trimmed = value.trim()
  if (!trimmed) {
    showDropdown.value = false
    highlightedIndex.value = 0
    return
  }

  showDropdown.value = true
  highlightedIndex.value = 0

  if (fileIndex.value.length === 0 && !indexing.value) {
    indexing.value = true
    try {
      const index = await buildFileIndex(props.workspaceId)
      fileIndex.value = index
    } finally {
      indexing.value = false
    }
  }
})

watch(
  () => props.workspaceId,
  () => {
    query.value = ''
    debouncedQuery.value = ''
    fileIndex.value = []
    showDropdown.value = false
    highlightedIndex.value = 0
  },
)

function selectResult(path: string): void {
  emit('select', path)
  query.value = ''
  debouncedQuery.value = ''
  showDropdown.value = false
  highlightedIndex.value = 0
}

function onKeydown(event: KeyboardEvent): void {
  if (!showDropdown.value || results.value.length === 0) {
    if (event.key === 'Escape') {
      query.value = ''
      debouncedQuery.value = ''
      showDropdown.value = false
    }
    return
  }

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    highlightedIndex.value = (highlightedIndex.value + 1) % results.value.length
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    highlightedIndex.value =
      (highlightedIndex.value - 1 + results.value.length) % results.value.length
  } else if (event.key === 'Enter') {
    event.preventDefault()
    const selected = results.value[highlightedIndex.value]
    if (selected) selectResult(selected.path)
  } else if (event.key === 'Escape') {
    showDropdown.value = false
    query.value = ''
    debouncedQuery.value = ''
  }
}

function onClickOutside(event: MouseEvent): void {
  if (rootEl.value && !rootEl.value.contains(event.target as Node)) {
    showDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('mousedown', onClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onClickOutside)
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<template>
  <div ref="rootEl" class="relative border-b border-slate-800 px-2 py-1.5">
    <div class="relative">
      <svg
        class="pointer-events-none absolute left-2 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-500"
        viewBox="0 0 16 16"
        fill="none"
        stroke="currentColor"
        stroke-width="1.75"
        aria-hidden="true"
      >
        <circle cx="7" cy="7" r="4.5" />
        <path d="M10.5 10.5L14 14" stroke-linecap="round" />
      </svg>
      <input
        v-model="query"
        type="search"
        class="w-full rounded border border-slate-700 bg-slate-800/60 py-1 pl-6 pr-2 text-xs text-slate-200 placeholder:text-slate-500 focus:border-onbober-primary/50 focus:outline-none focus:ring-1 focus:ring-onbober-primary/30"
        placeholder="Search files…"
        autocomplete="off"
        spellcheck="false"
        @focus="debouncedQuery.trim() && (showDropdown = true)"
        @keydown="onKeydown"
      />
    </div>

    <div
      v-if="showDropdown && debouncedQuery.trim()"
      class="absolute left-2 right-2 top-full z-50 mt-0.5 max-h-52 overflow-y-auto rounded border border-slate-700 bg-slate-900 shadow-lg"
    >
      <p v-if="indexing" class="px-2 py-1.5 text-xs text-slate-500">Indexing…</p>
      <template v-else-if="results.length > 0">
        <button
          v-for="(result, idx) in results"
          :key="result.path"
          type="button"
          class="flex w-full flex-col px-2 py-1.5 text-left hover:bg-slate-800"
          :class="idx === highlightedIndex ? 'bg-slate-800' : ''"
          @mousedown.prevent="selectResult(result.path)"
        >
          <span class="truncate text-xs text-slate-200">{{ result.name }}</span>
          <span class="truncate text-[10px] text-slate-500">{{ result.path }}</span>
        </button>
        <p v-if="hasMore" class="border-t border-slate-800 px-2 py-1 text-[10px] text-slate-500">
          +{{ totalMatches - results.length }} more
        </p>
      </template>
      <p v-else class="px-2 py-1.5 text-xs text-slate-500">No matches</p>
    </div>
  </div>
</template>
