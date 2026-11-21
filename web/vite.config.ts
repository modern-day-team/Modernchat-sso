import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import UnoCSS from 'unocss/vite'

// SSO 门户独立部署（base 为根路径，与主仓库 chat 应用部署在 /app/ 下不同）。
// dev / preview 统一把 /api 代理到本地 SSO 后端 :8090（Go 服务端口，主仓库是 :8080）。
export default defineConfig(({ mode }) => ({
  base: '/',
  // devtools 仅开发启用：生产构建不应包含它（体积与运行时开销）
  plugins: [vue(), UnoCSS(), ...(mode === 'development' ? [vueDevTools()] : [])],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    // 家族端口约定：主仓库 web :5173、官网 :5174，SSO 门户用 :5175，便于三站并行开发
    port: 5175,
    proxy: {
      '/api': 'http://localhost:8090',
    },
  },
  // preview（生产构建预览）默认不带代理，导致本地跑 preview 时接口全 404；与 dev 保持一致
  preview: {
    proxy: {
      '/api': 'http://localhost:8090',
    },
  },
}))
