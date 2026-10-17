<script setup lang="ts">
import { AlertTriangle } from 'lucide-vue-next'
import AppModal from './ui/AppModal.vue'
import AppButton from './ui/AppButton.vue'

withDefaults(
  defineProps<{
    open: boolean
    title: string
    description: string
    confirmText?: string
    loading?: boolean
    tone?: 'danger' | 'primary'
  }>(),
  { confirmText: '确认', loading: false, tone: 'danger' },
)

const emit = defineEmits<{ confirm: []; close: [] }>()
</script>

<template>
  <AppModal :open="open" width="sm" @close="emit('close')">
    <div class="flex items-start gap-3">
      <span
        class="flex size-10 shrink-0 items-center justify-center rounded-xl border"
        :class="
          tone === 'danger'
            ? 'border-signal-red/25 bg-signal-red/10 text-signal-red'
            : 'border-black/10 bg-black/5 text-mist-300'
        "
      >
        <AlertTriangle class="size-[18px]" />
      </span>
      <div class="min-w-0">
        <h2 class="text-base font-semibold text-mist-100">{{ title }}</h2>
        <p class="mt-1.5 text-sm leading-6 text-mist-400">{{ description }}</p>
      </div>
    </div>

    <div class="mt-6 flex items-center justify-end gap-3">
      <AppButton variant="ghost" size="sm" :disabled="loading" @click="emit('close')">取消</AppButton>
      <AppButton
        :variant="tone === 'danger' ? 'danger' : 'primary'"
        size="sm"
        :loading="loading"
        @click="emit('confirm')"
      >
        {{ confirmText }}
      </AppButton>
    </div>
  </AppModal>
</template>
