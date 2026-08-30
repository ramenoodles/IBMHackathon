import { ref } from 'vue'

const STORAGE_KEY = 'onbober:workspace-bobbers'

function loadShowBobbers(): boolean {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw === null) return true
    return JSON.parse(raw) === true
  } catch {
    return true
  }
}

function persistShowBobbers(value: boolean): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(value))
}

const showBobbers = ref(loadShowBobbers())

/**
 * Workspace easter-egg bobber visibility (persisted in localStorage).
 */
export function useWorkspaceBobbers() {
  function setShowBobbers(value: boolean): void {
    showBobbers.value = value
    persistShowBobbers(value)
  }

  return { showBobbers, setShowBobbers }
}
