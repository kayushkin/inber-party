import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // Separate vendor chunk for node_modules dependencies
          if (id.includes('node_modules')) {
            if (id.includes('react') || id.includes('react-dom')) {
              return 'react-vendor'
            }
            if (id.includes('zustand')) {
              return 'state-vendor'
            }
            return 'vendor'
          }
          
          // Separate chunk for large UI components
          if (id.includes('/components/CharacterSheet.tsx') || 
              id.includes('/components/BountyBoard.tsx') || 
              id.includes('/components/PerformanceDashboard.tsx')) {
            return 'ui-heavy'
          }
          
          // Separate chunk for utilities and constants
          if (id.includes('/utils/') || id.includes('/constants/')) {
            return 'utils'
          }
        }
      }
    },
    // Enable build optimizations
    target: 'esnext',
    minify: 'esbuild',
    // Optimize chunk size limits
    chunkSizeWarningLimit: 600
  }
})
