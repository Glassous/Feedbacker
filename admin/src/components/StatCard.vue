<script setup lang="ts">
import type { Component } from 'vue'

type Tone = 'default' | 'amber' | 'green'

const props = defineProps<{
  label: string
  value: number
  icon: Component
  tone?: Tone
}>()

const tones: Record<Tone, string> = {
  default: 'text-mist-500',
  amber: 'text-signal-amber/80',
  green: 'text-signal-green/80',
}
</script>

<template>
  <article class="rounded-xl border border-black/8 bg-ink-900 p-3.5 sm:p-5">
    <div class="flex items-start justify-between gap-2">
      <div class="min-w-0">
        <p class="text-[11px] text-mist-500">{{ label }}</p>
        <p class="mt-1.5 text-xl font-semibold tabular-nums text-mist-100 sm:mt-2 sm:text-2xl">
          {{ value.toLocaleString('zh-CN') }}
        </p>
      </div>
      <component :is="icon" class="size-4 shrink-0" :class="tones[props.tone ?? 'default']" />
    </div>
  </article>
</template>
