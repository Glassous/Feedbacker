/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** 接口根地址，例如 https://feedback.glassous.top/api；留空则使用同源 /api */
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>
  export default component
}
