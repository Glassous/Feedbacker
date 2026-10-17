<script setup lang="ts">
import { onMounted } from 'vue'
import ToastHost from './components/ToastHost.vue'
import { useSite } from './composables/useSite'

const { load } = useSite()

onMounted(() => {
  load()
})
</script>

<template>
  <RouterView />
  <ToastHost />
</template>
