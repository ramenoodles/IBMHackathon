import { ref } from 'vue'
import type { UserContext } from '@/store/userContext'

/**
 * Payload sent to the backend analyze endpoint.
 */
export interface AnalyzePayload {
  workspacePath: string
  filePath: string
  symbol: string
  userContext: UserContext
}

/**
 * Composable for consuming the backend SSE analyze stream.
 */
export function useLlmStream() {
  const text = ref('')
  const isStreaming = ref(false)
  const error = ref<string | null>(null)
  const isMock = ref(false)
  let abortController: AbortController | null = null

  /**
   * Parse a single SSE message chunk into event type and data payload.
   * @param chunk - Raw SSE text segment.
   */
  function parseSSEEvent(chunk: string): { event: string; data: string } | null {
    const lines = chunk.split('\n')
    let event = 'message'
    let data = ''
    for (const line of lines) {
      if (line.startsWith('event:')) event = line.slice(6).trim()
      if (line.startsWith('data:')) data += line.slice(5).trim()
    }
    if (!data) return null
    return { event, data }
  }

  /**
   * Start streaming an analysis from the backend.
   * @param payload - Analyze request including user context and symbol.
   */
  async function analyze(payload: AnalyzePayload): Promise<void> {
    abortController?.abort()
    abortController = new AbortController()

    text.value = ''
    error.value = null
    isMock.value = false
    isStreaming.value = true

    try {
      const response = await fetch('/api/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: abortController.signal,
      })

      if (!response.ok) {
        throw new Error(`Analyze request failed (${response.status})`)
      }

      const reader = response.body?.getReader()
      if (!reader) throw new Error('Response body is not readable')

      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const parts = buffer.split('\n\n')
        buffer = parts.pop() ?? ''

        for (const part of parts) {
          const parsed = parseSSEEvent(part)
          if (!parsed) continue

          if (parsed.event === 'token') {
            try {
              const json = JSON.parse(parsed.data) as { content?: string; mock?: boolean }
              if (json.content) text.value += json.content
              if (json.mock) isMock.value = true
            } catch {
              text.value += parsed.data
            }
          } else if (parsed.event === 'error') {
            error.value = parsed.data
          }
        }
      }
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') return
      error.value = err instanceof Error ? err.message : 'Stream failed'
    } finally {
      isStreaming.value = false
    }
  }

  /** Cancel any in-flight analyze stream. */
  function cancel(): void {
    abortController?.abort()
    isStreaming.value = false
  }

  return {
    text,
    isStreaming,
    error,
    isMock,
    analyze,
    cancel,
  }
}
