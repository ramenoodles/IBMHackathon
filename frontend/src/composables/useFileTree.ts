import { api } from '@/api'

/**
 * File tree entry returned by the workspace tree endpoint.
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
  const data = await api.tree<TreeEntry>(workspace, dir)
  return data.entries ?? []
}
