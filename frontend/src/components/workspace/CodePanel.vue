<script setup lang="ts">
/**
 * Minimal code viewer for the source modal — no duplicate symbol chips.
 */
import { highlightCode, languageFromPath } from '@/composables/useShiki'
import { onMounted, ref, watch } from 'vue'

const props = defineProps<{
  workspacePath: string
  filePath: string
}>()

const highlightedHtml = ref('')
const loading = ref(false)

async function loadFile(): Promise<void> {
  if (!props.workspacePath || !props.filePath) return
  loading.value = true
  try {
    const params = new URLSearchParams({ workspace: props.workspacePath, path: props.filePath })
    const res = await fetch(`/api/file?${params}`)
    if (!res.ok) return
    const data = (await res.json()) as { content: string; language: string }
    highlightedHtml.value = await highlightCode(data.content, data.language || languageFromPath(props.filePath))
  } finally {
    loading.value = false
  }
}

onMounted(loadFile)
watch(() => props.filePath, loadFile)
</script>

<template>
  <div class="text-sm">
    <p v-if="loading" class="text-slate-500">Loading...</p>
    <div v-else class="shiki-container overflow-x-auto" v-html="highlightedHtml" />
  </div>
</template>

<style scoped>
.shiki-container :deep(pre) {
  margin: 0;
  padding: 0;
  background: transparent !important;
}
</style>
