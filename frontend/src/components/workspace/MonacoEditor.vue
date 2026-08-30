<script setup lang="ts">
/**
 * Read-only Monaco editor with syntax highlighting, line numbers,
 * and optional scroll-to / decoration of a specific highlighted line.
 *
 * Monaco is loaded at runtime from CDN via @monaco-editor/loader so the
 * editor and monaco variables are typed as `any` — the real types live in
 * `monaco-editor` which we deliberately don't bundle.
 */
import loader from '@monaco-editor/loader'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps<{
  content: string
  language?: string
  highlightLine?: number
}>()

const container = ref<HTMLElement | null>(null)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let editor: any = null
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let monaco: any = null
let decorationIds: string[] = []

/** Map common language ids to Monaco's language ids */
function monacoLanguage(lang?: string): string {
  const map: Record<string, string> = {
    c: 'c',
    cpp: 'cpp',
    go: 'go',
    rust: 'rust',
    python: 'python',
    bash: 'shell',
    sh: 'shell',
    ts: 'typescript',
    js: 'javascript',
    json: 'json',
    yaml: 'yaml',
    text: 'plaintext',
  }
  return map[lang ?? ''] ?? lang ?? 'plaintext'
}

function applyHighlight(line?: number): void {
  if (!editor || !monaco || !line) {
    if (editor) {
      decorationIds = editor.deltaDecorations(decorationIds, [])
    }
    return
  }

  decorationIds = editor.deltaDecorations(decorationIds, [
    {
      range: new monaco.Range(line, 1, line, 1),
      options: {
        isWholeLine: true,
        className: 'monaco-highlighted-line',
        overviewRuler: {
          color: '#ff3366',
          // Use numeric value for OverviewRulerLane.Full (4)
          position: 4,
        },
      },
    },
  ])

  // ScrollType.Smooth = 0
  editor.revealLineInCenter(line, 0)
}

onMounted(async () => {
  if (!container.value) return

  // Configure CDN path for Monaco workers
  loader.config({
    paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs' },
  })

  monaco = await loader.init()

  editor = monaco.editor.create(container.value, {
    value: props.content,
    language: monacoLanguage(props.language),
    theme: 'vs-dark',
    readOnly: true,
    automaticLayout: true,
    scrollBeyondLastLine: false,
    minimap: { enabled: true },
    lineNumbers: 'on',
    renderLineHighlight: 'all',
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, 'Courier New', monospace",
    fontSize: 13,
    lineHeight: 22,
    padding: { top: 12, bottom: 12 },
    folding: true,
    wordWrap: 'off',
    scrollbar: {
      verticalScrollbarSize: 8,
      horizontalScrollbarSize: 8,
    },
  })

  applyHighlight(props.highlightLine)
})

watch(() => props.content, (val) => {
  if (editor) editor.setValue(val)
})

watch(() => props.language, (val) => {
  if (editor && monaco) {
    const model = editor.getModel()
    if (model) monaco.editor.setModelLanguage(model, monacoLanguage(val))
  }
})

watch(() => props.highlightLine, applyHighlight)

onBeforeUnmount(() => {
  editor?.dispose()
  editor = null
})
</script>

<template>
  <div ref="container" class="monaco-container" />
</template>

<style scoped>
.monaco-container {
  width: 100%;
  height: 100%;
}
</style>

<style>
/* Global — Monaco injects its own DOM outside the scoped tree */
.monaco-highlighted-line {
  background: rgba(255, 51, 102, 0.15) !important;
  border-left: 3px solid #ff3366 !important;
}
</style>
