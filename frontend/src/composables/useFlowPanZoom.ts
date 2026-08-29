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

    const instance = Panzoom(content, {
      maxScale: 3,
      minScale: 0.15,
      step: 0.12,
      excludeClass: 'node',
      contain: 'outside',
      canvas: true,
    })

    panzoom.value = instance
    viewport.addEventListener('wheel', instance.zoomWithWheel)
  }

  function unbind(): void {
    const viewport = viewportRef.value
    const instance = panzoom.value
    if (viewport && instance) {
      viewport.removeEventListener('wheel', instance.zoomWithWheel)
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

  function fitToView(): void {
    const viewport = viewportRef.value
    const content = contentRef.value
    const instance = panzoom.value
    if (!viewport || !content || !instance) return

    const svg = content.querySelector('svg')
    if (!svg) {
      instance.reset()
      return
    }

    instance.reset({ animate: false })
    const vp = viewport.getBoundingClientRect()
    const box = svg.getBoundingClientRect()
    if (box.width === 0 || box.height === 0) return

    const padding = 32
    const scaleX = (vp.width - padding) / box.width
    const scaleY = (vp.height - padding) / box.height
    const scale = Math.min(scaleX, scaleY, 1)

    instance.zoom(scale, { animate: false })
    const scaledBox = svg.getBoundingClientRect()
    const dx = vp.left + vp.width / 2 - (scaledBox.left + scaledBox.width / 2)
    const dy = vp.top + vp.height / 2 - (scaledBox.top + scaledBox.height / 2)
    instance.pan(dx, dy, { relative: true })
  }

  onUnmounted(unbind)

  return { bind, unbind, zoomIn, zoomOut, reset, fitToView }
}
