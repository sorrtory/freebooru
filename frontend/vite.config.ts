import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const environment = loadEnv(mode, '.', '')

  return {
    plugins: [vue()],
    build: {
      emptyOutDir: true,
      outDir: fileURLToPath(new URL('../internal/webui/dist', import.meta.url)),
    },
    server: {
      proxy: {
        '/api': environment.FREEBOORU_API_TARGET ?? 'http://127.0.0.1:52800',
      },
    },
    test: {
      environment: 'jsdom',
    },
  }
})
