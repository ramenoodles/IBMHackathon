/** Wall-clock duration for the beaver log-chewing animation (0 → 100%). */
export const BEAVER_LOADER_DURATION_MS = 7500

/** Brief pause on 100% before leaving the setup screen. */
export const BEAVER_LOADER_HOLD_MS = 500

export function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
