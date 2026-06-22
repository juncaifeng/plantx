import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/iam-admin-ui/',
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'IAMAdminUI',
      formats: ['iife'],
      fileName: 'iam-admin-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
