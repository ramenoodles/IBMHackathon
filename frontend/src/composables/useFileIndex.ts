import { fetchTreeEntries, type TreeEntry } from '@/composables/useFileTree'

/** Ranked file search result. Lower rank = better match. */
export interface FileSearchResult extends TreeEntry {
  rank: number
}

const indexCache = new Map<string, TreeEntry[]>()
const indexPromises = new Map<string, Promise<TreeEntry[]>>()
let activeWorkspaceId: string | null = null
let abortController: AbortController | null = null

const CONCURRENCY = 4

/**
 * Score a file entry against a query. Returns null when no match.
 * Lower score = better match.
 */
export function scoreFileMatch(entry: TreeEntry, query: string): number | null {
  const q = query.trim().toLowerCase()
  if (!q) return null

  const name = entry.name.toLowerCase()
  const path = entry.path.toLowerCase()

  if (name === q) return 0
  if (name.startsWith(q)) return 1
  if (name.includes(q)) return 2
  if (path.includes(q)) return 3
  return null
}

/**
 * Filter and rank files from a pre-built index.
 */
export function searchFiles(index: TreeEntry[], query: string, limit = 20): FileSearchResult[] {
  const q = query.trim()
  if (!q) return []

  const matches: FileSearchResult[] = []
  for (const entry of index) {
    if (entry.isDir) continue
    const rank = scoreFileMatch(entry, q)
    if (rank !== null) {
      matches.push({ ...entry, rank })
    }
  }

  matches.sort((a, b) => {
    if (a.rank !== b.rank) return a.rank - b.rank
    return a.path.localeCompare(b.path)
  })

  return matches.slice(0, limit)
}

/**
 * Walk the workspace tree and collect all file entries.
 * Results are cached per workspace until invalidated.
 */
export async function buildFileIndex(workspaceId: string): Promise<TreeEntry[]> {
  const cached = indexCache.get(workspaceId)
  if (cached) return cached

  const pending = indexPromises.get(workspaceId)
  if (pending) return pending

  if (abortController) {
    abortController.abort()
  }
  abortController = new AbortController()
  const signal = abortController.signal
  activeWorkspaceId = workspaceId

  const promise = walkTree(workspaceId, signal).then((files) => {
    if (!signal.aborted && activeWorkspaceId === workspaceId) {
      indexCache.set(workspaceId, files)
    }
    indexPromises.delete(workspaceId)
    return files
  })

  indexPromises.set(workspaceId, promise)
  return promise
}

/** Clear cached index for a workspace (e.g. on workspace change). */
export function invalidateFileIndex(workspaceId?: string): void {
  if (workspaceId) {
    indexCache.delete(workspaceId)
    indexPromises.delete(workspaceId)
    return
  }
  indexCache.clear()
  indexPromises.clear()
}

async function walkTree(workspaceId: string, signal: AbortSignal): Promise<TreeEntry[]> {
  const files: TreeEntry[] = []
  const queue: string[] = ['']

  while (queue.length > 0) {
    if (signal.aborted) return files

    const batch = queue.splice(0, CONCURRENCY)
    const results = await Promise.all(
      batch.map(async (dir) => {
        try {
          return await fetchTreeEntries(workspaceId, dir)
        } catch {
          return [] as TreeEntry[]
        }
      }),
    )

    for (let i = 0; i < batch.length; i++) {
      const entries = results[i] ?? []
      for (const entry of entries) {
        if (signal.aborted) return files
        if (entry.isDir) {
          queue.push(entry.path)
        } else {
          files.push(entry)
        }
      }
    }
  }

  return files
}
