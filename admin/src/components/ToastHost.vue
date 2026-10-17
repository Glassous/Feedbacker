<script setup lang="ts">
import { CheckCircle2, Info, X, XCircle } from 'lucide-vue-next'
import { useToast, type ToastVariant } from '../composables/useToast'

const { items, dismissToast } = useToast()

const accents: Record<ToastVariant, string> = {
  success: 'text-signal-green',
  error: 'text-signal-red',
  info: 'text-mist-300',
}
</script>

<template>
  <Teleport to="body">
    <div
      class="pointer-events-none fixed right-4 top-4 z-[70] flex w-[calc(100%-2rem)] max-w-sm flex-col gap-2"
    >
      <TransitionGroup
        enter-active-class="transition duration-300 ease-out"
        enter-from-class="-translate-y-2 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition duration-200 ease-in absolute w-full"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="-translate-y-2 opacity-0"
        move-class="transition duration-200"
      >
        <div
          v-for="toast in items"
          :key="toast.id"
          class="pointer-events-auto flex items-start gap-3 rounded-lg border border-black/10 bg-ink-800 p-3.5 shadow-[0_16px_40px_-24px_rgba(0,0,0,0.18)]"
        >
          <CheckCircle2 v-if="toast.variant === 'success'" class="mt-0.5 size-4 shrink-0" :class="accents.success" />
          <XCircle v-else-if="toast.variant === 'error'" class="mt-0.5 size-4 shrink-0" :class="accents.error" />
          <Info v-else class="mt-0.5 size-4 shrink-0" :class="accents.info" />
          <div class="min-w-0 flex-1">
            <p class="text-[13px] text-mist-100">{{ toast.title }}</p>
            <p v-if="toast.description" class="mt-0.5 text-[11px] leading-5 text-mist-500">
              {{ toast.description }}
            </p>
          </div>
          <button
            type="button"
            class="flex size-5 cursor-pointer items-center justify-center rounded text-mist-500 transition-colors hover:text-mist-200"
            aria-label="关闭提示"
            @click="dismissToast(toast.id)"
          >
            <X class="size-3" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
