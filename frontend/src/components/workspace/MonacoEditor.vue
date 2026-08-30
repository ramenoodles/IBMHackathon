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

function applyMonacoTheme(monacoInstance: any): void {
  monacoInstance.editor.defineTheme('onbober-dark', {
    base: 'vs-dark',
    inherit: true,
    rules: [
      { token: '', foreground: 'e2e8f0', background: '0f172a' },
      { token: 'comment', foreground: '64748b' },
      { token: 'keyword', foreground: 'c084fc' },
      { token: 'string', foreground: 'a5f3fc' },
      { token: 'number', foreground: 'f9a8d4' },
      { token: 'type', foreground: '93c5fd' },
      { token: 'delimiter', foreground: 'cbd5e1' },
    ],
    colors: {
      'editor.background': '#0f172a',
      'editor.foreground': '#e2e8f0',
      'editorCursor.foreground': '#f8fafc',
      'editor.selectionBackground': '#334155',
      'editor.inactiveSelectionBackground': '#1e293b',
      'editor.lineHighlightBackground': '#111827',
      'editor.lineHighlightBorder': '#1e293b',
      'editorLineNumber.foreground': '#64748b',
      'editorLineNumber.activeForeground': '#e2e8f0',
      'editorIndentGuide.background': '#334155',
      'editorIndentGuide.activeBackground': '#94a3b8',
      'editorWhitespace.foreground': '#334155',
      'editorGutter.background': '#0f172a',
      'scrollbarSlider.background': '#475569',
      'scrollbarSlider.hoverBackground': '#64748b',
      'scrollbarSlider.activeBackground': '#94a3b8',
      'minimap.background': '#0f172a',
      'minimapSlider.background': '#334155',
      'minimapSlider.hoverBackground': '#475569',
      'minimapSlider.activeBackground': '#64748b',
      'editorSuggestWidget.background': '#0f172a',
      'editorSuggestWidget.border': '#334155',
      'editorWidget.background': '#111827',
      'editorWidget.border': '#334155',
      'list.activeSelectionBackground': '#334155',
      'list.hoverBackground': '#1e293b',
      'editorHoverWidget.background': '#111827',
      'editorHoverWidget.border': '#334155',
    },
  })

  monacoInstance.editor.setTheme('onbober-dark')
}

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
  applyMonacoTheme(monaco)

  editor = monaco.editor.create(container.value, {
    value: props.content,
    language: monacoLanguage(props.language),
    theme: 'onbober-dark',
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
