import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/test-ui/',
  server: {
    port: 5175,
  },
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'TestUI',
      formats: ['iife'],
      fileName: 'test-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
