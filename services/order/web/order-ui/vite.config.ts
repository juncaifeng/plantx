import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/order-ui/',
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'OrderUI',
      formats: ['iife'],
      fileName: 'order-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
