import { ref } from 'vue'

/** Source type for workspace onboarding step. */
export type WorkspaceSource = 'local' | 'github' | 'zip'

/**
 * Composable for registering a workspace via local path, GitHub URL, or zip upload.
 */
export function useWorkspaceSetup() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * Register a workspace from a local directory path.
   * @param path - Absolute path on the user's machine.
   * @returns Resolved workspace path usable by the scanner.
   */
  async function setupLocal(path: string): Promise<string> {
    return setup({ source: 'local', path })
  }

  /**
   * Clone a GitHub repository into a managed workspace directory.
   * @param url - GitHub URL or owner/repo shorthand.
   * @returns Resolved workspace path after clone completes.
   */
  async function setupGitHub(url: string): Promise<string> {
    return setup({ source: 'github', url })
  }

  /**
   * Upload and extract a zip archive as the workspace.
   * @param file - Zip file selected by the user.
   * @returns Resolved workspace path after extraction.
   */
  async function setupZip(file: File): Promise<string> {
    loading.value = true
    error.value = null
    try {
      const form = new FormData()
      form.append('file', file)
      const res = await fetch('/api/workspace/upload', { method: 'POST', body: form })
      if (!res.ok) {
        const msg = await res.text()
        throw new Error(msg || `Upload failed (${res.status})`)
      }
      const data = (await res.json()) as { workspacePath: string }
      return data.workspacePath
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Upload failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Call the backend workspace setup endpoint for local or GitHub sources.
   */
  async function setup(body: {
    source: 'local' | 'github'
    path?: string
    url?: string
  }): Promise<string> {
    loading.value = true
    error.value = null
    try {
      const res = await fetch('/api/workspace/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        const msg = await res.text()
        throw new Error(msg || `Setup failed (${res.status})`)
      }
      const data = (await res.json()) as { workspacePath: string }
      return data.workspacePath
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Setup failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    setupLocal,
    setupGitHub,
    setupZip,
  }
}
