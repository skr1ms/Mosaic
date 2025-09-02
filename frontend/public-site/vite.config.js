import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': resolve(process.cwd(), './src'),
    },
  },
  server: {
    port: 5173,
    host: '0.0.0.0'
  },
  build: {
    outDir: 'build',
    sourcemap: false,
    minify: 'esbuild',
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // React core
          if (id.includes('react') || id.includes('react-dom')) {
            return 'react-vendor';
          }
          // Router
          if (id.includes('react-router')) {
            return 'router';
          }
          // UI Libraries
          if (id.includes('framer-motion')) {
            return 'animations';
          }
          if (id.includes('lucide-react')) {
            return 'icons';
          }
          // Internationalization
          if (id.includes('i18next') || id.includes('react-i18next')) {
            return 'i18n';
          }
          // HTTP client
          if (id.includes('axios') || id.includes('@tanstack/react-query')) {
            return 'api';
          }
          // State management
          if (id.includes('zustand')) {
            return 'store';
          }
          // Pages - split large pages into separate chunks
          if (id.includes('/pages/')) {
            const page = id.split('/pages/')[1]?.split('.')[0];
            if (page && ['DiamondMosaicPage', 'ShopPage', 'MosaicPreviewPage'].includes(page)) {
              return `page-${page.toLowerCase()}`;
            }
            return 'pages';
          }
          // Components
          if (id.includes('/components/')) {
            return 'components';
          }
          // Other vendor libraries
          if (id.includes('node_modules')) {
            return 'vendor';
          }
        },
      },
    },
  },
})
