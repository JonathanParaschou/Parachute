import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/storage-metadata": "http://localhost:8080",
      "/storage-roots": "http://localhost:8080",
      "/uploads": "http://localhost:8080",
      "/upload": "http://localhost:8080",
      "/heartbeat": "http://localhost:8080"
    }
  }
});
