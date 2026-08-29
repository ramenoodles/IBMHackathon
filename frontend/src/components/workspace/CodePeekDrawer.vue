<script setup lang="ts">
/**
 * Collapsible bottom drawer showing source code at a node's file/line.
 */
import { highlightCode, languageFromPath } from '@/composables/useShiki'
import { ref, watch } from 'vue'

const props = defineProps<{
  open: boolean
  workspaceId: string
  filePath: string
  highlightLine?: number
}>()

const emit = defineEmits<{
  close: []
}>()

const html = ref('')
const loading = ref(false)

async function load(): Promise<void> {
  if (!props.open || !props.workspaceId || !props.filePath) return
  loading.value = true
  try {
    const params = new URLSearchParams({ path: props.filePath })
    const res = await fetch(`/api/workspaces/${encodeURIComponent(props.workspaceId)}/file?${params}`)
    if (!res.ok) return
    const data = (await res.json()) as { content: string; language: string }
    html.value = await highlightCode(data.content, data.language || languageFromPath(props.filePath))
  } finally {
    loading.value = false
  }
}

watch(() => [props.open, props.filePath], load)
</script>

<template>
  <div
    v-if="open"
    class="border-t border-slate-800 bg-slate-950"
    style="height: 12rem"
  >
    <div class="flex items-center justify-between border-b border-slate-800 px-4 py-1">
      <span class="truncate text-xs text-slate-400">{{ filePath }}</span>
      <button type="button" class="text-xs text-slate-500 hover:text-white" @click="emit('close')">Close</button>
    </div>
    <div class="h-[calc(100%-2rem)] overflow-auto p-3 text-xs">
      <p v-if="loading" class="text-slate-500">Loading...</p>
      <div v-else class="shiki-peek" v-html="html" />
    </div>
  </div>
</template>

<style scoped>
.shiki-peek :deep(pre) {
  margin: 0;
  background: transparent !important;
}
</style>
