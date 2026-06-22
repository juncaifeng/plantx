import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  base: '/apps/gateway-admin-ui/',
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.tsx'),
      name: 'GatewayAdminUI',
      formats: ['iife'],
      fileName: 'gateway-admin-ui',
    },
    outDir: 'dist',
    emptyOutDir: true,
  },
});
