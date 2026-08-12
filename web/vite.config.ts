import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    vue(),
  ],

  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },

  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
    minify: false,
    reportCompressedSize: false,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/vue-router')) return 'router'
          if (id.includes('node_modules/pinia')) return 'pinia'
          if (id.includes('node_modules/vue')) return 'vue'
          if (id.includes('node_modules/axios')) return 'axios'
          if (id.includes('node_modules/lucide')) return 'lucide'
          if (id.includes('/modules/fleet/')) return 'fleet'
          if (id.includes('/modules/maintenance/')) return 'maintenance'
          if (id.includes('/modules/inventory/')) return 'inventory'
          if (id.includes('/modules/hr/')) return 'hr'
          if (id.includes('/modules/accounting/')) return 'accounting'
          if (id.includes('/modules/')) return 'modules'
        },
      },
      maxParallelFileOps: 1,
    },
    chunkSizeWarningLimit: 3000,
  },

  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },

  preview: {
    port: 4173,
  },
})
