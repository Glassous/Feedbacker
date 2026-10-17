<script setup lang="ts">
import { ChevronRight, FolderGit2 } from 'lucide-vue-next'
import type { Feedback } from '../lib/api'
import { excerpt, formatDateTime, relativeTime } from '../lib/format'
import StatusPill from './StatusPill.vue'

defineProps<{ items: Feedback[]; loading?: boolean }>()

const emit = defineEmits<{ select: [feedback: Feedback] }>()
</script>

<template>
  <div class="hidden overflow-x-auto md:block">
    <table class="w-full min-w-[860px] border-collapse text-left">
      <thead>
        <tr class="border-b border-black/6 text-[10px] font-medium uppercase tracking-[0.08em] text-mist-500">
          <th class="w-[100px] px-4 py-2.5 font-medium">状态</th>
          <th class="px-4 py-2.5 font-medium">内容</th>
          <th class="w-[130px] px-4 py-2.5 font-medium">来源</th>
          <th class="w-[160px] px-4 py-2.5 font-medium">联系方式</th>
          <th class="w-[120px] px-4 py-2.5 font-medium">时间</th>
          <th class="w-[36px]" />
        </tr>
      </thead>

      <tbody v-if="loading">
        <tr v-for="n in 6" :key="n" class="border-b border-black/5 last:border-0">
          <td class="px-4 py-3"><div class="skeleton h-5 w-14 rounded" /></td>
          <td class="px-4 py-3">
            <div class="skeleton h-3 w-1/3" />
            <div class="skeleton mt-2 h-2.5 w-2/3" />
          </td>
          <td class="px-4 py-3"><div class="skeleton h-4 w-16" /></td>
          <td class="px-4 py-3"><div class="skeleton h-2.5 w-24" /></td>
          <td class="px-4 py-3"><div class="skeleton h-2.5 w-14" /></td>
          <td />
        </tr>
      </tbody>

      <tbody v-else>
        <tr
          v-for="item in items"
          :key="item.id"
          class="group cursor-pointer border-b border-black/5 transition-colors duration-150 last:border-0 hover:bg-black/[0.025]"
          @click="emit('select', item)"
        >
          <td class="relative px-4 py-3">
            <span
              v-if="item.status === 'open'"
              class="absolute inset-y-3 left-0 w-[2px] rounded-r bg-black/25"
            />
            <StatusPill :status="item.status" />
          </td>

          <td class="max-w-0 px-4 py-3">
            <p
              class="truncate text-[13px]"
              :class="item.status === 'open' ? 'text-mist-100' : 'text-mist-400'"
            >
              {{ item.title || '（无标题）' }}
            </p>
            <p class="mt-0.5 truncate text-[11px] text-mist-500">
              {{ excerpt(item.content, 90) || '（无正文）' }}
            </p>
          </td>

          <td class="px-4 py-3">
            <span
              v-if="item.project"
              class="inline-flex max-w-full items-center gap-1 rounded border border-black/8 bg-black/[0.03] px-1.5 py-0.5 text-[11px] text-mist-400"
            >
              <FolderGit2 class="size-3 shrink-0" />
              <span class="truncate">{{ item.project }}</span>
            </span>
            <span v-else class="text-[11px] text-mist-500">—</span>
          </td>

          <td class="px-4 py-3">
            <p v-if="item.contact" class="truncate text-[11px] text-mist-400">{{ item.contact }}</p>
            <span v-else class="text-[11px] text-mist-500">—</span>
          </td>

          <td class="px-4 py-3">
            <p class="text-[11px] text-mist-400">{{ formatDateTime(item.createdAt) }}</p>
            <p class="mt-0.5 text-[10px] text-mist-500">{{ relativeTime(item.createdAt) }}</p>
          </td>

          <td class="pr-3">
            <ChevronRight
              class="size-3.5 text-mist-500 opacity-0 transition-all duration-200 group-hover:translate-x-0.5 group-hover:opacity-100"
            />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
