<script setup lang="ts">
import { computed } from 'vue'
import { Loader2 } from 'lucide-vue-next'

type Variant = 'primary' | 'ghost' | 'danger' | 'subtle' | 'link'
type Size = 'sm' | 'md' | 'icon'

const props = withDefaults(
  defineProps<{
    variant?: Variant
    size?: Size
    type?: 'button' | 'submit' | 'reset'
    disabled?: boolean
    loading?: boolean
    block?: boolean
  }>(),
  { variant: 'primary', size: 'md', type: 'button', disabled: false, loading: false, block: false },
)

const base =
  'relative inline-flex cursor-pointer items-center justify-center gap-2 rounded-lg font-medium transition-all duration-200 outline-none focus-visible:ring-2 focus-visible:ring-black/25 focus-visible:ring-offset-2 focus-visible:ring-offset-ink-950 disabled:cursor-not-allowed disabled:opacity-50 active:scale-[0.99]'

const variants: Record<Variant, string> = {
  primary: 'bg-mist-100 text-ink-950 hover:bg-mist-200',
  ghost:
    'border border-black/10 bg-black/[0.04] text-mist-200 hover:border-black/20 hover:bg-black/[0.08] hover:text-mist-100',
  danger: 'border border-signal-red/25 bg-signal-red/10 text-signal-red hover:bg-signal-red/15',
  subtle: 'bg-black/[0.08] text-mist-100 hover:bg-black/[0.12]',
  link: 'text-mist-400 hover:text-mist-100',
}

const sizes: Record<Size, string> = {
  sm: 'h-8 px-3 text-[12px]',
  md: 'h-10 px-4 text-sm',
  icon: 'size-9 p-0',
}

const classes = computed(() => [
  base,
  variants[props.variant],
  sizes[props.size],
  props.block ? 'w-full' : '',
])
</script>

<template>
  <button :type="type" :disabled="disabled || loading" :class="classes">
    <Loader2 v-if="loading" class="size-4 animate-spin" />
    <slot />
  </button>
</template>
