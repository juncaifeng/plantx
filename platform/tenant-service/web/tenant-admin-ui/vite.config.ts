import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/tenant-admin-ui/',
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'TenantAdminUI',
      formats: ['iife'],
      fileName: 'tenant-admin-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
