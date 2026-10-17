<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { X } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    description?: string
    closable?: boolean
    width?: 'sm' | 'md' | 'lg'
  }>(),
  { closable: true, width: 'md' },
)

const emit = defineEmits<{ close: [] }>()

const widths = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-xl',
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && props.closable) {
    emit('close')
  }
}

watch(
  () => props.open,
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
  document.body.style.overflow = ''
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div
          class="absolute inset-0 bg-black/40 backdrop-blur-sm"
          @click="closable && emit('close')"
        />
        <div
          class="relative w-full animate-[fade-up_0.25s_cubic-bezier(0.22,1,0.36,1)_both] rounded-xl border border-black/10 bg-ink-800 p-5 shadow-[0_24px_60px_-24px_rgba(0,0,0,0.2)]"
          :class="widths[width]"
          role="dialog"
          aria-modal="true"
        >
          <button
            v-if="closable"
            type="button"
            class="absolute right-4 top-4 flex size-8 cursor-pointer items-center justify-center rounded-lg text-mist-400 transition-colors hover:bg-black/5 hover:text-mist-100"
            aria-label="关闭"
            @click="emit('close')"
          >
            <X class="size-4" />
          </button>
          <div v-if="title" class="pr-8">
            <h2 class="text-[15px] font-medium text-mist-100">{{ title }}</h2>
            <p v-if="description" class="mt-1 text-xs leading-5 text-mist-500">
              {{ description }}
            </p>
          </div>
          <div class="mt-5">
            <slot />
          </div>
          <div v-if="$slots.footer" class="mt-6 flex items-center justify-end gap-3">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
