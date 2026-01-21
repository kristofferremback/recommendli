import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/recommendations': {
        target: 'http://127.0.0.1:9999',
        changeOrigin: true,
        cookieDomainRewrite: '127.0.0.1',
        secure: false,
        ws: true,
      },
    },
  },
  build: {
    outDir: '../static/dist',
    emptyOutDir: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
