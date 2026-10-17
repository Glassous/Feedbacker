<script setup lang="ts">
import { ChevronRight, FolderGit2 } from 'lucide-vue-next'
import type { Feedback } from '../lib/api'
import { excerpt, formatDateTime } from '../lib/format'
import StatusPill from './StatusPill.vue'

defineProps<{ item: Feedback }>()

const emit = defineEmits<{ select: [feedback: Feedback] }>()
</script>

<template>
  <button
    type="button"
    class="group flex w-full cursor-pointer flex-col gap-2 px-4 py-3.5 text-left transition-colors duration-150 hover:bg-black/[0.025]"
    @click="emit('select', item)"
  >
    <div class="flex items-center justify-between gap-3">
      <StatusPill :status="item.status" />
      <span class="text-[10px] text-mist-500">{{ formatDateTime(item.createdAt) }}</span>
    </div>

    <div class="flex items-start justify-between gap-2">
      <div class="min-w-0 flex-1">
        <p class="truncate text-[13px] text-mist-100">{{ item.title || '（无标题）' }}</p>
        <p class="mt-0.5 line-clamp-2 text-[11px] leading-5 text-mist-500">
          {{ excerpt(item.content, 80) || '（无正文）' }}
        </p>
      </div>
      <ChevronRight class="mt-1 size-3.5 shrink-0 text-mist-500 transition-transform group-hover:translate-x-0.5" />
    </div>

    <span v-if="item.project" class="inline-flex items-center gap-1 text-[10px] text-mist-500">
      <FolderGit2 class="size-3" />
      {{ item.project }}
    </span>
  </button>
</template>
