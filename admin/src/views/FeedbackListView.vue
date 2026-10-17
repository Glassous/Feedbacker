<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { CheckCircle2, CircleDot, Inbox, RefreshCw, SearchX } from 'lucide-vue-next'
import type { Feedback, FeedbackCounts } from '../lib/api'
import {
  errorMessage,
  fetchFeedbackList,
  fetchProjects,
  isCanceledError,
} from '../lib/api'
import FilterBar from '../components/FilterBar.vue'
import StatCard from '../components/StatCard.vue'
import FeedbackTable from '../components/FeedbackTable.vue'
import FeedbackCard from '../components/FeedbackCard.vue'
import PaginationBar from '../components/PaginationBar.vue'
import EmptyState from '../components/EmptyState.vue'
import FeedbackDrawer from '../components/FeedbackDrawer.vue'
import AppButton from '../components/ui/AppButton.vue'
import { pushToast } from '../composables/useToast'
import { useAuth } from '../composables/useAuth'

const { state } = useAuth()

const items = ref<Feedback[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const counts = ref<FeedbackCounts>({ total: 0, open: 0, resolved: 0 })
const loading = ref(true)
const projects = ref<string[]>([])
const selected = ref<Feedback | null>(null)

const filters = reactive({ q: '', status: '', project: '' })
const hasFilters = computed(() => Boolean(filters.q || filters.status || filters.project))

let controller: AbortController | null = null
let requestSeq = 0
let debounceTimer: number | undefined

function load() {
  // 一次性密码尚未修改时，后端会拒绝业务请求，这里直接等待改密完成。
  if (state.mustChangePassword) {
    loading.value = false
    return
  }

  controller?.abort()
  controller = new AbortController()
  const seq = ++requestSeq
  loading.value = true

  fetchFeedbackList(
    {
      page: page.value,
      pageSize: pageSize.value,
      q: filters.q.trim() || undefined,
      status: filters.status || undefined,
      project: filters.project || undefined,
    },
    controller.signal,
  )
    .then((data) => {
      if (seq !== requestSeq) return
      items.value = data.items
      total.value = data.total
      counts.value = data.counts
      page.value = data.page
    })
    .catch((error) => {
      if (isCanceledError(error)) return
      console.error(error)
      pushToast({ title: '加载反馈失败', description: errorMessage(error), variant: 'error' })
    })
    .finally(() => {
      if (seq === requestSeq) loading.value = false
    })
}

function loadProjects() {
  if (state.mustChangePassword) return

  fetchProjects()
    .then((data) => {
      projects.value = data
    })
    .catch((error) => {
      console.error('加载来源项目失败', error)
    })
}

watch(
  () => filters.q,
  () => {
    page.value = 1
    window.clearTimeout(debounceTimer)
    debounceTimer = window.setTimeout(load, 300)
  },
)

watch([() => filters.status, () => filters.project], () => {
  page.value = 1
  load()
})

function changePage(next: number) {
  if (next < 1) return
  page.value = next
  load()
}

function resetFilters() {
  window.clearTimeout(debounceTimer)
  filters.q = ''
  filters.status = ''
  filters.project = ''
  page.value = 1
  load()
}

function onUpdated(updated: Feedback) {
  const previous = items.value.find((item) => item.id === updated.id)
  items.value = items.value.map((item) => (item.id === updated.id ? updated : item))
  if (selected.value?.id === updated.id) {
    selected.value = updated
  }

  if (previous && previous.status !== updated.status) {
    counts.value = {
      ...counts.value,
      open: Math.max(0, counts.value.open + (updated.status === 'open' ? 1 : -1)),
      resolved: Math.max(0, counts.value.resolved + (updated.status === 'resolved' ? 1 : -1)),
    }
  }

  if (filters.status && filters.status !== updated.status) {
    load()
  }
}

function onDeleted(id: string) {
  if (selected.value?.id === id) {
    selected.value = null
  }
  pushToast({ title: '反馈已删除', variant: 'success' })
  if (items.value.length === 1 && page.value > 1) {
    page.value -= 1
  }
  load()
  loadProjects()
}

onMounted(() => {
  if (state.mustChangePassword) return
  load()
  loadProjects()
})

watch(
  () => state.mustChangePassword,
  (mustChange) => {
    if (!mustChange) {
      load()
      loadProjects()
    }
  },
)

onBeforeUnmount(() => {
  controller?.abort()
  window.clearTimeout(debounceTimer)
})
</script>

<template>
  <div class="space-y-5">
    <section class="grid grid-cols-3 gap-2.5 sm:gap-3">
      <StatCard label="总数" :value="counts.total" :icon="Inbox" />
      <StatCard label="待处理" :value="counts.open" :icon="CircleDot" tone="amber" />
      <StatCard label="已处理" :value="counts.resolved" :icon="CheckCircle2" tone="green" />
    </section>

    <section class="overflow-hidden rounded-xl border border-black/8 bg-ink-900">
      <div class="border-b border-black/6 p-3.5 sm:p-4">
        <div class="mb-3 flex items-center justify-between gap-3">
          <h2 class="text-[13px] text-mist-300">全部反馈</h2>
          <AppButton variant="ghost" size="icon" aria-label="刷新" :loading="loading" @click="load">
            <RefreshCw v-if="!loading" class="size-3.5" />
          </AppButton>
        </div>

        <FilterBar
          v-model:q="filters.q"
          v-model:status="filters.status"
          v-model:project="filters.project"
          :projects="projects"
          @reset="resetFilters"
        />
      </div>

      <div v-if="loading && items.length === 0" class="space-y-4 p-5">
        <div v-for="n in 6" :key="n" class="flex items-center gap-4">
          <div class="skeleton h-6 w-20 rounded-full" />
          <div class="flex-1 space-y-2">
            <div class="skeleton h-3.5 w-1/3" />
            <div class="skeleton h-3 w-2/3" />
          </div>
          <div class="skeleton h-3 w-16" />
          <div class="skeleton h-3 w-20" />
        </div>
      </div>

      <EmptyState
        v-else-if="items.length === 0"
        :icon="hasFilters ? SearchX : Inbox"
        :title="hasFilters ? '没有匹配的反馈' : '还没有收到反馈'"
        :description="hasFilters ? '调整筛选条件后重试。' : '通过接口提交的反馈会显示在这里。'"
      >
        <template v-if="hasFilters" #action>
          <AppButton variant="ghost" size="sm" @click="resetFilters">重置筛选</AppButton>
        </template>
      </EmptyState>

      <template v-else>
        <div :class="loading ? 'pointer-events-none opacity-60 transition-opacity' : 'transition-opacity'">
          <FeedbackTable :items="items" @select="selected = $event" />
        </div>
        <div class="divide-y divide-black/5 md:hidden">
          <FeedbackCard
            v-for="item in items"
            :key="item.id"
            :item="item"
            @select="selected = $event"
          />
        </div>
        <PaginationBar :page="page" :page-size="pageSize" :total="total" @change="changePage" />
      </template>
    </section>

    <FeedbackDrawer
      :feedback="selected"
      :open="selected !== null"
      @close="selected = null"
      @updated="onUpdated"
      @deleted="onDeleted"
    />
  </div>
</template>
