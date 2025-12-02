import { vanillaExtractPlugin } from "@vanilla-extract/vite-plugin";
import react from "@vitejs/plugin-react-swc";
import { defineConfig } from "vite";
// import * as path from "node:path"
import tsconfigPaths from "vite-tsconfig-paths";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    tsconfigPaths(),
    react(),
    vanillaExtractPlugin({
      identifiers: process.env.NODE_ENV === "production" ? "short" : "debug",
    }),
  ],
  server: {
    host: "0.0.0.0",
    port: 5080,
  },
});
