import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/demo-ui/',
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'DemoUI',
      formats: ['iife'],
      fileName: 'demo-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
