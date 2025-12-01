/// <reference types="vite/client" />

// 这一段是告诉 TS 如何理解 .vue 文件
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}
