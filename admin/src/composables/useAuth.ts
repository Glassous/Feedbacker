import { computed, reactive } from 'vue'

const TOKEN_KEY = 'feedbacker.token'
const USERNAME_KEY = 'feedbacker.username'
const MUST_CHANGE_KEY = 'feedbacker.mustChange'

const state = reactive({
  token: localStorage.getItem(TOKEN_KEY) ?? '',
  username: localStorage.getItem(USERNAME_KEY) ?? '',
  mustChangePassword: localStorage.getItem(MUST_CHANGE_KEY) === '1',
})

function persist() {
  if (state.token) {
    localStorage.setItem(TOKEN_KEY, state.token)
  } else {
    localStorage.removeItem(TOKEN_KEY)
  }
  localStorage.setItem(USERNAME_KEY, state.username)
  localStorage.setItem(MUST_CHANGE_KEY, state.mustChangePassword ? '1' : '0')
}

export interface SessionPayload {
  token: string
  username: string
  mustChangePassword: boolean
}

export function useAuth() {
  const isAuthenticated = computed(() => state.token.length > 0)

  function setSession(payload: SessionPayload) {
    state.token = payload.token
    state.username = payload.username
    state.mustChangePassword = payload.mustChangePassword
    persist()
  }

  function updateProfile(payload: { username: string; mustChangePassword: boolean }) {
    state.username = payload.username
    state.mustChangePassword = payload.mustChangePassword
    persist()
  }

  function markMustChange() {
    state.mustChangePassword = true
    persist()
  }

  function markPasswordChanged() {
    state.mustChangePassword = false
    persist()
  }

  function clearSession() {
    state.token = ''
    state.username = ''
    state.mustChangePassword = false
    persist()
  }

  return {
    state,
    isAuthenticated,
    setSession,
    updateProfile,
    markMustChange,
    markPasswordChanged,
    clearSession,
  }
}
