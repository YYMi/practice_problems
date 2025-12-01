import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  // ★★★ 关键配置 ★★★
  // 如果部署在根目录，填 './' 或者 '/'
  // 如果部署在子目录（例如 /admin/），填 '/admin/'
  base: './', 
  
  build: {
    outDir: 'dist', // 打包输出的文件夹名称，默认叫 dist
    assetsDir: 'assets', // 静态资源放哪里
    sourcemap: false, // 是否生成 map 文件（生产环境建议关掉，为了安全和体积）
  },
   server: {
    // 代理配置
    proxy: {
      '/api': {
        target: 'http://localhost:8080', // 你的后端地址
        changeOrigin: true, // 允许跨域
        // rewrite: (path) => path.replace(/^\/api/, '') // 如果后端接口没有 /api 前缀，才需要这行
      }
    }
  }
})
