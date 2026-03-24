import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'ContentExtractor',
      formats: ['es', 'cjs'],
      fileName: 'index',
    },
    rollupOptions: {
      external: [/^pdfjs-dist/, 'tesseract.js'],
    },
    target: 'es2020',
    sourcemap: true,
  },
});
