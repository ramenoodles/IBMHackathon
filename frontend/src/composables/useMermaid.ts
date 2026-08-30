import { ref, watch, type Ref } from 'vue'
import mermaid from 'mermaid'

let initialized = false

/**
 * Initialize Mermaid with dark theme settings (idempotent).
 */
function ensureMermaidInit(): void {
  if (initialized) return
  mermaid.initialize({
    startOnLoad: false,
    theme: 'dark',
    securityLevel: 'loose',
    flowchart: {
      useMaxWidth: false,
      htmlLabels: true,
    },
    themeVariables: {
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, 'Courier New', monospace",
      fontSize: '15px',
    },
  })
  initialized = true
}

/**
 * Extract the first complete mermaid fenced code block from markdown text.
 * @param markdown - LLM response text that may contain ```mermaid blocks.
 * @returns Mermaid source or null when no complete block is found.
 */
export function extractMermaidBlock(markdown: string): string | null {
  const match = markdown.match(/```mermaid\s*([\s\S]*?)```/)
  return match?.[1]?.trim() ?? null
}

/**
 * Render a Mermaid diagram into the given DOM element.
 * @param container - Target element to mount the SVG.
 * @param code - Mermaid diagram source.
 */
export async function renderMermaid(container: HTMLElement, code: string): Promise<void> {
  ensureMermaidInit()
  const id = `mermaid-${Date.now()}`
  const { svg } = await mermaid.render(id, code)
  container.innerHTML = svg
}

/**
 * Compile Mermaid source to an SVG string without writing to the DOM.
 * Allows the caller to guard against stale renders before committing.
 */
export async function compileMermaidSvg(code: string): Promise<string> {
  ensureMermaidInit()
  const id = `mermaid-${Date.now()}`
  const { svg } = await mermaid.render(id, code)
  return svg
}

/**
 * Composable that watches streamed markdown and renders Mermaid when a block completes.
 * @param markdown - Reactive ref of the full LLM response text.
 */
export function useMermaid(markdown: Ref<string>) {
  const mermaidCode = ref<string | null>(null)
  const renderError = ref<string | null>(null)
  const containerRef = ref<HTMLElement | null>(null)

  watch(
    markdown,
    async (text) => {
      const block = extractMermaidBlock(text)
      mermaidCode.value = block
      if (!block || !containerRef.value) return

      try {
        renderError.value = null
        await renderMermaid(containerRef.value, block)
      } catch (err) {
        renderError.value = err instanceof Error ? err.message : 'Failed to render diagram'
      }
    },
    { flush: 'post' },
  )

  /**
   * Bind the DOM container used for Mermaid rendering.
   * @param el - HTMLElement or null when unmounted.
   */
  function setContainer(el: HTMLElement | null): void {
    containerRef.value = el
    const block = extractMermaidBlock(markdown.value)
    if (el && block) {
      renderMermaid(el, block).catch((err) => {
        renderError.value = err instanceof Error ? err.message : 'Failed to render diagram'
      })
    }
  }

  return {
    mermaidCode,
    renderError,
    setContainer,
  }
}
