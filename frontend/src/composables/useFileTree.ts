/**
 * File tree entry returned by the backend /api/tree endpoint.
 */
export interface TreeEntry {
  name: string
  path: string
  isDir: boolean
}

/**
 * Fetch directory entries for a workspace subdirectory.
 * @param workspace - Absolute workspace root path.
 * @param dir - Relative subdirectory path (empty for root).
 * @returns List of files and folders with paths relative to workspace root.
 */
export async function fetchTreeEntries(workspace: string, dir = ''): Promise<TreeEntry[]> {
  const params = new URLSearchParams()
  if (dir) params.set('dir', dir)

  const res = await fetch(`/api/workspaces/${encodeURIComponent(workspace)}/tree?${params}`)
  if (!res.ok) {
    const msg = await res.text()
    throw new Error(msg || `Failed to load tree (${res.status})`)
  }

  const data = (await res.json()) as { entries: TreeEntry[] }
  return data.entries ?? []
}
