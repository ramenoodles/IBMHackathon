import { createHighlighter, type Highlighter } from 'shiki'

let highlighterPromise: Promise<Highlighter> | null = null

/**
 * Lazily initialize the Shiki highlighter with the github-dark theme.
 * @returns Shared highlighter instance.
 */
async function getHighlighter(): Promise<Highlighter> {
  if (!highlighterPromise) {
    highlighterPromise = createHighlighter({
      themes: ['github-dark'],
      langs: ['c', 'cpp', 'go', 'rust', 'python', 'bash', 'text'],
    })
  }
  return highlighterPromise
}

/**
 * Highlight source code with Shiki and return HTML markup.
 * @param source - Raw source code text.
 * @param lang - Language identifier for Shiki (defaults to text).
 * @returns HTML string with syntax highlighting spans.
 */
export async function highlightCode(source: string, lang = 'text'): Promise<string> {
  const highlighter = await getHighlighter()
  const loadedLangs = highlighter.getLoadedLanguages()
  const resolvedLang = loadedLangs.includes(lang as never) ? lang : 'text'
  return highlighter.codeToHtml(source, {
    lang: resolvedLang,
    theme: 'github-dark',
  })
}

/**
 * Map file extensions to Shiki language identifiers.
 * @param filePath - Path or filename to inspect.
 * @returns Shiki language id.
 */
export function languageFromPath(filePath: string): string {
  const ext = filePath.split('.').pop()?.toLowerCase() ?? ''
  const map: Record<string, string> = {
    c: 'c',
    h: 'c',
    cpp: 'cpp',
    hpp: 'cpp',
    cc: 'cpp',
    go: 'go',
    rs: 'rust',
    py: 'python',
    sh: 'bash',
  }
  return map[ext] ?? 'text'
}
