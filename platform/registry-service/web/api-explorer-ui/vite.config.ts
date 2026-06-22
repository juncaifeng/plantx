import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/api-explorer-ui/',
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'ApiExplorerUI',
      formats: ['iife'],
      fileName: 'api-explorer-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
