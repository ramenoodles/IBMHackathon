import { ref } from 'vue'
import { api, type Workspace } from '@/api'
import { DEMO_REPO_URL } from '@/constants/demoRepo'

/** Source type for workspace onboarding step. */
export type WorkspaceSource = 'demo' | 'local' | 'github' | 'zip'

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
  async function setupLocal(path: string): Promise<Workspace> {
    return setup({ source: 'local', path })
  }

  /**
   * Clone a GitHub repository into a managed workspace directory.
   * @param url - GitHub URL or owner/repo shorthand.
   * @returns Resolved workspace path after clone completes.
   */
  async function setupGitHub(url: string): Promise<Workspace> {
    return setup({ source: 'github', url })
  }

  /** Clone the bundled IBM Bob demo repository. */
  async function setupDemo(): Promise<Workspace> {
    return setupGitHub(DEMO_REPO_URL)
  }

  /**
   * Upload and extract a zip archive as the workspace.
   * @param file - Zip file selected by the user.
   * @returns Resolved workspace path after extraction.
   */
  async function setupZip(file: File): Promise<Workspace> {
    loading.value = true
    error.value = null
    try {
      return await api.uploadWorkspace(file)
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
  }): Promise<Workspace> {
    loading.value = true
    error.value = null
    try {
      return await api.createWorkspace(body)
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
    setupDemo,
    setupZip,
  }
}
