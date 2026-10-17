<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { CheckCircle2, Clock3, RotateCcw, Trash2, X } from 'lucide-vue-next'
import type { Feedback } from '../lib/api'
import { deleteFeedback, errorMessage, updateFeedbackStatus } from '../lib/api'
import { browserLabel, formatFullDateTime, relativeTime } from '../lib/format'
import { pushToast } from '../composables/useToast'
import AppButton from './ui/AppButton.vue'
import ConfirmDialog from './ConfirmDialog.vue'
import CopyButton from './CopyButton.vue'
import StatusPill from './StatusPill.vue'

const props = defineProps<{ feedback: Feedback | null; open: boolean }>()

const emit = defineEmits<{
  close: []
  updated: [feedback: Feedback]
  deleted: [id: string]
}>()

const actionLoading = ref(false)
const confirmOpen = ref(false)
const deleteLoading = ref(false)

const isOpen = computed(() => props.open && props.feedback !== null)
const isResolved = computed(() => props.feedback?.status === 'resolved')

function close() {
  if (actionLoading.value) return
  emit('close')
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && isOpen.value && !confirmOpen.value) {
    close()
  }
}

watch(
  isOpen,
  (open) => {
    document.body.style.overflow = open ? 'hidden' : ''
    if (open) {
      window.addEventListener('keydown', onKeydown)
    } else {
      window.removeEventListener('keydown', onKeydown)
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})

function toggleStatus() {
  const current = props.feedback
  if (!current) return

  const nextStatus = current.status === 'open' ? 'resolved' : 'open'
  actionLoading.value = true

  updateFeedbackStatus(current.id, nextStatus)
    .then((updated) => {
      emit('updated', updated)
      pushToast({
        title: nextStatus === 'resolved' ? '已标记为已处理' : '已重新打开',
        description: nextStatus === 'resolved' ? '这条反馈已完成跟进' : '这条反馈回到待处理列表',
        variant: 'success',
      })
    })
    .catch((error) => {
      console.error(error)
      pushToast({ title: '更新状态失败', description: errorMessage(error), variant: 'error' })
    })
    .finally(() => {
      actionLoading.value = false
    })
}

function removeFeedback() {
  const current = props.feedback
  if (!current) return

  deleteLoading.value = true
  deleteFeedback(current.id)
    .then(() => {
      confirmOpen.value = false
      emit('deleted', current.id)
    })
    .catch((error) => {
      console.error(error)
      pushToast({ title: '删除失败', description: errorMessage(error), variant: 'error' })
    })
    .finally(() => {
      deleteLoading.value = false
    })
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isOpen"
        class="fixed inset-0 z-40 bg-black/35 backdrop-blur-sm"
        @click="close"
      />
    </Transition>

    <Transition
      enter-active-class="transition-transform duration-300 ease-[cubic-bezier(0.22,1,0.36,1)]"
      enter-from-class="translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition-transform duration-200 ease-in"
      leave-from-class="translate-x-0"
      leave-to-class="translate-x-full"
    >
      <aside
        v-if="isOpen && feedback"
        class="fixed inset-y-0 right-0 z-40 flex w-full max-w-[520px] flex-col border-l border-black/8 bg-ink-900 shadow-[-30px_0_80px_-40px_rgba(0,0,0,0.2)]"
      >
        <header class="border-b border-black/6 px-5 py-4 sm:px-6">
          <div class="flex items-start justify-between gap-4">
            <div class="flex items-center gap-2">
              <StatusPill :status="feedback.status" />
              <span class="flex items-center gap-1 text-[10px] text-mist-500">
                <Clock3 class="size-3" />
                {{ relativeTime(feedback.createdAt) }}
              </span>
            </div>
            <button
              type="button"
              class="flex size-7 cursor-pointer items-center justify-center rounded-md text-mist-500 transition-colors hover:bg-black/[0.04] hover:text-mist-200"
              aria-label="关闭详情"
              @click="close"
            >
              <X class="size-4" />
            </button>
          </div>

          <h2 class="mt-2.5 text-base font-medium leading-6 text-mist-100">
            {{ feedback.title || '（无标题反馈）' }}
          </h2>
        </header>

        <div class="flex-1 overflow-y-auto px-5 py-5 sm:px-6">
          <p
            v-if="feedback.content"
            class="whitespace-pre-wrap break-words text-[13px] leading-7 text-mist-200"
          >
            {{ feedback.content }}
          </p>
          <p v-else class="text-[13px] italic text-mist-500">（无正文）</p>

          <dl class="mt-5 grid grid-cols-2 gap-2">
            <div class="rounded-lg border border-black/8 bg-black/[0.02] p-3">
              <dt class="text-[10px] text-mist-500">来源</dt>
              <dd class="mt-1 truncate text-[11px] text-mist-300">{{ feedback.project || '—' }}</dd>
            </div>

            <div class="rounded-lg border border-black/8 bg-black/[0.02] p-3">
              <dt class="text-[10px] text-mist-500">联系方式</dt>
              <dd class="mt-1 flex items-center justify-between gap-2">
                <span class="truncate text-[11px] text-mist-300">{{ feedback.contact || '—' }}</span>
                <CopyButton v-if="feedback.contact" :value="feedback.contact" label="" />
              </dd>
            </div>

            <div class="col-span-2 rounded-lg border border-black/8 bg-black/[0.02] p-3">
              <dt class="text-[10px] text-mist-500">页面地址</dt>
              <dd class="mt-1 flex items-center justify-between gap-3">
                <a
                  v-if="feedback.pageUrl"
                  :href="feedback.pageUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="truncate text-[11px] text-mist-300 underline-offset-4 transition-colors hover:text-mist-100 hover:underline"
                >
                  {{ feedback.pageUrl }}
                </a>
                <span v-else class="text-[11px] text-mist-500">—</span>
                <CopyButton v-if="feedback.pageUrl" :value="feedback.pageUrl" label="" />
              </dd>
            </div>

            <div class="rounded-lg border border-black/8 bg-black/[0.02] p-3">
              <dt class="text-[10px] text-mist-500">客户端</dt>
              <dd class="mt-1 truncate text-[11px] text-mist-300">{{ browserLabel(feedback.userAgent) }}</dd>
            </div>

            <div class="rounded-lg border border-black/8 bg-black/[0.02] p-3">
              <dt class="text-[10px] text-mist-500">IP</dt>
              <dd class="mt-1 truncate font-mono text-[11px] text-mist-300">{{ feedback.ip || '—' }}</dd>
            </div>

            <div class="rounded-lg border border-black/8 bg-black/[0.02] p-3">
              <dt class="text-[10px] text-mist-500">提交时间</dt>
              <dd class="mt-1 text-[11px] text-mist-300">{{ formatFullDateTime(feedback.createdAt) }}</dd>
            </div>

            <div class="rounded-lg border border-black/8 bg-black/[0.02] p-3">
              <dt class="text-[10px] text-mist-500">更新时间</dt>
              <dd class="mt-1 text-[11px] text-mist-300">{{ formatFullDateTime(feedback.updatedAt) }}</dd>
            </div>

            <div class="col-span-2 rounded-lg border border-black/8 bg-black/[0.02] p-3">
              <dt class="text-[10px] text-mist-500">User-Agent</dt>
              <dd class="mt-1 break-all font-mono text-[10px] leading-5 text-mist-500">
                {{ feedback.userAgent || '—' }}
              </dd>
            </div>
          </dl>
        </div>

        <footer class="border-t border-black/6 px-5 py-3.5 sm:px-6">
          <div class="flex items-center gap-2">
            <AppButton variant="ghost" class="flex-1" :loading="actionLoading" @click="toggleStatus">
              <RotateCcw v-if="isResolved" class="size-3.5" />
              <CheckCircle2 v-else class="size-3.5" />
              {{ isResolved ? '重新打开' : '标记已处理' }}
            </AppButton>
            <AppButton variant="danger" class="shrink-0" @click="confirmOpen = true">
              <Trash2 class="size-3.5" />
              删除
            </AppButton>
          </div>
        </footer>
      </aside>
    </Transition>

    <ConfirmDialog
      :open="confirmOpen"
      title="删除这条反馈？"
      description="删除后不可恢复。"
      confirm-text="删除"
      :loading="deleteLoading"
      @confirm="removeFeedback"
      @close="confirmOpen = false"
    />
  </Teleport>
</template>
