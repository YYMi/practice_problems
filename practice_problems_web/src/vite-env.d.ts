/// <reference types="vite/client" />

// 扩展 ImportMetaEnv 接口，定义环境变量类型
interface ImportMetaEnv {
  readonly VITE_SHOW_DEV_MODE?: string
  // 你可以在这里添加其他环境变量的类型定义
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

// 这一段是告诉 TS 如何理解 .vue 文件
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// CSS 文件类型声明
declare module '*.css' {
  const content: string
  export default content
}

// wangeditor CSS 类型声明
declare module '@wangeditor/editor/dist/css/style.css';
