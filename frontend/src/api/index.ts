import type {
  EnrichNodeInput,
  EnrichResult,
  FlowGraph,
  GraphExpandPayload,
  GraphRootPayload,
  NodeDetail,
} from '@/types/flowGraph'

export interface Workspace {
  id: string
  name: string
  source?: string
}

export interface FileResponse {
  path: string
  content: string
  language: string
}

export interface SymbolResponse<T> {
  symbols: T[]
  count: number
}

export interface TreeResponse<T> {
  entries: T[]
}

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

function workspacePath(id: string, resource: string): string {
  return `/api/workspaces/${encodeURIComponent(id)}/${resource}`
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(path, options)
  if (!response.ok) {
    let message = `Request failed (${response.status})`
    try {
      const body = (await response.json()) as { error?: { message?: string } }
      message = body.error?.message || message
    } catch {
      const body = await response.text()
      if (body) message = body
    }
    throw new ApiError(message, response.status)
  }
  return (await response.json()) as T
}

function jsonBody(body: unknown): RequestInit {
  return {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  }
}

export const api = {
  createWorkspace(body: { source: 'local' | 'github'; path?: string; url?: string }): Promise<Workspace> {
    return request<Workspace>('/api/workspaces', jsonBody(body))
  },

  uploadWorkspace(file: File): Promise<Workspace> {
    const body = new FormData()
    body.append('file', file)
    return request<Workspace>('/api/workspaces', { method: 'POST', body })
  },

  tree<T>(workspaceId: string, dir = ''): Promise<TreeResponse<T>> {
    const params = new URLSearchParams()
    if (dir) params.set('path', dir)
    const query = params.toString()
    return request<TreeResponse<T>>(`${workspacePath(workspaceId, 'tree')}${query ? `?${query}` : ''}`)
  },

  file(workspaceId: string, filePath: string): Promise<FileResponse> {
    const params = new URLSearchParams({ path: filePath })
    return request<FileResponse>(`${workspacePath(workspaceId, 'file')}?${params}`)
  },

  symbols<T>(workspaceId: string, filePath: string): Promise<SymbolResponse<T>> {
    const params = new URLSearchParams({ path: filePath })
    return request<SymbolResponse<T>>(`${workspacePath(workspaceId, 'symbols')}?${params}`)
  },

  graph(payload: GraphRootPayload): Promise<FlowGraph> {
    return request<FlowGraph>(workspacePath(payload.workspaceId, 'graphs'), jsonBody(payload))
  },

  expand(payload: GraphExpandPayload): Promise<FlowGraph> {
    return request<FlowGraph>(workspacePath(payload.workspaceId, 'graphs/expand'), jsonBody(payload))
  },

  explain(workspaceId: string, body: {
    name: string
    question: string
    language?: string
    file?: string
    line?: number
    code?: string
    kind?: string
    title?: string
    experience?: string
  }, signal?: AbortSignal): Promise<NodeDetail> {
    return request<NodeDetail>(workspacePath(workspaceId, 'explain'), { ...jsonBody(body), signal })
  },

  enrich(workspaceId: string, body: {
    workspaceId: string
    filePath: string
    symbol: string
    nodes: EnrichNodeInput[]
    userContext: GraphRootPayload['userContext']
  }): Promise<EnrichResult> {
    return request<EnrichResult>(workspacePath(workspaceId, 'graphs/enrich'), jsonBody(body))
  },

  async explainStream(workspaceId: string, body: {
    name: string
    question: string
    language?: string
    file?: string
    line?: number
    code?: string
    kind?: string
    title?: string
    experience?: string
  }, signal?: AbortSignal): Promise<Response> {
    const response = await fetch(workspacePath(workspaceId, 'explain'), { ...jsonBody(body), signal })
    if (!response.ok) {
      let message = `Request failed (${response.status})`
      try {
        const body_2 = (await response.json()) as { error?: { message?: string} }
        message = body_2.error?.message || message
      } catch {
      }
      throw new ApiError(message, response.status)
    }
    return response
  },
}
