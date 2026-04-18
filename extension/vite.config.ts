import { defineConfig, type Plugin } from "vite";
import tailwindcss from "@tailwindcss/vite";
import { resolve } from "path";
import { readFileSync } from "fs";
import { Resvg } from "@resvg/resvg-js";

const browser = (process.env.BROWSER ?? "chrome") as "chrome" | "firefox" | "safari";
const ICON_SIZES = [16, 48, 128] as const;

// Reads src/manifest.json and emits it to the output directory with:
// - version synced from package.json
// - background field set for the target browser (service_worker vs scripts)
// - PNG icons generated from public/icons/icon.svg
function extensionManifest(): Plugin {
  return {
    name: "extension-manifest",
    generateBundle(_options, bundle) {
      const svgContent = readFileSync(resolve(__dirname, "public/icons/icon.svg"), "utf-8");
      for (const size of ICON_SIZES) {
        const resvg = new Resvg(svgContent, { fitTo: { mode: "width", value: size } });
        this.emitFile({
          type: "asset",
          fileName: `icons/icon-${size}.png`,
          source: resvg.render().asPng(),
        });
      }

      const pkg = JSON.parse(readFileSync(resolve(__dirname, "package.json"), "utf-8"));
      const manifest = JSON.parse(readFileSync(resolve(__dirname, "src/manifest.json"), "utf-8"));
      manifest.version = pkg.version;

      const backgroundChunk = Object.values(bundle).find((chunk) => chunk.type === "chunk" && chunk.name === "background");
      if (backgroundChunk) {
        manifest.background =
          browser === "firefox" ? { scripts: [backgroundChunk.fileName] } : { service_worker: backgroundChunk.fileName };
      }

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
    outDir: resolve(__dirname, `dist/${browser}`),
    emptyOutDir: true,
    rollupOptions: {
      input: {
        popup: resolve(__dirname, "src/index.html"),
        background: resolve(__dirname, "src/background.ts"),
      },
    },
  },
});
