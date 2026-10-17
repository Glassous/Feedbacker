<script setup lang="ts">
import { computed } from 'vue'
import { RotateCcw, Search, X } from 'lucide-vue-next'
import AppSelect from './ui/AppSelect.vue'

const props = defineProps<{
  q: string
  status: string
  project: string
  projects: string[]
}>()

const emit = defineEmits<{
  'update:q': [value: string]
  'update:status': [value: string]
  'update:project': [value: string]
  reset: []
}>()

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '未处理', value: 'open' },
  { label: '已处理', value: 'resolved' },
]

const projectOptions = computed(() => [
  { label: '全部来源', value: '' },
  ...props.projects.map((project) => ({ label: project, value: project })),
])

const hasFilters = computed(() => Boolean(props.q || props.status || props.project))
</script>

<template>
  <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
    <div class="relative w-full xl:max-w-xs">
      <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-mist-500" />
      <input
        :value="q"
        type="search"
        placeholder="搜索反馈…"
        class="h-9 w-full rounded-lg border border-black/10 bg-ink-950 pl-9 pr-9 text-[13px] text-mist-100 outline-none transition-all duration-200 placeholder:text-mist-500 focus:border-black/30 focus:ring-2 focus:ring-black/10"
        @input="emit('update:q', ($event.target as HTMLInputElement).value)"
      />
      <button
        v-if="q"
        type="button"
        class="absolute right-2 top-1/2 flex size-5 -translate-y-1/2 cursor-pointer items-center justify-center rounded text-mist-500 transition-colors hover:text-mist-200"
        aria-label="清空搜索"
        @click="emit('update:q', '')"
      >
        <X class="size-3" />
      </button>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <div class="w-[124px]">
        <AppSelect
          :model-value="status"
          :options="statusOptions"
          @update:model-value="emit('update:status', $event)"
        />
      </div>
      <div class="w-[140px]">
        <AppSelect
          :model-value="project"
          :options="projectOptions"
          @update:model-value="emit('update:project', $event)"
        />
      </div>
      <button
        v-if="hasFilters"
        type="button"
        class="inline-flex h-9 cursor-pointer items-center gap-1 rounded-lg px-2 text-[11px] text-mist-500 transition-colors hover:text-mist-200"
        @click="emit('reset')"
      >
        <RotateCcw class="size-3" />
        重置
      </button>
    </div>
  </div>
</template>
