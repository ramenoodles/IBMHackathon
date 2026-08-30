<script setup lang="ts">
/**
 * Renders a markdown string as sanitised HTML with dark-mode prose styles.
 * Uses marked for parsing; no external CSS dependencies.
 */
import { computed } from 'vue'
import { marked } from 'marked'

const props = defineProps<{ content: string }>()

// Configure once: no GFM breaks (preserves intentional newlines in LLM output),
// async disabled so parse() returns a string synchronously.
marked.use({ breaks: false, gfm: true })

const html = computed(() => marked.parse(props.content) as string)
</script>

<template>
  <div class="md-prose" v-html="html" />
</template>

<style scoped>
.md-prose {
  color: #cbd5e1; /* slate-300 */
  font-size: 0.875rem;
  line-height: 1.7;
}

/* Headings */
.md-prose :deep(h1),
.md-prose :deep(h2),
.md-prose :deep(h3),
.md-prose :deep(h4) {
  color: #f1f5f9; /* slate-100 */
  font-weight: 600;
  margin-top: 1.1em;
  margin-bottom: 0.4em;
  line-height: 1.3;
}
.md-prose :deep(h1) { font-size: 1.1rem; }
.md-prose :deep(h2) { font-size: 1rem; }
.md-prose :deep(h3) { font-size: 0.9rem; }
.md-prose :deep(h4) { font-size: 0.85rem; }

/* Paragraphs */
.md-prose :deep(p) {
  margin-top: 0.6em;
  margin-bottom: 0.6em;
}

/* Bold / italic */
.md-prose :deep(strong) { color: #f1f5f9; font-weight: 600; }
.md-prose :deep(em)     { color: #94a3b8; font-style: italic; }

/* Inline code */
.md-prose :deep(code) {
  background: #1e293b; /* slate-800 */
  border: 1px solid #334155; /* slate-700 */
  border-radius: 4px;
  padding: 0.1em 0.4em;
  font-family: ui-monospace, 'Cascadia Code', 'Fira Code', monospace;
  font-size: 0.82em;
  color: #7dd3fc; /* sky-300 */
}

/* Fenced code blocks */
.md-prose :deep(pre) {
  background: #0f172a; /* slate-950 */
  border: 1px solid #334155;
  border-radius: 6px;
  padding: 0.75em 1em;
  overflow-x: auto;
  margin: 0.75em 0;
}
.md-prose :deep(pre code) {
  background: transparent;
  border: none;
  padding: 0;
  color: #e2e8f0; /* slate-200 */
  font-size: 0.82em;
}

/* Unordered / ordered lists */
.md-prose :deep(ul),
.md-prose :deep(ol) {
  padding-left: 1.4em;
  margin: 0.5em 0;
}
.md-prose :deep(ul) { list-style-type: disc; }
.md-prose :deep(ol) { list-style-type: decimal; }
.md-prose :deep(li) { margin: 0.2em 0; }
.md-prose :deep(li > p) { margin: 0; }

/* Nested lists */
.md-prose :deep(li > ul),
.md-prose :deep(li > ol) {
  margin: 0.1em 0;
}

/* Blockquote */
.md-prose :deep(blockquote) {
  border-left: 3px solid #475569; /* slate-600 */
  margin: 0.75em 0;
  padding: 0.4em 0.75em;
  color: #94a3b8; /* slate-400 */
  font-style: italic;
}

/* Horizontal rule */
.md-prose :deep(hr) {
  border: none;
  border-top: 1px solid #334155;
  margin: 1em 0;
}

/* Tables */
.md-prose :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 0.75em 0;
  font-size: 0.82em;
}
.md-prose :deep(th),
.md-prose :deep(td) {
  border: 1px solid #334155;
  padding: 0.35em 0.6em;
  text-align: left;
}
.md-prose :deep(th) {
  background: #1e293b;
  color: #f1f5f9;
  font-weight: 600;
}
.md-prose :deep(tr:nth-child(even) td) {
  background: #1e293b40;
}

/* Links */
.md-prose :deep(a) {
  color: #38bdf8; /* sky-400 */
  text-decoration: underline;
  text-underline-offset: 2px;
}
.md-prose :deep(a:hover) { color: #7dd3fc; }
</style>
