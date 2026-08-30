import Panzoom, { type PanzoomObject } from '@panzoom/panzoom'
import { onUnmounted, ref, type Ref } from 'vue'

/**
 * Pan/zoom viewport for large Mermaid flowcharts.
 */
export function useFlowPanZoom(viewportRef: Ref<HTMLElement | null>, contentRef: Ref<HTMLElement | null>) {
  const panzoom = ref<PanzoomObject | null>(null)

  function bind(): void {
    unbind()
    const viewport = viewportRef.value
    const content = contentRef.value
    if (!viewport || !content) return

    // Clear any leftover transform from the previous Panzoom instance so the new
    // one starts at a true 1× baseline rather than inheriting the prior scale.
    content.style.transform = ''

    const instance = Panzoom(content, {
      minScale: 0.05,
      maxScale: 20,
      step: 0.12,
      excludeClass: 'node',
      canvas: true,
    })

    panzoom.value = instance

    const onWheel = (e: WheelEvent) => {
      instance.zoomWithWheel(e)
    }
    viewport.addEventListener('wheel', onWheel)
    // Store the handler so unbind can remove it
    ;(viewport as HTMLElement & { _panzoomWheel?: (e: WheelEvent) => void })._panzoomWheel = onWheel
  }

  function unbind(): void {
    const viewport = viewportRef.value
    const instance = panzoom.value
    if (viewport && instance) {
      const handler = (viewport as HTMLElement & { _panzoomWheel?: (e: WheelEvent) => void })._panzoomWheel
      if (handler) {
        viewport.removeEventListener('wheel', handler)
        delete (viewport as HTMLElement & { _panzoomWheel?: (e: WheelEvent) => void })._panzoomWheel
      }
    }
    panzoom.value?.destroy()
    panzoom.value = null
  }

  function zoomIn(): void {
    panzoom.value?.zoomIn()
  }

  function zoomOut(): void {
    panzoom.value?.zoomOut()
  }

  function reset(): void {
    panzoom.value?.reset()
  }

  function centerView(): void {
    const viewport = viewportRef.value
    const content = contentRef.value
    const instance = panzoom.value
    if (!viewport || !content || !instance) return

    const svg = content.querySelector('svg')
    if (!svg) return

    const vp = viewport.getBoundingClientRect()

    // Find the topmost node — Mermaid always assigns "flowchart-n0-*" to the first node
    const topNode = content.querySelector<SVGGElement>('[id^="flowchart-n0-"]')

    if (topNode) {
      const nodeBox = topNode.getBoundingClientRect()
      const topBuffer = 64
      // Centre horizontally on the top node, place it near the top of the viewport
      const dx = vp.left + vp.width / 2 - (nodeBox.left + nodeBox.width / 2)
      const dy = vp.top + topBuffer - nodeBox.top
      instance.pan(dx, dy, { relative: true })
    } else {
      // Fallback: centre the whole SVG
      const box = svg.getBoundingClientRect()
      if (box.width === 0 || box.height === 0) return
      const dx = vp.left + vp.width / 2 - (box.left + box.width / 2)
      const dy = vp.top + vp.height / 2 - (box.top + box.height / 2)
      instance.pan(dx, dy, { relative: true })
    }
  }

  onUnmounted(unbind)

  return { bind, unbind, zoomIn, zoomOut, reset, centerView }
}
