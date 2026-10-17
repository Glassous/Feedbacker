<script setup lang="ts">
import { computed, ref, useId } from 'vue'
import { Eye, EyeOff } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    modelValue: string
    label?: string
    type?: string
    placeholder?: string
    autocomplete?: string
    error?: string
    hint?: string
    disabled?: boolean
    autofocus?: boolean
  }>(),
  { type: 'text', disabled: false, autofocus: false },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const fieldId = useId()
const reveal = ref(false)
const isPassword = computed(() => props.type === 'password')
const inputType = computed(() => (isPassword.value && reveal.value ? 'text' : props.type))

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>

<template>
  <label :for="fieldId" class="block">
    <span v-if="label" class="mb-1.5 block text-xs text-mist-400">{{ label }}</span>
    <div class="group relative flex items-center">
      <span
        v-if="$slots.leading"
        class="pointer-events-none absolute left-3 flex items-center text-mist-500 transition-colors group-focus-within:text-mist-200"
      >
        <slot name="leading" />
      </span>
      <input
        :id="fieldId"
        :value="modelValue"
        :type="inputType"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :disabled="disabled"
        :autofocus="autofocus"
        :class="[
          'h-10 w-full rounded-lg border bg-ink-950 text-sm text-mist-100 outline-none transition-all duration-200 placeholder:text-mist-500',
          $slots.leading ? 'pl-9' : 'pl-3',
          isPassword ? 'pr-10' : 'pr-3',
          error
            ? 'border-signal-red/50 focus:border-signal-red/70'
            : 'border-black/10 focus:border-black/30 focus:ring-2 focus:ring-black/10',
          disabled ? 'cursor-not-allowed opacity-60' : '',
        ]"
        @input="onInput"
      />
      <button
        v-if="isPassword"
        type="button"
        class="absolute right-2 flex size-7 cursor-pointer items-center justify-center rounded-md text-mist-500 transition-colors hover:bg-black/5 hover:text-mist-200"
        :aria-label="reveal ? '隐藏密码' : '显示密码'"
        @click="reveal = !reveal"
      >
        <EyeOff v-if="reveal" class="size-4" />
        <Eye v-else class="size-4" />
      </button>
    </div>
    <span v-if="error" class="mt-1.5 block text-xs text-signal-red">{{ error }}</span>
    <span v-else-if="hint" class="mt-1.5 block text-xs text-mist-500">{{ hint }}</span>
  </label>
</template>
