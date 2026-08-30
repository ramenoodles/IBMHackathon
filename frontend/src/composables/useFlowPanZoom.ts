import Panzoom, { type PanzoomObject } from '@panzoom/panzoom'
import { onUnmounted, ref, type Ref } from 'vue'

/**
 * Pan/zoom viewport for large Mermaid flowcharts.
 */
export function useFlowPanZoom(viewportRef: Ref<HTMLElement | null>, contentRef: Ref<HTMLElement | null>) {
  const panzoom = ref<PanzoomObject | null>(null)
  const isFitted = ref(false)

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

    // Any wheel zoom counts as a user interaction — clear fitted state
    const onWheel = (e: WheelEvent) => {
      isFitted.value = false
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
    isFitted.value = false
  }

  function zoomIn(): void {
    isFitted.value = false
    panzoom.value?.zoomIn()
  }

  function zoomOut(): void {
    isFitted.value = false
    panzoom.value?.zoomOut()
  }

  function reset(): void {
    isFitted.value = false
    panzoom.value?.reset()
  }

  function fitToView(): void {
    if (isFitted.value) return

    const viewport = viewportRef.value
    const content = contentRef.value
    const instance = panzoom.value
    if (!viewport || !content || !instance) return

    const svg = content.querySelector('svg')
    if (!svg) {
      instance.reset()
      return
    }

    // Reset to the true Panzoom origin so getBoundingClientRect reflects 1× size
    instance.reset({ animate: false })

    const vp = viewport.getBoundingClientRect()
    const baseBox = svg.getBoundingClientRect()
    if (baseBox.width === 0 || baseBox.height === 0) return

    const padding = 40
    const scaleX = (vp.width - padding * 2) / baseBox.width
    const scaleY = (vp.height - padding * 2) / baseBox.height
    // No artificial upper cap — allow scaling up small charts to fill the viewport
    const scale = Math.min(scaleX, scaleY)

    // Apply the computed scale, then measure the new rendered position and centre
    instance.zoom(scale, { animate: false })
    const scaledBox = svg.getBoundingClientRect()
    const dx = vp.left + vp.width / 2 - (scaledBox.left + scaledBox.width / 2)
    const dy = vp.top + vp.height / 2 - (scaledBox.top + scaledBox.height / 2)
    instance.pan(dx, dy, { relative: true })

    isFitted.value = true
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

  return { bind, unbind, zoomIn, zoomOut, reset, fitToView, centerView, isFitted }
}
