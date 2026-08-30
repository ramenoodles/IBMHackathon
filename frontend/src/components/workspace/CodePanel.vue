<script setup lang="ts">
/**
 * Code viewer using Monaco Editor — syntax highlighting, line numbers,
 * and optional jump-to / highlight of a specific line.
 */
import { languageFromPath } from '@/composables/useShiki'
import { api } from '@/api'
import { onMounted, ref, watch } from 'vue'
import MonacoEditor from './MonacoEditor.vue'

const props = defineProps<{
  workspaceId: string
  filePath: string
  highlightLine?: number
}>()

const content = ref('')
const language = ref('text')
const loading = ref(false)

async function loadFile(): Promise<void> {
  if (!props.workspaceId || !props.filePath) return
  loading.value = true
  try {
    const data = await api.file(props.workspaceId, props.filePath)
    content.value = data.content
    language.value = data.language || languageFromPath(props.filePath)
  } finally {
    loading.value = false
  }
}

onMounted(loadFile)
watch(() => props.filePath, loadFile)
</script>

<template>
  <div class="flex h-full flex-col">
    <p v-if="loading" class="flex flex-1 items-center justify-center text-sm text-slate-500">
      Loading...
    </p>
    <MonacoEditor
      v-else
      :content="content"
      :language="language"
      :highlight-line="highlightLine"
      class="flex-1"
    />
  </div>
</template>
