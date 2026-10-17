<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ChevronDown,
  KeyRound,
  LogOut,
  Menu,
  MessageSquare,
  PanelLeftClose,
  PanelLeftOpen,
  Settings,
  X,
} from 'lucide-vue-next'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import { errorMessage, fetchMe, logout } from '../lib/api'
import { useAuth } from '../composables/useAuth'
import { useSite } from '../composables/useSite'
import { pushToast } from '../composables/useToast'

const route = useRoute()
const router = useRouter()
const { state, updateProfile, clearSession } = useAuth()
const { siteName, versionLabel } = useSite()

const navItems = [
  { name: 'feedback', label: '反馈', icon: MessageSquare },
  { name: 'settings', label: '设置', icon: Settings },
]

const collapsed = ref(localStorage.getItem('feedbacker.sidebarCollapsed') === '1')
const mobileOpen = ref(false)
const accountOpen = ref(false)
const passwordOpen = ref(false)
const accountRef = ref<HTMLElement | null>(null)

const pageTitle = computed(() => (route.meta.title as string | undefined) ?? '反馈')
const avatarLetter = computed(() => (state.username || 'A').slice(0, 1).toUpperCase())

watch(collapsed, (value) => {
  localStorage.setItem('feedbacker.sidebarCollapsed', value ? '1' : '0')
})

watch(mobileOpen, (open) => {
  document.body.style.overflow = open ? 'hidden' : ''
})

function onDocumentClick(event: MouseEvent) {
  if (!accountOpen.value) return
  const target = event.target as Node
  if (accountRef.value && !accountRef.value.contains(target)) {
    accountOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  fetchMe()
    .then((info) => updateProfile(info))
    .catch((error) => {
      console.error(error)
      if (error?.response?.status !== 401) {
        pushToast({ title: '读取账号信息失败', description: errorMessage(error), variant: 'error' })
      }
    })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  document.body.style.overflow = ''
})

function openPassword() {
  accountOpen.value = false
  passwordOpen.value = true
}

function signOut() {
  accountOpen.value = false
  logout()
    .catch((error) => console.error(error))
    .finally(() => {
      clearSession()
      router.replace({ name: 'login' })
    })
}
</script>

