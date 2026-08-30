<script setup lang="ts">
/**
 * Horizontal symbol picker — purely presentational.
 * The parent controls which page of symbols is shown via symbolNames.
 */
import { computed, ref } from 'vue'

const props = defineProps<{
  workspaceId: string
  filePath: string
  activeSymbol: string
  symbolNames?: string[]
  hasNextPage?: boolean
  hasPrevPage?: boolean
  currentPage?: number
  totalPages?: number
}>()

const emit = defineEmits<{
  pick: [symbol: string]
  nextPage: []
  prevPage: []
  goToPage: [page: number]
}>()

// 0-based internally, display is 1-based
const page = computed(() => props.currentPage ?? 0)
const total = computed(() => props.totalPages ?? 1)

/** Up to 5 page numbers centred around current page. */
const pageWindow = computed<number[]>(() => {
  if (total.value <= 0) return []
  const WINDOW = 5
  const half = Math.floor(WINDOW / 2)
  let start = Math.max(0, page.value - half)
  let end = Math.min(total.value - 1, start + WINDOW - 1)
  // shift window left if at the end
  if (end - start < WINDOW - 1) start = Math.max(0, end - WINDOW + 1)
  const pages: number[] = []
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
})

const inputValue = ref('')
const inputFocused = ref(false)

function onInputKeydown(e: KeyboardEvent): void {
  if (e.key !== 'Enter') return
  const n = parseInt(inputValue.value, 10)
  if (!isNaN(n) && n >= 1 && n <= total.value) {
    emit('goToPage', n - 1) // convert to 0-based
  }
  inputValue.value = ''
  ;(e.target as HTMLInputElement).blur()
}

function onInputBlur(): void {
  inputFocused.value = false
  inputValue.value = ''
}
</script>

<template>
  <div v-if="filePath" class="flex items-center gap-2 border-b border-slate-800 bg-slate-900/80 px-4 py-2">
    <span class="shrink-0 text-xs text-slate-500">Trace</span>

    <!-- Symbol buttons -->
    <div class="symbol-scroll flex min-w-0 flex-1 gap-1.5 overflow-x-auto pb-0.5">
      <button
        v-for="fn in symbolNames"
        :key="fn"
        type="button"
        class="shrink-0 rounded-md px-3 py-1 text-xs font-medium transition"
        :class="
          activeSymbol === fn
            ? 'bg-onbober-primary text-white'
            : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
        "
        @click="emit('pick', fn)"
      >
        {{ fn }}
      </button>
      <span v-if="!symbolNames?.length" class="text-xs text-slate-500">No symbols on this page</span>
    </div>

    <!-- Pagination control -->
    <div v-if="filePath" class="flex shrink-0 items-center gap-0.5 rounded-lg border border-slate-700 bg-slate-900 px-1 py-0.5">
      <!-- Prev arrow -->
      <button
        type="button"
        class="rounded px-1.5 py-0.5 text-xs text-slate-400 transition disabled:cursor-not-allowed disabled:opacity-30 hover:enabled:bg-slate-700 hover:enabled:text-slate-200"
        :disabled="!hasPrevPage"
        title="Previous page"
        @click="emit('prevPage')"
      >
        ‹
      </button>

      <!-- Page number buttons -->
      <button
        v-for="p in pageWindow"
        :key="p"
        type="button"
        class="min-w-[22px] rounded px-1.5 py-0.5 text-xs font-medium transition"
        :class="
          p === page
            ? 'bg-onbober-primary text-white'
            : 'text-slate-400 hover:bg-slate-700 hover:text-slate-200'
        "
        :title="`Page ${p + 1}`"
        @click="emit('goToPage', p)"
      >
        {{ p + 1 }}
      </button>

      <!-- Next arrow -->
      <button
        type="button"
        class="rounded px-1.5 py-0.5 text-xs text-slate-400 transition disabled:cursor-not-allowed disabled:opacity-30 hover:enabled:bg-slate-700 hover:enabled:text-slate-200"
        :disabled="!hasNextPage"
        title="Next page"
        @click="emit('nextPage')"
      >
        ›
      </button>

      <!-- Divider -->
      <span class="mx-1 text-slate-700">|</span>

      <!-- Direct page entry -->
      <input
        v-model="inputValue"
        type="text"
        inputmode="numeric"
        class="w-8 rounded border border-slate-700 bg-slate-800 px-1 py-0.5 text-center text-xs text-slate-300 outline-none transition focus:border-onbober-primary focus:text-white"
        :placeholder="`${page + 1}`"
        title="Go to page"
        @focus="inputFocused = true"
        @blur="onInputBlur"
        @keydown="onInputKeydown"
      />
    </div>
  </div>
</template>

<style scoped>
.symbol-scroll {
  scrollbar-width: thin;
  scrollbar-color: #475569 transparent;
}
.symbol-scroll::-webkit-scrollbar {
  height: 5px;
}
.symbol-scroll::-webkit-scrollbar-thumb {
  background: #475569;
  border-radius: 4px;
}
.symbol-scroll::-webkit-scrollbar-track {
  background: transparent;
}
</style>
