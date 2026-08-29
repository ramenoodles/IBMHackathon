/**
 * Maps file/folder names to VS Code Material Icon Theme icon URLs.
 */
import type { MaterialIcon } from 'vscode-material-icons'
import { getIconForDirectoryPath, getIconForFilePath, getIconUrlByName } from 'vscode-material-icons'

const ICONS_BASE_URL = '/icons/vscode-material-icons'

/** Icon URL for a file, based on its name/extension. */
export function getFileIconUrl(name: string): string {
  return getIconUrlByName(getIconForFilePath(name), ICONS_BASE_URL)
}

/** Icon URL for a folder, based on its name and expanded state. */
export function getFolderIconUrl(name: string, expanded: boolean): string {
  const iconName = getIconForDirectoryPath(name)
  const withState = expanded ? (`${iconName}-open` as MaterialIcon) : iconName
  return getIconUrlByName(withState, ICONS_BASE_URL)
}
