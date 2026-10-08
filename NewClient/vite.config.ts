import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    plugins: [vue()],
    server: {
      port: 3001, strictPort: true,
      watch: { ignored: ['**/.tmp/**', '**/.npm-cache/**', '**/test-results/**', '**/playwright-report/**'] },
      proxy: {
        '/api': { target: env.API_TARGET || 'http://127.0.0.1:8888', changeOrigin: true },
        '/ws': { target: env.WS_TARGET || 'ws://127.0.0.1:8889', ws: true, changeOrigin: true },
        // Keep public avatars, but never expose Filer bucket contents or directory listing.
        '^/filer/avatars/[0-9]+/[A-Za-z0-9_-]+\\.[A-Za-z0-9]+(?:\\?.*)?$': {
          target: env.FILER_TARGET || 'http://127.0.0.1:8890', changeOrigin: true,
          rewrite: p => p.replace(/^\/filer/, ''),
          bypass(req, res) { if (!['GET', 'HEAD'].includes(req.method || '')) { res.statusCode = 405; res.end(); return false } },
        },
      },
    },
  }
})
