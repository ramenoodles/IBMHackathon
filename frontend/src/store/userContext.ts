import { readonly, ref } from 'vue'

/**
 * Experience level tiers used to tailor LLM explanations.
 */
export type ExperienceLevel = 'junior' | 'mid' | 'senior'

/**
 * Developer context collected during onboarding and sent to the backend.
 */
export interface UserContext {
  /** Primary programming language the developer is most comfortable with. */
  primaryLanguage: string
  /** Self-reported experience level. */
  experienceLevel: ExperienceLevel
  /** Absolute path to the local codebase workspace. */
  workspaceId: string
  workspaceName: string
}

const STORAGE_KEY = 'onbober:user-context'

const defaultContext: UserContext = {
  primaryLanguage: '',
  experienceLevel: 'junior',
  workspaceId: '',
  workspaceName: '',
}

/**
 * Reactive singleton holding the current user context.
 */
const state = ref<UserContext>(loadFromStorage())

/**
 * Read-only view of the user context for components.
 */
export const userContext = readonly(state)

/**
 * Load persisted user context from sessionStorage.
 * @returns Parsed user context or defaults when storage is empty.
 */
function loadFromStorage(): UserContext {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...defaultContext }
    return { ...defaultContext, ...JSON.parse(raw) }
  } catch {
    return { ...defaultContext }
  }
}

/**
 * Persist the current user context to sessionStorage.
 */
export function persistUserContext(): void {
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state.value))
}

/**
 * Merge partial updates into the user context and persist.
 * @param partial - Fields to update on the current context.
 */
export function updateUserContext(partial: Partial<UserContext>): void {
  state.value = { ...state.value, ...partial }
  persistUserContext()
}

/**
 * Determine whether onboarding has collected enough context for the workspace.
 * @returns True when language, experience, and workspace path are set.
 */
export function isUserContextComplete(): boolean {
  const ctx = state.value
  return Boolean(ctx.primaryLanguage && ctx.experienceLevel && ctx.workspaceId.trim())
}

/**
 * Reset user context to defaults and clear sessionStorage.
 */
export function clearUserContext(): void {
  state.value = { ...defaultContext }
  sessionStorage.removeItem(STORAGE_KEY)
}

const LANG_ALIASES: Record<string, string> = {
  'c/c++': 'cpp',
  'c++': 'cpp',
  'c#': 'csharp',
  'cs': 'csharp',
  'py': 'python',
  'js': 'javascript',
  'ts': 'typescript',
  'rs': 'rust',
}
const VALID_LANGS = new Set(['auto', 'c', 'cpp', 'csharp', 'go', 'java', 'javascript', 'python', 'rust', 'typescript'])

function normalizeSingle(lang: string): string {
  const key = lang.toLowerCase().trim()
  if (VALID_LANGS.has(key)) return key
  if (LANG_ALIASES[key]) return LANG_ALIASES[key]
  return 'auto'
}

/**
 * Map a stored language string to the best backend-safe value for the explain
 * endpoint. Handles comma-separated multi-select values: returns the first
 * non-auto language, or "auto" if all entries are unsupported.
 *
 * The backend accepts: auto, c, cpp, csharp, go, java, javascript, python, rust, typescript.
 */
export function normalizeLanguage(lang: string): string {
  const candidates = lang.split(',').map((s) => normalizeSingle(s.trim()))
  return candidates.find((v) => v !== 'auto') ?? 'auto'
}
