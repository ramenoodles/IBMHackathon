<script setup lang="ts">
/**
 * AI explanation drawer with streaming text, verification badges, and Mermaid diagrams.
 */
import { computed, ref, watch } from 'vue'
import { useMermaid } from '@/composables/useMermaid'

const props = defineProps<{
  /** Full streamed markdown/text from the LLM. */
  text: string
  /** Whether the backend is currently streaming tokens. */
  isStreaming: boolean
  /** Optional stream error message. */
  error: string | null
  /** True when the backend served a mock/demo response. */
  isMock: boolean
}>()

const emit = defineEmits<{
  /** Emitted when the collapse toggle is clicked. */
  toggle: []
}>()

const { mermaidCode, renderError, setContainer } = useMermaid(computed(() => props.text))
const mermaidContainer = ref<HTMLElement | null>(null)

watch(mermaidContainer, (el) => setContainer(el))

/**
 * Convert simple markdown badges and headings to HTML for display.
 * @param markdown - Raw LLM markdown text without mermaid blocks.
 */
function renderMarkdown(markdown: string): string {
  const withoutMermaid = markdown.replace(/```mermaid[\s\S]*?```/g, '')
  return withoutMermaid
    .replace(/\[VERIFIED\]/g, '<span class="badge verified">VERIFIED</span>')
    .replace(/\[INFERRED\]/g, '<span class="badge inferred">INFERRED</span>')
    .replace(/^## (.+)$/gm, '<h3 class="text-base font-semibold text-white mt-4 mb-2">$1</h3>')
    .replace(/^# (.+)$/gm, '<h2 class="text-lg font-bold text-white mt-4 mb-2">$1</h2>')
    .replace(/\n/g, '<br/>')
}
</script>

<template>
  <aside class="flex h-full w-96 flex-col border-l border-slate-800 bg-slate-900">
    <div class="flex items-center justify-between border-b border-slate-800 px-3 py-2">
      <span class="text-xs font-semibold uppercase tracking-wide text-slate-400">AI Analysis</span>
      <button
        type="button"
        class="text-slate-400 hover:text-white md:hidden"
        aria-label="Close AI panel"
        @click="emit('toggle')"
      >
        ✕
      </button>
    </div>

    <div v-if="isMock" class="border-b border-amber-800/50 bg-amber-900/20 px-3 py-2 text-xs text-amber-300">
      Ollama offline — showing demo response
    </div>

    <div class="flex-1 overflow-y-auto p-4 text-sm leading-relaxed text-slate-300">
      <p v-if="!text && !isStreaming" class="text-slate-500">
        Open a file, then click a symbol chip or identifier in the code to analyze it.
      </p>
      <p v-if="error" class="mb-3 text-red-400">{{ error }}</p>
      <div v-if="text" class="prose-invert" v-html="renderMarkdown(text)" />
      <p v-if="isStreaming" class="mt-2 animate-pulse text-onbober-primary">Streaming...</p>

      <div v-if="mermaidCode || renderError" class="mt-6">
        <h3 class="mb-2 text-sm font-semibold text-slate-200">Architecture Diagram</h3>
        <p v-if="renderError" class="mb-2 text-xs text-amber-400">{{ renderError }}</p>
        <pre
          v-if="renderError && mermaidCode"
          class="mb-2 overflow-x-auto rounded border border-slate-700 bg-slate-950 p-2 text-xs text-slate-400"
        >{{ mermaidCode }}</pre>
        <div ref="mermaidContainer" class="overflow-x-auto rounded border border-slate-700 bg-slate-950 p-2" />
      </div>
    </div>
  </aside>
</template>

<style scoped>
:deep(.badge) {
  display: inline-block;
  margin-right: 0.25rem;
  border-radius: 0.25rem;
  padding: 0.125rem 0.375rem;
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.05em;
}
:deep(.badge.verified) {
  background: rgba(34, 197, 94, 0.2);
  color: #4ade80;
}
:deep(.badge.inferred) {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}
</style>
