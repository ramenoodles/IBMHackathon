import { ref } from 'vue'
import type { FlowConfidence, NodeDetail } from '@/types/flowGraph'
import { api } from '@/api'

export interface NodeDetailParams {
  workspaceId: string
  nodeId: string
  symbol: string
  file?: string
  line?: number
  title?: string
  confidence?: FlowConfidence
  code?: string
  kind?: string
  summary?: string
  experience?: string
  language?: string
}

function cacheKey(params: NodeDetailParams): string {
  return `${params.file ?? ''}::${params.symbol}::${params.nodeId}::${params.experience ?? ''}`
}

function buildQuery(params: NodeDetailParams): URLSearchParams {
  const q = new URLSearchParams({ nodeId: params.nodeId, symbol: params.symbol })
  if (params.file) q.set('file', params.file)
  if (params.line) q.set('line', String(params.line))
  if (params.title) q.set('title', params.title)
  if (params.code) q.set('code', params.code)
  if (params.kind) q.set('kind', params.kind)
  if (params.summary) q.set('summary', params.summary)
  if (params.confidence) q.set('confidence', params.confidence)
  if (params.experience) q.set('experience', params.experience)
  if (params.language) q.set('language', params.language)
  return q
}

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
 * Fetches detailed explanation for a flow graph node with cache and streaming.
 */
export function useNodeDetail() {
  const detail = ref<NodeDetail | null>(null)
  const loading = ref(false)
  const streaming = ref(false)
  const error = ref<string | null>(null)
  const cache = new Map<string, NodeDetail>()
  let abortController: AbortController | null = null

  function getCached(params: NodeDetailParams): NodeDetail | undefined {
    return cache.get(cacheKey(params))
  }

  function setCached(params: NodeDetailParams, value: NodeDetail): void {
    cache.set(cacheKey(params), value)
  }

  function clearCache(): void {
    cache.clear()
  }

  function clear(): void {
    abortController?.abort()
    detail.value = null
    error.value = null
    loading.value = false
    streaming.value = false
  }

  async function loadDetail(params: NodeDetailParams, opts?: { stream?: boolean }): Promise<void> {
    const key = cacheKey(params)
    const cached = cache.get(key)
    if (cached) {
      detail.value = cached
      return
    }

    abortController?.abort()
    abortController = new AbortController()
    const signal = abortController.signal

    loading.value = true
    streaming.value = Boolean(opts?.stream)
    error.value = null
    detail.value = {
      id: params.nodeId,
      title: params.title ?? params.nodeId,
      summary: params.summary ?? '',
      explanation: '',
      confidence: params.confidence ?? 'verified',
      file: params.file,
      line: params.line,
    }

    try {
      const parsed = await api.explain(params.workspaceId, {
        name: params.symbol,
        question: '',
        language: params.language,
        file: params.file,
        line: params.line,
        code: params.code,
        kind: params.kind,
        title: params.title,
        experience: params.experience,
      }, signal)
      detail.value = { ...detail.value, ...parsed, id: params.nodeId, title: params.title ?? params.nodeId, summary: params.summary ?? '' }
      setCached(params, detail.value)
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') return
      error.value = err instanceof Error ? err.message : 'Failed to load detail'
      if (!detail.value?.explanation) detail.value = null
    } finally {
      loading.value = false
      streaming.value = false
    }
  }

  async function loadDetailStream(params: NodeDetailParams, signal: AbortSignal): Promise<void> {
    const q = buildQuery(params)
    const res = await api.explainStream(params.workspaceId, {
      name: params.symbol,
      question: '',
      language: params.language,
      file: params.file,
      line: params.line,
      code: params.code,
      kind: params.kind,
      title: params.title,
      experience: params.experience,
    }, signal)

    const reader = res.body?.getReader()
    if (!reader) throw new Error('Response body is not readable')

    const decoder = new TextDecoder()
    let buffer = ''
    let explanation = ''

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
            if (json.content) {
              explanation += json.content
              if (detail.value) {
                detail.value = { ...detail.value, explanation, mock: json.mock ?? detail.value.mock }
              }
            }
          } catch {
            explanation += parsed.data
            if (detail.value) detail.value = { ...detail.value, explanation }
          }
        } else if (parsed.event === 'meta') {
          try {
            const meta = JSON.parse(parsed.data) as NodeDetail
            detail.value = { ...detail.value!, ...meta, explanation }
          } catch {
            /* ignore malformed meta */
          }
        } else if (parsed.event === 'done') {
          try {
            const final = JSON.parse(parsed.data) as NodeDetail
            detail.value = {
              ...detail.value!,
              ...final,
              explanation: final.explanation || explanation.trim(),
            }
            setCached(params, detail.value)
          } catch {
            if (detail.value) {
              const finalDetail = { ...detail.value, explanation: explanation.trim() || detail.value.explanation }
              detail.value = finalDetail
              setCached(params, finalDetail)
            }
          }
        } else if (parsed.event === 'error') {
          throw new Error(parsed.data)
        }
      }
    }

    if (detail.value && !cache.has(cacheKey(params))) {
      const finalDetail = { ...detail.value, explanation: explanation.trim() || detail.value.explanation }
      detail.value = finalDetail
      setCached(params, finalDetail)
    }
  }

  return { detail, loading, streaming, error, loadDetail, getCached, clear, clearCache }
}