<template>
  <div class="min-h-screen bg-ink-950">
    <aside
      class="fixed inset-y-0 left-0 z-30 hidden flex-col border-r border-black/6 bg-ink-900 transition-[width] duration-300 lg:flex"
      :class="collapsed ? 'w-[72px]' : 'w-[224px]'"
    >
      <div class="flex h-14 items-center gap-2.5 px-4">
        <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-mist-100 text-ink-950">
          <MessageSquare class="size-4" />
        </span>
        <span v-if="!collapsed" class="truncate text-[13px] font-medium text-mist-100">{{ siteName }}</span>
      </div>

      <nav class="mt-3 flex-1 space-y-0.5 px-2.5">
        <RouterLink
          v-for="item in navItems"
          :key="item.name"
          v-slot="{ href, navigate, isActive }"
          :to="{ name: item.name }"
          custom
        >
          <a
            :href="href"
            class="group flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] transition-colors duration-150"
            :class="
              isActive
                ? 'bg-black/[0.06] text-mist-100'
                : 'text-mist-400 hover:bg-black/[0.03] hover:text-mist-200'
            "
            @click="navigate"
          >
            <component
              :is="item.icon"
              class="size-4 shrink-0 transition-colors"
              :class="isActive ? 'text-mist-300' : 'text-mist-500 group-hover:text-mist-400'"
            />
            <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
          </a>
        </RouterLink>
      </nav>

      <div
        class="flex items-center gap-2 border-t border-black/6 p-2.5"
        :class="collapsed ? 'justify-center' : 'justify-between'"
      >
        <span v-if="!collapsed" class="px-1 font-mono text-[10px] text-mist-500">{{ versionLabel }}</span>
        <button
          type="button"
          class="flex size-8 cursor-pointer items-center justify-center rounded-lg text-mist-500 transition-colors hover:bg-black/[0.04] hover:text-mist-200"
          :aria-label="collapsed ? '展开侧栏' : '收起侧栏'"
          @click="collapsed = !collapsed"
        >
          <PanelLeftOpen v-if="collapsed" class="size-4" />
          <PanelLeftClose v-else class="size-4" />
        </button>
      </div>
    </aside>

    <div
      class="min-h-screen transition-[padding] duration-300"
      :class="collapsed ? 'lg:pl-[72px]' : 'lg:pl-[224px]'"
    >
      <header
        class="sticky top-0 z-20 flex h-14 items-center justify-between gap-4 border-b border-black/6 bg-ink-900/85 px-4 backdrop-blur sm:px-6 lg:px-8"
      >
        <div class="flex min-w-0 items-center gap-3">
          <button
            type="button"
            class="flex size-8 cursor-pointer items-center justify-center rounded-lg border border-black/8 text-mist-400 transition-colors hover:text-mist-100 lg:hidden"
            aria-label="打开导航"
            @click="mobileOpen = true"
          >
            <Menu class="size-4" />
          </button>
          <h1 class="truncate text-[13px] font-medium text-mist-200">{{ pageTitle }}</h1>
        </div>

        <div ref="accountRef" class="relative">
          <button
            type="button"
            class="flex cursor-pointer items-center gap-2 rounded-lg border border-black/8 bg-black/[0.03] py-1 pl-1 pr-2 transition-colors hover:bg-black/[0.06]"
            @click="accountOpen = !accountOpen"
          >
            <span class="flex size-6 items-center justify-center rounded-md bg-mist-100 text-[10px] font-semibold text-ink-950">
              {{ avatarLetter }}
            </span>
            <span class="hidden max-w-[100px] truncate text-xs text-mist-300 sm:block">
              {{ state.username || 'admin' }}
            </span>
            <ChevronDown class="size-3.5 text-mist-500 transition-transform" :class="accountOpen ? 'rotate-180' : ''" />
          </button>

          <Transition
            enter-active-class="transition duration-150 ease-out"
            enter-from-class="-translate-y-1 opacity-0"
            enter-to-class="translate-y-0 opacity-100"
            leave-active-class="transition duration-100 ease-in"
            leave-from-class="opacity-100"
            leave-to-class="-translate-y-1 opacity-0"
          >
            <div
              v-if="accountOpen"
              class="absolute right-0 top-full mt-1.5 w-40 rounded-lg border border-black/10 bg-ink-800 p-1 shadow-xl"
            >
              <button
                type="button"
                class="flex w-full cursor-pointer items-center gap-2 rounded-md px-2.5 py-2 text-left text-xs text-mist-300 transition-colors hover:bg-black/[0.05] hover:text-mist-100"
                @click="openPassword"
              >
                <KeyRound class="size-3.5" />
                修改密码
              </button>
              <button
                type="button"
                class="flex w-full cursor-pointer items-center gap-2 rounded-md px-2.5 py-2 text-left text-xs text-mist-400 transition-colors hover:bg-signal-red/10 hover:text-signal-red"
                @click="signOut"
              >
                <LogOut class="size-3.5" />
                退出登录
              </button>
            </div>
          </Transition>
        </div>
      </header>

      <main class="mx-auto w-full max-w-[1400px] px-4 py-5 sm:px-6 lg:px-8 lg:py-6">
        <RouterView />
      </main>
    </div>

    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="mobileOpen" class="fixed inset-0 z-40 lg:hidden">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="mobileOpen = false" />
        <div class="absolute inset-y-0 left-0 flex w-[260px] flex-col border-r border-black/8 bg-ink-900">
          <div class="flex h-14 items-center justify-between px-4">
            <div class="flex items-center gap-2.5">
              <span class="flex size-8 items-center justify-center rounded-lg bg-mist-100 text-ink-950">
                <MessageSquare class="size-4" />
              </span>
              <span class="text-[13px] font-medium text-mist-100">{{ siteName }}</span>
            </div>
            <button
              type="button"
              class="flex size-8 cursor-pointer items-center justify-center rounded-lg text-mist-500 hover:text-mist-200"
              aria-label="关闭导航"
              @click="mobileOpen = false"
            >
              <X class="size-4" />
            </button>
          </div>

          <nav class="mt-3 flex-1 space-y-0.5 px-2.5">
            <RouterLink
              v-for="item in navItems"
              :key="item.name"
              v-slot="{ href, navigate, isActive }"
              :to="{ name: item.name }"
              custom
            >
              <a
                :href="href"
                class="flex items-center gap-2.5 rounded-lg px-2.5 py-2.5 text-[13px]"
                :class="isActive ? 'bg-black/[0.06] text-mist-100' : 'text-mist-400'"
                @click="
                  navigate();
                  mobileOpen = false
                "
              >
                <component :is="item.icon" class="size-4" />
                {{ item.label }}
              </a>
            </RouterLink>
          </nav>

          <div class="border-t border-black/6 p-4">
            <p class="text-xs text-mist-300">{{ state.username || 'admin' }}</p>
            <p class="mt-1 font-mono text-[10px] text-mist-500">{{ versionLabel }}</p>
          </div>
        </div>
      </div>
    </Transition>

    <ChangePasswordDialog
      :open="state.mustChangePassword || passwordOpen"
      :forced="state.mustChangePassword"
      @close="passwordOpen = false"
      @success="passwordOpen = false"
    />
  </div>
</template>
