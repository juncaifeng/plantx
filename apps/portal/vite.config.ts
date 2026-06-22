import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  base: "/",
  server: {
    port: 5174,
    proxy: {
      "/api": { target: "http://localhost:80", changeOrigin: true },
      "/oauth/token": { target: "http://localhost:80", changeOrigin: true },
      "/apps/order-ui/": {
        target: "http://localhost:5173",
        changeOrigin: true,
        rewrite: (path) => {
          if (path.endsWith("order-ui.js")) {
            return path.replace(/order-ui\.js$/, "order-ui.iife.js");
          }
          return path;
        },
      },
      "/apps/test-ui/": {
        target: "http://localhost:5175",
        changeOrigin: true,
        rewrite: (path) =>
          path.endsWith("test-ui.js")
            ? path.replace(/test-ui\.js$/, "test-ui.iife.js")
            : path,
      },
      "/apps/gateway-admin-ui/": {
        target: "http://localhost:80",
        changeOrigin: true,
      },
      "/apps/tenant-admin-ui/": {
        target: "http://localhost:80",
        changeOrigin: true,
      },
      "/apps/iam-admin-ui/": {
        target: "http://localhost:80",
        changeOrigin: true,
      },
      "/apps/audit-admin-ui/": {
        target: "http://localhost:80",
        changeOrigin: true,
      },
    },
  },
  build: {
    target: "esnext",
    outDir: "dist",
  },
});
