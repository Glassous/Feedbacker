<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import AppModal from './ui/AppModal.vue'
import AppInput from './ui/AppInput.vue'
import AppButton from './ui/AppButton.vue'
import { changePassword, errorMessage } from '../lib/api'
import { useAuth } from '../composables/useAuth'
import { pushToast } from '../composables/useToast'

const props = withDefaults(defineProps<{ open: boolean; forced?: boolean }>(), { forced: false })
const emit = defineEmits<{ close: []; success: [] }>()

const { setSession } = useAuth()

const form = reactive({ oldPassword: '', newPassword: '', confirm: '' })
const errors = reactive({ old: '', next: '', confirm: '' })
const serverError = ref('')
const submitting = ref(false)

watch(
  () => props.open,
  (open) => {
    if (!open) return
    form.oldPassword = ''
    form.newPassword = ''
    form.confirm = ''
    errors.old = ''
    errors.next = ''
    errors.confirm = ''
    serverError.value = ''
  },
)

function submit() {
  errors.old = form.oldPassword ? '' : '请输入当前密码'
  errors.next = form.newPassword.length >= 8 ? '' : '新密码至少需要 8 个字符'
  errors.confirm = form.confirm === form.newPassword ? '' : '两次输入的密码不一致'
  if (errors.old || errors.next || errors.confirm) return

  submitting.value = true
  serverError.value = ''

  changePassword({ oldPassword: form.oldPassword, newPassword: form.newPassword })
    .then((result) => {
      setSession({
        token: result.token,
        username: result.username,
        mustChangePassword: result.mustChangePassword,
      })
      pushToast({ title: '密码已更新', description: '其他设备上的登录状态已失效', variant: 'success' })
      emit('success')
    })
    .catch((error) => {
      console.error(error)
      serverError.value = errorMessage(error, '修改密码失败，请稍后重试')
    })
    .finally(() => {
      submitting.value = false
    })
}

function close() {
  if (props.forced) return
  emit('close')
}
</script>

<template>
  <AppModal
    :open="open"
    :closable="!forced"
    title="修改密码"
    :description="forced ? '设置新密码后继续。' : '其他设备将需要重新登录。'"
    @close="close"
  >
    <form class="space-y-4" @submit.prevent="submit">
      <p
        v-if="forced"
        class="rounded-lg border border-signal-amber/20 bg-signal-amber/[0.06] px-3 py-2 text-[11px] text-signal-amber"
      >
        一次性密码仅用于首次登录。
      </p>

      <AppInput
        v-model="form.oldPassword"
        label="当前密码"
        type="password"
        autocomplete="current-password"
        placeholder="输入登录时使用的密码"
        :error="errors.old"
        autofocus
      />
      <AppInput
        v-model="form.newPassword"
        label="新密码"
        type="password"
        autocomplete="new-password"
        placeholder="至少 8 个字符"
        :error="errors.next"
      />
      <AppInput
        v-model="form.confirm"
        label="确认新密码"
        type="password"
        autocomplete="new-password"
        placeholder="再次输入新密码"
        :error="errors.confirm"
      />

      <p v-if="serverError" class="rounded-lg border border-signal-red/25 bg-signal-red/10 px-3 py-2 text-xs text-signal-red">
        {{ serverError }}
      </p>

      <div class="flex items-center justify-end gap-3 pt-2">
        <AppButton v-if="!forced" variant="ghost" size="sm" @click="close">取消</AppButton>
        <AppButton type="submit" size="sm" :loading="submitting">保存</AppButton>
      </div>
    </form>
  </AppModal>
</template>
