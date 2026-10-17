<script setup lang="ts">
import { ref } from 'vue'
import { Check, Copy } from 'lucide-vue-next'
import { copyText } from '../lib/clipboard'
import { pushToast } from '../composables/useToast'

const props = withDefaults(defineProps<{ value: string; label?: string }>(), { label: '复制' })

const copied = ref(false)

function copy() {
  copyText(props.value)
    .then(() => {
      copied.value = true
      window.setTimeout(() => {
        copied.value = false
      }, 1600)
    })
    .catch((error) => {
      console.error(error)
      pushToast({ title: '复制失败', description: '请手动选择文本复制', variant: 'error' })
    })
}
</script>

<template>
  <button
    type="button"
    class="inline-flex cursor-pointer items-center gap-1.5 rounded-lg border border-black/10 bg-black/5 px-2.5 py-1.5 text-xs text-mist-300 transition-all duration-200 hover:border-black/20 hover:text-mist-100"
    @click.stop="copy"
  >
    <Check v-if="copied" class="size-3.5 text-signal-green" />
    <Copy v-else class="size-3.5" />
    <span>{{ copied ? '已复制' : label }}</span>
  </button>
</template>
