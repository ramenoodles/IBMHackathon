<script setup lang="ts">
/**
 * Primary call-to-action button with OnBober styling variants.
 */
import { computed } from 'vue'

/** Visual style variant for the button. */
export type ButtonVariant = 'primary' | 'ghost' | 'outline'

const props = withDefaults(
  defineProps<{
    /** Button style variant. */
    variant?: ButtonVariant
    /** Disables interaction and dims the button. */
    disabled?: boolean
    /** Renders as a full-width block element. */
    block?: boolean
    /** HTML button type attribute. */
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    variant: 'primary',
    disabled: false,
    block: false,
    type: 'button',
  },
)

const classes = computed(() => {
  const base =
    'inline-flex items-center justify-center rounded-lg px-5 py-2.5 text-sm font-semibold transition focus:outline-none focus:ring-2 focus:ring-onbober-primary/50 disabled:cursor-not-allowed disabled:opacity-50'
  const variants: Record<ButtonVariant, string> = {
    primary: 'bg-onbober-primary text-white hover:bg-[#e62e5c]',
    ghost: 'bg-transparent text-slate-300 hover:bg-slate-800',
    outline: 'border border-slate-600 text-slate-200 hover:border-onbober-primary hover:text-white',
  }
  return [base, variants[props.variant], props.block ? 'w-full' : ''].join(' ')
})
</script>

<template>
  <button :type="type" :class="classes" :disabled="disabled">
    <slot />
  </button>
</template>
