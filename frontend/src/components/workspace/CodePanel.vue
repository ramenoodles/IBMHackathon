<script setup lang="ts">
/**
 * Minimal code viewer for the source modal — no duplicate symbol chips.
 */
import { highlightCode, languageFromPath } from '@/composables/useShiki'
import { api } from '@/api'
import { onMounted, ref, watch } from 'vue'

const props = defineProps<{
  workspaceId: string
  filePath: string
}>()

const highlightedHtml = ref('')
const loading = ref(false)

async function loadFile(): Promise<void> {
  if (!props.workspaceId || !props.filePath) return
  loading.value = true
  try {
    const data = await api.file(props.workspaceId, props.filePath)
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
