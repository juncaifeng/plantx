import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/audit-admin-ui/',
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'AuditAdminUI',
      formats: ['iife'],
      fileName: 'audit-admin-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
