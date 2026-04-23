import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

// Benchmark Dashboard build configuration with Shadow DOM support
export default defineConfig(({ mode }) => ({
  server: {
    host: "::",
    port: 8080,
  },
  plugins: [
    react(),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: false,
    rollupOptions: {
      input: {
        main: path.resolve(__dirname, 'src/main.tsx'),
      },
      output: {
        entryFileNames: 'benchmark-dashboard.js',
                  chunkFileNames: '[name]-[hash].js',
          assetFileNames: (assetInfo) => {
            // CSS files for Hugo build
            if (assetInfo.names && assetInfo.names[0] && assetInfo.names[0].endsWith('.css')) {
              return 'benchmark-dashboard.css';
            }
            return '[name][extname]';
          },
        // Ensure CSS is inlined into JS for shadow DOM
        inlineDynamicImports: false,
      },
    },
    // Extract CSS for Shadow DOM injection while also inlining
    cssCodeSplit: false,
  },
  define: {
    __BUILD_MODE__: JSON.stringify(mode),
  },
  // Minimal PostCSS config for shadow DOM build
  css: {
    postcss: {
      plugins: [],
    },
  },
  assetsInclude: ['**/*.css'],
}));
