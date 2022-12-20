import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
// import * as path from "node:path"
import tsconfigPaths from "vite-tsconfig-paths";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [tsconfigPaths(), react()],
  server: {
    host: "0.0.0.0",
    port: 5080,
  },
});
