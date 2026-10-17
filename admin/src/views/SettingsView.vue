<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { LogOut } from 'lucide-vue-next'
import AppButton from '../components/ui/AppButton.vue'
import AppInput from '../components/ui/AppInput.vue'
import CopyButton from '../components/CopyButton.vue'
import { changePassword, errorMessage, logout } from '../lib/api'
import { useAuth } from '../composables/useAuth'
import { pushToast } from '../composables/useToast'

const router = useRouter()
const { state, setSession, clearSession } = useAuth()

const form = reactive({ oldPassword: '', newPassword: '', confirm: '' })
const errors = reactive({ old: '', next: '', confirm: '' })
const submitting = ref(false)

const endpoint = computed(() => `${window.location.origin}/api/v1/feedback`)

const sample = computed(() =>
  [
    `curl -X POST ${endpoint.value} \\`,
    `  -H "Content-Type: application/json" \\`,
    `  -d '{"title":"无法导出报表","content":"点击导出后没有反应","contact":"you@example.com","project":"web-console"}'`,
  ].join('\n'),
)

function submitPassword() {
  errors.old = form.oldPassword ? '' : '请输入当前密码'
  errors.next = form.newPassword.length >= 8 ? '' : '至少需要 8 个字符'
  errors.confirm = form.confirm === form.newPassword ? '' : '两次输入不一致'
  if (errors.old || errors.next || errors.confirm) return

  submitting.value = true
  changePassword({ oldPassword: form.oldPassword, newPassword: form.newPassword })
    .then((result) => {
      setSession({
        token: result.token,
        username: result.username,
        mustChangePassword: result.mustChangePassword,
      })
      form.oldPassword = ''
      form.newPassword = ''
      form.confirm = ''
      pushToast({ title: '密码已更新', variant: 'success' })
    })
    .catch((error) => {
      console.error(error)
      errors.old = errorMessage(error, '修改密码失败')
    })
    .finally(() => {
      submitting.value = false
    })
}

function signOut() {
  logout()
    .catch((error) => console.error(error))
    .finally(() => {
      clearSession()
      router.replace({ name: 'login' })
    })
}
</script>

<template>
  <div class="mx-auto max-w-3xl space-y-5">
    <section class="rounded-xl border border-black/8 bg-ink-900">
      <div class="flex items-center justify-between border-b border-black/6 px-5 py-3.5">
        <h2 class="text-[13px] text-mist-300">修改密码</h2>
        <span class="font-mono text-[11px] text-mist-500">{{ state.username || 'admin' }}</span>
      </div>

      <form class="p-5" @submit.prevent="submitPassword">
        <div class="grid gap-3 sm:grid-cols-2">
          <div class="sm:col-span-2">
            <AppInput
              v-model="form.oldPassword"
              label="当前密码"
              type="password"
              autocomplete="current-password"
              placeholder="输入当前密码"
              :error="errors.old"
            />
          </div>
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
            placeholder="再次输入"
            :error="errors.confirm"
          />
        </div>

        <div class="mt-5 flex items-center justify-between">
          <AppButton variant="link" size="sm" type="button" @click="signOut">
            <LogOut class="size-3.5" />
            退出登录
          </AppButton>
          <AppButton type="submit" size="sm" :loading="submitting">保存</AppButton>
        </div>
      </form>
    </section>

    <section class="rounded-xl border border-black/8 bg-ink-900">
      <div class="flex items-center justify-between border-b border-black/6 px-5 py-3.5">
        <h2 class="text-[13px] text-mist-300">接口接入</h2>
        <CopyButton :value="endpoint" label="复制地址" />
      </div>

      <div class="p-5">
        <code class="block truncate rounded-lg bg-ink-950 px-3 py-2.5 font-mono text-[11px] text-mist-300">
          {{ endpoint }}
        </code>

        <div class="mt-4 flex items-center justify-between">
          <span class="font-mono text-[10px] text-mist-500">POST</span>
          <CopyButton :value="sample" label="复制示例" />
        </div>

        <pre class="mt-2 overflow-x-auto rounded-lg bg-ink-950 p-3.5 font-mono text-[11px] leading-6 text-mist-300"><code>{{ sample }}</code></pre>
      </div>
    </section>
  </div>
</template>
