import { defineConfig, type Plugin } from "vite";
import tailwindcss from "@tailwindcss/vite";
import { resolve } from "path";
import { readFileSync } from "fs";

// Reads src/manifest.json and emits it to the output directory with the
// version field set from package.json, so the extension version always
// matches the package version without manual updates.
function extensionManifest(): Plugin {
  return {
    name: "extension-manifest",
    generateBundle() {
      const pkg = JSON.parse(readFileSync(resolve(__dirname, "package.json"), "utf-8"));
      const manifest = JSON.parse(readFileSync(resolve(__dirname, "src/manifest.json"), "utf-8"));
      manifest.version = pkg.version;
      this.emitFile({
        type: "asset",
        fileName: "manifest.json",
        source: JSON.stringify(manifest, null, 2),
      });
    },
  };
}

export default defineConfig({
  esbuild: {
    jsxImportSource: "preact",
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
    },
  },
  root: resolve(__dirname, "src"),
  publicDir: resolve(__dirname, "public"),
  plugins: [tailwindcss(), extensionManifest()],
  build: {
    outDir: resolve(__dirname, "dist"),
    emptyOutDir: true,
    rollupOptions: {
      input: {
        popup: resolve(__dirname, "src/index.html"),
        background: resolve(__dirname, "src/background.ts"),
      },
    },
  },
});
