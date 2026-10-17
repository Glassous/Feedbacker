import { ref } from 'vue'

export type ToastVariant = 'success' | 'error' | 'info'

export interface ToastItem {
  id: number
  title: string
  description?: string
  variant: ToastVariant
}

const items = ref<ToastItem[]>([])
let sequence = 0

export function dismissToast(id: number) {
  items.value = items.value.filter((item) => item.id !== id)
}

export function pushToast(input: { title: string; description?: string; variant?: ToastVariant }) {
  sequence += 1
  const toast: ToastItem = {
    id: sequence,
    title: input.title,
    description: input.description,
    variant: input.variant ?? 'info',
  }
  items.value = [...items.value, toast]
  window.setTimeout(() => dismissToast(toast.id), 3800)
}

export function useToast() {
  return { items, pushToast, dismissToast }
}
