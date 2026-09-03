import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  resolve: process.env.VITEST ? { conditions: ['browser'] } : undefined,
  plugins: [sveltekit()],
  server: {
    proxy: {
      '/api': { target: 'http://127.0.0.1:7331', changeOrigin: true },
      '/ws': { target: 'ws://127.0.0.1:7331', ws: true }
    }
  },
  test: {
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{js,ts}'],
    setupFiles: ['./vitest-setup.ts'],
    globals: true
  }
});
