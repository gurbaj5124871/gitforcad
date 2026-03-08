import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],

  // Clear browser cache on changes
  server: {
    hmr: {
      overlay: true,
    }
  },

  // Ensure Tauri-specific packages don't break browser preview
  build: {
    rollupOptions: {
      external: []
    }
  }
})
