import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// ==============================================
// ★★★ 开关控制区 ★★★
// true  = 连接远程服务器 (120.78...)
// false = 连接本地后端 (localhost)
const isRemote = true;

// 定义地址
const remoteUrl = 'http://pp.yugams.com' // 远程地址
const localUrl  = 'http://localhost:19527';     // 本地地址 (注意确认本地Go也是这个端口吗？还是8080？)
// ==============================================

export default defineConfig({
  plugins: [vue()],
  base: './', 
  
  build: {
    outDir: 'dist', 
    assetsDir: 'assets', 
    sourcemap: false, 
  },
  
  server: {
    port: 19528, 
    open: true,
    
    proxy: {
      '/api': {
        // ★★★ 核心修改：使用三元运算符自动切换 ★★★
        target: isRemote ? remoteUrl : localUrl,
        
        changeOrigin: true, 
        // rewrite: (path) => path.replace(/^\/api/, '') 
      }
    }
  }
})
