import { computed, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'

export const BEAVER_LOG_START_X = 220
export const BEAVER_LOG_END_X = 980
export const BEAVER_LOG_TOTAL_TRAVEL = BEAVER_LOG_END_X - BEAVER_LOG_START_X
const BEAVER_OFFSET_X = -50
const BEAVER_BASE_Y = 252
const MAX_SHAVINGS = 50
const CHIP_COUNT = 32
const chipColors = ['#faeed7', '#e6c49c', '#cca06f', '#8d562c', '#6a3d1c', '#ffe8cc']

export type BeaverLogAnimationMode = 'timed' | 'progress' | 'indeterminate'

interface Chip {
  el: SVGPolygonElement
  active: boolean
  x: number
  y: number
  vx: number
  vy: number
  rot: number
  vRot: number
  scale: number
  life: number
  maxLife: number
}

interface Shaving {
  el: SVGEllipseElement
  x: number
  y: number
  rot: number
  revealed: boolean
}

export interface UseBeaverLogAnimationOptions {
  mode: Ref<BeaverLogAnimationMode>
  active: Ref<boolean>
  externalProgress?: Ref<number>
  durationMs?: number
}

/**
 * Shared rAF loop for the beaver log-chewing SVG animation.
 */
export function useBeaverLogAnimation(options: UseBeaverLogAnimationOptions) {
  const progress = ref(0)
  const isComplete = computed(() => progress.value >= 100)

  const beaverRig = ref<SVGGElement | null>(null)
  const beaverBobbing = ref<SVGGElement | null>(null)
  const intactClipRect = ref<SVGRectElement | null>(null)
  const gnawedClipRect = ref<SVGRectElement | null>(null)
  const activeNotch = ref<SVGGElement | null>(null)
  const chipsLayer = ref<SVGGElement | null>(null)
  const shavingsGroup = ref<SVGGElement | null>(null)

  const chipPool: Chip[] = []
  const shavings: Shaving[] = []

  let mainLoopId: number | null = null
  let lastTime = performance.now()
  let spawnCounter = 0
  let animStartTime: number | null = null
  let finishing = false

  function initParticles(): void {
    if (!chipsLayer.value || !shavingsGroup.value) return
    if (chipPool.length > 0) return

    for (let i = 0; i < CHIP_COUNT; i++) {
      const p = document.createElementNS('http://www.w3.org/2000/svg', 'polygon')
      p.setAttribute('points', i % 2 === 0 ? '-4,-2 4,-4 2,3 -3,3' : '-3,-3 4,-1 1,4 -3,2')
      p.setAttribute('fill', chipColors[i % chipColors.length]!)
      p.setAttribute('opacity', '0')
      chipsLayer.value.appendChild(p)
      chipPool.push({
        el: p,
        active: false,
        x: 0,
        y: 0,
        vx: 0,
        vy: 0,
        rot: 0,
        vRot: 0,
        scale: 1,
        life: 0,
        maxLife: 1,
      })
    }

    for (let i = 0; i < MAX_SHAVINGS; i++) {
      const sh = document.createElementNS('http://www.w3.org/2000/svg', 'ellipse')
      sh.setAttribute('rx', (Math.random() * 3.5 + 2).toFixed(1))
      sh.setAttribute('ry', (Math.random() * 1.8 + 0.8).toFixed(1))
      sh.setAttribute('fill', chipColors[Math.floor(Math.random() * chipColors.length)]!)
      sh.setAttribute('opacity', '0')
      shavingsGroup.value.appendChild(sh)
      shavings.push({
        el: sh,
        x: 0,
        y: 416 + Math.random() * 8,
        rot: Math.random() * 360,
        revealed: false,
      })
    }
  }

  function spawnWoodChip(originX: number, originY: number): void {
    const chip = chipPool.find((c) => !c.active)
    if (!chip) return
    chip.active = true
    chip.x = originX + (Math.random() * 8 - 4)
    chip.y = originY + (Math.random() * 14 - 7)
    chip.vx = -(Math.random() * 4 + 2)
    chip.vy = Math.random() * 5.5 - 3.2
    chip.rot = Math.random() * 360
    chip.vRot = Math.random() * 24 - 12
    chip.scale = Math.random() * 0.6 + 0.65
    chip.life = 0
    chip.maxLife = Math.random() * 0.4 + 0.35
    chip.el.setAttribute('opacity', '1')
  }

  function resetShavings(): void {
    for (const sh of shavings) {
      sh.revealed = false
      sh.el.setAttribute('opacity', '0')
    }
  }

  function resetChips(): void {
    chipPool.forEach((chip) => {
      chip.active = false
      chip.el.setAttribute('opacity', '0')
    })
  }

  function resetAnimation(): void {
    progress.value = 0
    animStartTime = null
    finishing = false
    spawnCounter = 0
    resetShavings()
    resetChips()
  }

  function advanceProgress(dt: number, now: number): void {
    const mode = options.mode.value

    if (mode === 'progress' && options.externalProgress) {
      progress.value = Math.min(100, Math.max(0, options.externalProgress.value))
      return
    }

    if (mode === 'timed') {
      if (options.active.value && animStartTime === null) {
        animStartTime = now
      }
      if (animStartTime !== null && (options.active.value || progress.value < 100)) {
        const elapsed = now - animStartTime
        const duration = options.durationMs ?? 7500
        progress.value = Math.min(100, (elapsed / duration) * 100)
      }
      return
    }

    // indeterminate
    if (options.active.value && !finishing) {
      const target = 90
      const speed = 8 + Math.sin(progress.value * 0.08) * 2
      if (progress.value < target) {
        progress.value += speed * dt
        if (progress.value > target) progress.value = target
      }
    } else if (finishing || !options.active.value) {
      progress.value += 40 * dt
      if (progress.value > 100) progress.value = 100
    }
  }

  function update(now: number): void {
    const dt = Math.min((now - lastTime) / 1000, 0.1)
    lastTime = now

    advanceProgress(dt, now)

    const cutX = BEAVER_LOG_START_X + (progress.value / 100) * BEAVER_LOG_TOTAL_TRAVEL
    const chewing = progress.value < 100

    if (beaverRig.value && beaverBobbing.value) {
      const beaverX = cutX + BEAVER_OFFSET_X
      const chewBobY = chewing ? Math.sin(now * 0.038) * 3.2 : 0
      const chewBobRot = chewing ? Math.cos(now * 0.038) * 2.0 : 0
      const chewScaleX = chewing ? 1 + Math.sin(now * 0.076) * 0.02 : 1
      beaverRig.value.setAttribute('transform', `translate(${beaverX}, ${BEAVER_BASE_Y})`)
      beaverBobbing.value.setAttribute(
        'transform',
        `translate(0, ${chewBobY}) rotate(${chewBobRot} 65 75) scale(${chewScaleX} 1)`,
      )
    }

    if (intactClipRect.value) {
      intactClipRect.value.setAttribute('x', String(cutX))
      intactClipRect.value.setAttribute('width', String(Math.max(0, BEAVER_LOG_END_X - cutX + 10)))
    }
    if (gnawedClipRect.value) {
      gnawedClipRect.value.setAttribute('x', String(BEAVER_LOG_START_X - 5))
      gnawedClipRect.value.setAttribute('width', String(Math.max(0, cutX - BEAVER_LOG_START_X + 5)))
    }
    if (activeNotch.value) {
      activeNotch.value.setAttribute('transform', `translate(${cutX}, 380)`)
      activeNotch.value.setAttribute('opacity', chewing ? '1' : '0')
    }

    if (chewing) {
      spawnCounter += dt
      if (spawnCounter > 0.045) {
        spawnCounter = 0
        spawnWoodChip(cutX - 10, 380)
        spawnWoodChip(cutX - 4, 376)
      }
    }

    for (const chip of chipPool) {
      if (!chip.active) continue
      chip.life += dt
      if (chip.life >= chip.maxLife) {
        chip.active = false
        chip.el.setAttribute('opacity', '0')
        continue
      }
      chip.vy += 22 * dt
      chip.x += chip.vx
      chip.y += chip.vy
      chip.rot += chip.vRot
      const alpha = 1 - chip.life / chip.maxLife
      chip.el.setAttribute(
        'transform',
        `translate(${chip.x.toFixed(1)}, ${chip.y.toFixed(1)}) rotate(${chip.rot.toFixed(1)}) scale(${chip.scale})`,
      )
      chip.el.setAttribute('opacity', Math.max(0, alpha).toFixed(2))
    }

    const revealCount = Math.floor((progress.value / 100) * MAX_SHAVINGS)
    for (let i = 0; i < MAX_SHAVINGS; i++) {
      const sh = shavings[i]!
      if (i < revealCount && !sh.revealed) {
        sh.revealed = true
        sh.x = BEAVER_LOG_START_X + (i / MAX_SHAVINGS) * BEAVER_LOG_TOTAL_TRAVEL + (Math.random() * 20 - 10)
        sh.el.setAttribute('cx', sh.x.toFixed(1))
        sh.el.setAttribute('cy', sh.y.toFixed(1))
        sh.el.setAttribute('transform', `rotate(${sh.rot.toFixed(1)} ${sh.x.toFixed(1)} ${sh.y.toFixed(1)})`)
        sh.el.setAttribute('opacity', '0.75')
      }
    }

    mainLoopId = requestAnimationFrame(update)
  }

  function startLoop(): void {
    if (mainLoopId !== null) return
    lastTime = performance.now()
    mainLoopId = requestAnimationFrame(update)
  }

  function stopLoop(): void {
    if (mainLoopId !== null) {
      cancelAnimationFrame(mainLoopId)
      mainLoopId = null
    }
  }

  watch(
    () => options.active.value,
    (isActive) => {
      if (options.mode.value === 'indeterminate' && !isActive && progress.value < 100) {
        finishing = true
      }
      if (isActive) {
        if (options.mode.value !== 'progress') {
          resetAnimation()
          animStartTime = performance.now()
        }
        startLoop()
      }
    },
    { immediate: true },
  )

  watch(
    () => options.externalProgress?.value,
    () => {
      if (options.mode.value === 'progress' && options.active.value) {
        startLoop()
      }
    },
  )

  onMounted(() => {
    initParticles()
    startLoop()
  })

  onUnmounted(() => {
    stopLoop()
  })

  return {
    progress,
    isComplete,
    beaverRig,
    beaverBobbing,
    intactClipRect,
    gnawedClipRect,
    activeNotch,
    chipsLayer,
    shavingsGroup,
    initParticles,
    resetAnimation,
  }
}
