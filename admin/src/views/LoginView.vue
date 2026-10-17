<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ChevronDown, MessageSquare } from 'lucide-vue-next'
import AppButton from '../components/ui/AppButton.vue'
import AppInput from '../components/ui/AppInput.vue'
import CopyButton from '../components/CopyButton.vue'
import { errorMessage, login } from '../lib/api'
import { pushToast } from '../composables/useToast'
import { useAuth } from '../composables/useAuth'
import { useSite } from '../composables/useSite'

const router = useRouter()
const route = useRoute()
const { setSession } = useAuth()
const { siteName } = useSite()

const form = reactive({ username: 'admin', password: '' })
const loading = ref(false)
const error = ref('')
const showUsername = ref(false)
const showHelp = ref(false)

const resetCommand = 'docker compose exec app ./feedbacker admin reset-password'

function submit() {
  if (!form.password.trim()) {
    error.value = '请输入密码'
    return
  }

  loading.value = true
  error.value = ''

  login({ username: form.username.trim() || 'admin', password: form.password })
    .then((result) => {
      setSession({
        token: result.token,
        username: result.username,
        mustChangePassword: result.mustChangePassword,
      })
      pushToast({
        title: '登录成功',
        description: result.mustChangePassword ? '请先修改密码' : undefined,
        variant: 'success',
      })
      const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
      router.replace(redirect)
    })
    .catch((err) => {
      console.error(err)
      error.value = errorMessage(err, '登录失败')
    })
    .finally(() => {
      loading.value = false
    })
}
</script>

<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-ink-950 px-4 py-10">
    <div class="pointer-events-none absolute inset-0 grid-overlay" />

    <div class="relative w-full max-w-[360px]">
      <div class="flex flex-col items-center text-center">
        <span class="flex size-10 items-center justify-center rounded-xl bg-mist-100 text-ink-950">
          <MessageSquare class="size-[18px]" />
        </span>
        <p class="mt-4 text-[15px] font-medium text-mist-100">{{ siteName }}</p>
        <p class="mt-1 text-[11px] text-mist-400">反馈管理后台</p>
      </div>

      <div class="mt-7 rounded-xl border border-black/10 bg-ink-900 p-6 shadow-[0_18px_40px_-30px_rgba(0,0,0,0.18)]">
        <div class="flex items-center justify-between">
          <h1 class="text-[13px] text-mist-300">管理员登录</h1>
          <button
            type="button"
            class="cursor-pointer text-[11px] text-mist-400 transition-colors hover:text-mist-200"
            @click="showUsername = !showUsername"
          >
            切换账号
          </button>
        </div>

        <form class="mt-5 space-y-4" @submit.prevent="submit">
          <input
            type="text"
            name="username"
            autocomplete="username"
            :value="form.username"
            class="sr-only"
            tabindex="-1"
            aria-hidden="true"
          />
          <AppInput
            v-if="showUsername"
            v-model="form.username"
            label="用户名"
            autocomplete="username"
            placeholder="admin"
          />
          <AppInput
            v-model="form.password"
            label="密码"
            type="password"
            autocomplete="current-password"
            placeholder="输入密码"
            autofocus
          />

          <p v-if="error" class="text-xs text-signal-red">{{ error }}</p>

          <AppButton type="submit" block :loading="loading">登录</AppButton>
        </form>

        <button
          type="button"
          class="mt-4 flex w-full cursor-pointer items-center justify-center gap-1 text-[11px] text-mist-400 transition-colors hover:text-mist-200"
          @click="showHelp = !showHelp"
        >
          忘记密码
          <ChevronDown class="size-3 transition-transform" :class="showHelp ? 'rotate-180' : ''" />
        </button>

        <div v-if="showHelp" class="mt-3 border-t border-black/5 pt-3">
          <p class="text-[11px] leading-5 text-mist-400">密码在首次启动的容器日志中，或执行：</p>
          <div class="mt-2 flex items-center gap-2 rounded-lg bg-ink-950 px-2.5 py-2">
            <code class="truncate font-mono text-[10px] text-mist-300">{{ resetCommand }}</code>
            <CopyButton :value="resetCommand" label="" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
