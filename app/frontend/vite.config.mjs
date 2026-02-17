import Vue from "@vitejs/plugin-vue";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import Vuetify, { transformAssetUrls } from "vite-plugin-vuetify";
import { nodePolyfills } from "vite-plugin-node-polyfills";

import { defineConfig } from "vite";
import { fileURLToPath, URL } from "node:url";
import wails from "@wailsio/runtime/plugins/vite";

export default defineConfig({
  plugins: [
    Vue({
      template: { transformAssetUrls },
    }),
    AutoImport({
      imports: [
        "vue",
        {
          pinia: ["defineStore", "storeToRefs"],
        },
      ],
      dts: "src/auto-imports.d.ts",
      eslintrc: {
        enabled: true,
      },
      vueTemplate: true,
    }),
    Components({
      dts: "src/components.d.ts",
    }),
    Vuetify({
      autoImport: true,
      styles: {
        configFile: "src/styles/settings.scss",
      },
    }),
    nodePolyfills({
      globals: {
        Buffer: true,
        global: true,
        process: true,
      },
      protocolImports: true,
    }),
    wails("src/bindings"),
  ],
  optimizeDeps: {
    exclude: ["vuetify"],
  },
  define: { "process.env": {}, "process.version": '"v20.0.0"' },
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("src", import.meta.url)),
      process: "vite-plugin-node-polyfills/shims/process",
    },
    extensions: [".js", ".json", ".jsx", ".mjs", ".ts", ".tsx", ".vue"],
  },
  server: {
    port: 3000,
    proxy: {
      "/ws": {
        target: "ws://localhost:8097",
        ws: "true",
      },
    },
  },
});
