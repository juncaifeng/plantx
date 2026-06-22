import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/registry-admin-ui/',
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'RegistryAdminUI',
      formats: ['iife'],
      fileName: 'registry-admin-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
