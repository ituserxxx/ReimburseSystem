import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 开发时代理目标端口：与后端 SERVER_PORT 保持一致（默认 11118）
// 这样只需设置 SERVER_PORT 一个变量，前后端 API 端口即同步
const serverPort = process.env.SERVER_PORT || '11118'
const apiTarget = `http://localhost:${serverPort}`

export default defineConfig({
  // 打包后的资源以根路径引用，后端同源托管时可直接加载
  base: '/',
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true
  }
})
