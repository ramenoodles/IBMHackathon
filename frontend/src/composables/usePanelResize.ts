import { onUnmounted, type Ref } from 'vue'

export interface HorizontalResizeOptions {
  width: Ref<number>
  min: number
  max: number
  /** Dragging handle on the right edge grows width; on the left shrinks from the right. */
  side: 'left' | 'right'
  enabled?: Ref<boolean>
  onEnd?: () => void
}

/**
 * Pointer-driven horizontal panel resize.
 */
export function useHorizontalResize(opts: HorizontalResizeOptions) {
  let startX = 0
  let startWidth = 0
  let active = false

  function onPointerMove(e: PointerEvent): void {
    if (!active) return
    const delta = e.clientX - startX
    const next =
      opts.side === 'right' ? startWidth + delta : startWidth - delta
    opts.width.value = Math.min(opts.max, Math.max(opts.min, next))
  }

  function onPointerUp(): void {
    if (!active) return
    active = false
    window.removeEventListener('pointermove', onPointerMove)
    window.removeEventListener('pointerup', onPointerUp)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
    opts.onEnd?.()
  }

  function onPointerDown(e: PointerEvent): void {
    if (opts.enabled && !opts.enabled.value) return
    active = true
    startX = e.clientX
    startWidth = opts.width.value
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerup', onPointerUp)
    e.preventDefault()
  }

  onUnmounted(onPointerUp)

  return { onPointerDown }
}
