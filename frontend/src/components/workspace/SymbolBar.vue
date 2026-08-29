<script setup lang="ts">
/**
 * Horizontal function picker — single entry point to load a flow graph.
 */
import { onMounted, ref, watch } from 'vue'
import { extractFunctions } from '@/composables/useSymbols'

const props = defineProps<{
  workspacePath: string
  filePath: string
  activeSymbol: string
  symbolNames?: string[]
}>()

const emit = defineEmits<{
  pick: [symbol: string]
}>()

const functions = ref<string[]>([])
const loading = ref(false)

async function load(): Promise<void> {
  if (!props.workspacePath || !props.filePath) {
    functions.value = []
    return
  }
  if (props.symbolNames?.length) {
    functions.value = props.symbolNames
    return
  }
  loading.value = true
  try {
    const params = new URLSearchParams({ workspace: props.workspacePath, path: props.filePath })
    const res = await fetch(`/api/file?${params}`)
    if (!res.ok) return
    const data = (await res.json()) as { content: string }
    functions.value = extractFunctions(data.content, props.filePath)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => [props.filePath, props.symbolNames] as const, load)
</script>

<template>
  <div v-if="filePath" class="flex items-center gap-2 border-b border-slate-800 bg-slate-900/80 px-4 py-2">
    <span class="shrink-0 text-xs text-slate-500">Trace</span>
    <div class="symbol-scroll flex min-w-0 flex-1 gap-1.5 overflow-x-auto pb-0.5">
      <span v-if="loading" class="text-xs text-slate-500">Loading...</span>
      <button
        v-for="fn in functions"
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
      <span v-if="!loading && !functions.length" class="text-xs text-slate-500">No functions in this file</span>
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
