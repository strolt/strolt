import babel from "@rolldown/plugin-babel";
import { vanillaExtractPlugin } from "@vanilla-extract/vite-plugin";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    tsconfigPaths(),
    react(),
    babel({ presets: [reactCompilerPreset()] }),
    vanillaExtractPlugin({
      identifiers: process.env.NODE_ENV === "production" ? "short" : "debug",
    }),
  ],
  server: {
    host: "0.0.0.0",
    port: 5080,
  },
});
