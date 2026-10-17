<script setup lang="ts">
import type { FeedbackStatus } from '../lib/api'

const props = defineProps<{ status: FeedbackStatus }>()
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-md border border-black/8 bg-black/[0.04] px-2 py-0.5 text-[11px] text-mist-300"
  >
    <span
      class="size-1.5 rounded-full"
      :class="props.status === 'open' ? 'bg-signal-amber' : 'bg-signal-green'"
    />
    {{ props.status === 'open' ? '未处理' : '已处理' }}
  </span>
</template>
