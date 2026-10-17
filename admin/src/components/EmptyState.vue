<script setup lang="ts">
import type { Component } from 'vue'
import { Inbox } from 'lucide-vue-next'

withDefaults(defineProps<{ icon?: Component; title: string; description: string }>(), {
  icon: Inbox,
})
</script>

<template>
  <div class="flex flex-col items-center justify-center px-6 py-16 text-center">
    <component :is="icon" class="size-5 text-mist-500" />
    <p class="mt-3 text-sm text-mist-300">{{ title }}</p>
    <p class="mt-1 max-w-xs text-xs leading-5 text-mist-500">{{ description }}</p>
    <div v-if="$slots.action" class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>
