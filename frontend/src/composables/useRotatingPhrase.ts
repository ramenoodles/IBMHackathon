import { computed, onUnmounted, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

/**
 * Cycles through a list of phrases on an interval while active.
 */
export function useRotatingPhrase(
  active: MaybeRefOrGetter<boolean>,
  phrases: MaybeRefOrGetter<readonly string[]>,
  intervalMs = 2500,
) {
  const phraseIndex = ref(0)
  const currentPhrase = computed(() => {
    const list = toValue(phrases)
    return list[phraseIndex.value] ?? ''
  })

  let phraseTimer: ReturnType<typeof setInterval> | null = null

  function clearTimer(): void {
    if (phraseTimer !== null) {
      clearInterval(phraseTimer)
      phraseTimer = null
    }
  }

  watch(
    () => toValue(active),
    (isActive) => {
      clearTimer()
      const list = toValue(phrases)
      if (!isActive || list.length === 0) return

      phraseIndex.value = Math.floor(Math.random() * list.length)
      phraseTimer = setInterval(() => {
        const currentList = toValue(phrases)
        if (currentList.length === 0) return
        phraseIndex.value = (phraseIndex.value + 1) % currentList.length
      }, intervalMs)
    },
    { immediate: true },
  )

  onUnmounted(clearTimer)

  return { phraseIndex, currentPhrase }
}
