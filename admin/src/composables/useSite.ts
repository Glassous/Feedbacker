import { computed, ref } from 'vue'

const siteName = ref('Feedbacker')
const version = ref('')
const versionLabel = computed(() => {
  const raw = version.value || 'dev'
  return /^\d/.test(raw) ? `v${raw}` : raw
})
let loaded = false
let pending = false

export function useSite() {
  function load() {
    if (loaded || pending) return
    pending = true
    fetch('/api/meta')
      .then((response) => (response.ok ? response.json() : Promise.reject(new Error('meta unavailable'))))
      .then((data: { siteName?: string; version?: string }) => {
        if (data.siteName) {
          siteName.value = data.siteName
          document.title = `${data.siteName} · 反馈管理台`
        }
        if (data.version) {
          version.value = data.version
        }
        loaded = true
      })
      .catch((error) => {
        console.error('读取站点信息失败', error)
      })
      .finally(() => {
        pending = false
      })
  }

  return { siteName, version, versionLabel, load }
}
