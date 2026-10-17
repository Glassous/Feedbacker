<script setup lang="ts">
import { ChevronDown } from 'lucide-vue-next'

export interface SelectOption {
  label: string
  value: string
}

withDefaults(
  defineProps<{
    modelValue: string
    options: SelectOption[]
    placeholder?: string
    label?: string
    disabled?: boolean
  }>(),
  { disabled: false },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function onChange(event: Event) {
  emit('update:modelValue', (event.target as HTMLSelectElement).value)
}
</script>

<template>
  <label class="block">
    <span v-if="label" class="mb-1.5 block text-xs text-mist-400">{{ label }}</span>
    <div class="relative">
      <select
        :value="modelValue"
        :disabled="disabled"
        class="h-9 w-full cursor-pointer appearance-none rounded-lg border border-black/10 bg-ink-950 pl-3 pr-8 text-[13px] text-mist-200 outline-none transition-all duration-200 focus:border-black/30 focus:ring-2 focus:ring-black/10 disabled:cursor-not-allowed disabled:opacity-60"
        @change="onChange"
      >
        <option v-if="placeholder" value="">{{ placeholder }}</option>
        <option v-for="option in options" :key="option.value" :value="option.value">
          {{ option.label }}
        </option>
      </select>
      <ChevronDown
        class="pointer-events-none absolute right-2.5 top-1/2 size-3.5 -translate-y-1/2 text-mist-500"
      />
    </div>
  </label>
</template>
