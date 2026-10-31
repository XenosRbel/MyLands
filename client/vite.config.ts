import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// `base: "./"` keeps asset paths relative so the same build works when served
// by the Go server, inside a Yandex Games ZIP (index.html at the root) and in
// a Capacitor WebView.
export default defineConfig({
  plugins: [react()],
  base: "./",
  server: {
    port: 5173,
    proxy: { "/api": "http://127.0.0.1:8080", "/packs": "http://127.0.0.1:8080" },
  },
  build: { target: "es2022", sourcemap: false, chunkSizeWarningLimit: 600 },
  test: { environment: "jsdom" },
});
