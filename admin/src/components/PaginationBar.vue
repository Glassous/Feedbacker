<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{ change: [page: number] }>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
</script>

<template>
  <div class="flex items-center justify-between border-t border-black/6 px-4 py-3">
    <p class="text-[11px] text-mist-500">{{ total }} 条</p>

    <div class="flex items-center gap-1">
      <button
        type="button"
        class="flex size-7 cursor-pointer items-center justify-center rounded-md text-mist-500 transition-colors hover:bg-black/[0.04] hover:text-mist-200 disabled:cursor-not-allowed disabled:opacity-30"
        :disabled="page <= 1"
        aria-label="上一页"
        @click="emit('change', page - 1)"
      >
        <ChevronLeft class="size-4" />
      </button>
      <span class="min-w-[52px] text-center font-mono text-[11px] text-mist-400">
        {{ page }} / {{ totalPages }}
      </span>
      <button
        type="button"
        class="flex size-7 cursor-pointer items-center justify-center rounded-md text-mist-500 transition-colors hover:bg-black/[0.04] hover:text-mist-200 disabled:cursor-not-allowed disabled:opacity-30"
        :disabled="page >= totalPages"
        aria-label="下一页"
        @click="emit('change', page + 1)"
      >
        <ChevronRight class="size-4" />
      </button>
    </div>
  </div>
</template>
